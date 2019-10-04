package insaneJSON

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

// ===LATEST BENCH RESULTS===
// BenchmarkFair/complex-stable-flavor|complex-4         	    1000	   2014381 ns/op	 638.58 MB/s	    2107 B/op	      10 allocs/op
// BenchmarkFair/complex-chaotic-flavor|complex-4        	    3000	    516549 ns/op	 160.88 MB/s	   36291 B/op	     756 allocs/op
// BenchmarkFair/get-stable-flavor|get-4                 	 2000000	       968 ns/op	664405.82 MB/s	       0 B/op	       0 allocs/op
// BenchmarkFair/get-chaotic-flavor|get-4                	  200000	     10884 ns/op	3817.32 MB/s	    4032 B/op	      84 allocs/op
// BenchmarkValueDecodeInt-4                             	 2000000	       920 ns/op	     288 B/op	       6 allocs/op
// BenchmarkValueEscapeString-4                          	 1000000	      1821 ns/op	       0 B/op	       0 allocs/op

const (
	Object Type = 0
	Array  Type = 1
	String Type = 2
	Number Type = 3
	True   Type = 4
	False  Type = 5
	Null   Type = 6
	Field  Type = 7

	// internal
	objectEnd     Type = 8
	arrayEnd      Type = 9
	escapedString Type = 10
	escapedField  Type = 11

	hex = "0123456789abcdef"
)

type Type int

var (
	StartNodePoolSize = 16

	decoderPool      = make([]*decoder, 0, 16)
	decoderPoolIndex = -1
	decoderPoolMu    = &sync.Mutex{}

	numbersMap = make([]byte, 256)

	// decode errors
	ErrEmptyJSON                    = errors.New("json is empty")
	ErrUnexpectedJSONEnding         = errors.New("unexpected ending of json")
	ErrUnexpectedEndOfString        = errors.New("unexpected end of string")
	ErrUnexpectedEndOfTrue          = errors.New("unexpected end of true")
	ErrUnexpectedEndOfFalse         = errors.New("unexpected end of false")
	ErrUnexpectedEndOfNull          = errors.New("unexpected end of null")
	ErrUnexpectedEndOfObjectField   = errors.New("unexpected end of object field")
	ErrExpectedObjectField          = errors.New("expected object field")
	ErrExpectedObjectFieldSeparator = errors.New("expected object field separator")
	ErrExpectedValue                = errors.New("expected value")
	ErrExpectedComma                = errors.New("expected comma")

	// api errors
	ErrRootIsNil = errors.New("root is nil")
	ErrNotFound  = errors.New("node isn't found")
	ErrNotObject = errors.New("node isn't an object")
	ErrNotArray  = errors.New("node isn't an array")
	ErrNotBool   = errors.New("node isn't a bool")
	ErrNotString = errors.New("node isn't a string")
	ErrNotNumber = errors.New("node isn't a number")
	ErrNotField  = errors.New("node isn't an object field")
)

const (
	FlagFieldMap = 1 << 0
)

func init() {
	numbersMap['.'] = 1
	numbersMap['-'] = 1
	numbersMap['e'] = 1
	numbersMap['E'] = 1
	numbersMap['+'] = 1
}

type Root struct {
	*Node
}

type Last struct {
	*Node
}

type Node struct {
	Type   Type
	next   *Node
	parent *Node
	value  string
	data   *data
}

type StrictNode struct {
	*Node
}

type data struct {
	values   []*Node
	end      *Node
	index    int
	flags    int
	dirtySeq int
	fields   *map[string]int
	err      *StrictNode
	decoder  *decoder
}

type decoder struct {
	id int

	json []byte

	root     Root
	nodePool []*Node
	nodes    int
}

// ReleaseMem sends node pool and internal buffer to GC
// useful to reduce memory usage after decoding big JSON
func (r *Root) ReleaseMem() {
	r.ReleasePoolMem()
	r.ReleaseBufMem()
}

// ReleasePoolMem sends node pool to GC
func (r *Root) ReleasePoolMem() {
	r.data.decoder.initPool()
}

// ReleasePoolMem sends internal buffer to GC
func (r *Root) ReleaseBufMem() {
	r.data.decoder.json = make([]byte, 0, 0)
}

// BuffCap returns current size of internal buffer
func (r *Root) BuffCap() int {
	return cap(r.data.decoder.json)
}

// PullSize returns current size of node pool
func (r *Root) PullSize() int {
	return len(r.data.decoder.nodePool)
}

// ******************** //
//      MAIN SHIT       //
// ******************** //

// legendary insane decode function
func (d *decoder) decode(json string, shouldReset bool) (*Node, error) {
	if shouldReset {
		d.nodes = 0
		d.json = d.json[:0]
	}
	o := len(d.json)

	d.json = append(d.json, json...)
	json = toString(d.json)

	l := len(json)
	if l == 0 {
		return nil, insaneErr(ErrEmptyJSON, json, o)
	}

	nodePool := d.nodePool
	nodePoolLen := len(nodePool)
	nodes := d.nodes

	root := nodePool[nodes]
	root.parent = nil
	curNode := nodePool[nodes]
	topNode := root.parent

	c := byte('i') // i means insane
	t := 0
	x := 0
	goto decode
decodeObject:
	if o == l {
		return nil, insaneErr(ErrUnexpectedJSONEnding, json, o)
	}

	// skip wc
	c = json[o]
	o++
	if c <= 0x20 {
		for o != l {
			c = json[o]
			o++
			if c == 0x20 || c == 0x0A || c == 0x09 || c == 0x0D {
				continue
			}
			break
		}
	}

	if c == '}' {
		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = objectEnd
		curNode.parent = topNode

		topNode.data.end = curNode
		topNode = topNode.parent

		goto pop
	}

	if c != ',' {
		if len(topNode.data.values) > 0 {
			return nil, insaneErr(ErrExpectedComma, json, o)
		}
		o--
	} else {
		if len(topNode.data.values) == 0 {
			return nil, insaneErr(ErrExpectedObjectField, json, o)
		}
		if o == l {
			return nil, insaneErr(ErrUnexpectedJSONEnding, json, o)
		}
	}

	// skip wc
	c = json[o]
	o++
	if c <= 0x20 {
		for o != l {
			c = json[o]
			o++
			if c == 0x20 || c == 0x0A || c == 0x09 || c == 0x0D {
				continue
			}
			break
		}
	}

	if c != '"' {
		return nil, insaneErr(ErrExpectedObjectField, json, o)
	}

	t = o - 1
	for {
		x = strings.IndexByte(json[o:], '"')
		o += x + 1
		if x < 0 {
			return nil, insaneErr(ErrUnexpectedEndOfObjectField, json, o)
		}

		if x == 0 || json[o-2] != '\\' || (json[o-2] == '\\' && json[o-3] == '\\') {
			break
		}
	}
	if o == l {
		return nil, insaneErr(ErrExpectedObjectFieldSeparator, json, o)
	}

	curNode.next = nodePool[nodes]
	curNode = curNode.next
	nodes++

	// skip wc
	c = json[o]
	o++
	if c <= 0x20 {
		for o != l {
			c = json[o]
			o++
			if c == 0x20 || c == 0x0A || c == 0x09 || c == 0x0D {
				continue
			}
			break
		}
	}

	if c != ':' {
		return nil, insaneErr(ErrExpectedObjectFieldSeparator, json, o)
	}
	if o == l {
		return nil, insaneErr(ErrExpectedValue, json, o)
	}
	curNode.Type = escapedField
	curNode.value = json[t:o]
	curNode.parent = topNode
	topNode.data.values = append(topNode.data.values, curNode)

	goto decode
decodeArray:
	if o == l {
		return nil, insaneErr(ErrUnexpectedJSONEnding, json, o)
	}
	// skip wc
	c = json[o]
	o++
	if c <= 0x20 {
		for o != l {
			c = json[o]
			o++
			if c == 0x20 || c == 0x0A || c == 0x09 || c == 0x0D {
				continue
			}
			break
		}
	}

	if c == ']' {
		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = arrayEnd
		curNode.parent = topNode

		topNode.data.end = curNode
		topNode = topNode.parent

		goto pop
	}

	if c != ',' {
		if len(topNode.data.values) > 0 {
			return nil, insaneErr(ErrExpectedComma, json, o)
		}
		o--
	} else {
		if len(topNode.data.values) == 0 {
			return nil, insaneErr(ErrExpectedValue, json, o)
		}
		if o == l {
			return nil, insaneErr(ErrUnexpectedJSONEnding, json, o)
		}
	}

	topNode.data.values = append(topNode.data.values, nodePool[nodes])
decode:
	if nodes > nodePoolLen-16 {
		nodePool = d.expandPool()
		nodePoolLen = len(nodePool)
	}
	// skip wc
	c = json[o]
	o++
	if c <= 0x20 {
		for o != l {
			c = json[o]
			o++
			if c == 0x20 || c == 0x0A || c == 0x09 || c == 0x0D {
				continue
			}
			break
		}
	}
	switch c {
	case '{':
		if o == l {
			return nil, insaneErr(ErrExpectedObjectField, json, o)
		}

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = Object
		curNode.data.values = curNode.data.values[:0]
		curNode.data.flags = 0
		curNode.data.dirtySeq = -1
		curNode.parent = topNode

		topNode = curNode
		goto decodeObject
	case '[':
		if o == l {
			return nil, insaneErr(ErrExpectedValue, json, o)
		}
		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = Array
		curNode.data.values = curNode.data.values[:0]
		curNode.data.dirtySeq = -1
		curNode.parent = topNode

		topNode = curNode
		goto decodeArray
	case '"':
		t = o
		for {
			x := strings.IndexByte(json[t:], '"')
			t += x + 1
			if x < 0 {
				return nil, insaneErr(ErrUnexpectedEndOfString, json, o)
			}
			if x == 0 || json[t-2] != '\\' || (json[t-2] == '\\' && json[t-3] == '\\') {
				break
			}
		}

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = escapedString
		curNode.value = json[o-1 : t]
		curNode.data.flags = 0
		curNode.data.dirtySeq = -1
		curNode.parent = topNode

		o = t
	case 't':
		if len(json) < o+3 || json[o:o+3] != "rue" {
			return nil, insaneErr(ErrUnexpectedEndOfTrue, json, o)
		}
		o += 3

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = True
		curNode.data.dirtySeq = -1
		curNode.parent = topNode

	case 'f':
		if len(json) < o+4 || json[o:o+4] != "alse" {
			return nil, insaneErr(ErrUnexpectedEndOfFalse, json, o)
		}
		o += 4

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = False
		curNode.data.dirtySeq = -1
		curNode.parent = topNode

	case 'n':
		if len(json) < o+3 || json[o:o+3] != "ull" {
			return nil, insaneErr(ErrUnexpectedEndOfNull, json, o)
		}
		o += 3

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = Null
		curNode.data.dirtySeq = -1
		curNode.parent = topNode
	default:
		o--
		t = o
		for ; o != l && ((json[o] >= '0' && json[o] <= '9') || numbersMap[json[o]] == 1); o++ {
		}
		if t == o {
			return nil, insaneErr(ErrExpectedValue, json, o)
		}

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.Type = Number
		curNode.value = json[t:o]
		curNode.data.dirtySeq = -1
		curNode.parent = topNode
	}
pop:
	if topNode == nil {
		goto exit
	}

	if nodes > nodePoolLen-16 {
		nodePool = d.expandPool()
		nodePoolLen = len(nodePool)
	}

	if topNode.Type == Object {
		goto decodeObject
	} else {
		goto decodeArray
	}
exit:
	if o != l {
		// skip wc
		c = json[o]
		if c <= 0x20 {
			for o != l {
				c = json[o]
				o++
				if c != 0x20 && c != 0x0A && c != 0x09 && c != 0x0D {
					break
				}
			}
		}

		if o != l {
			return nil, insaneErr(ErrUnexpectedJSONEnding, json, o)
		}
	}

	curNode.next = nil
	d.nodes = nodes

	return root, nil
}

func (d *decoder) decodeHeadless(json string, isPooled bool) (*Root, error) {
	root, err := d.decode(json, true)
	if err != nil {
		if isPooled {
			backToPool(d)
		}
		return nil, err
	}

	d.root.Node = root
	d.root.data.decoder = d

	return &d.root, nil
}

// EncodeToByte legendary insane encode function
// slow because it allocates new byte buffer on every call
// use Encode to reuse already created buffer and gain more performance
func (n *Node) EncodeToByte() []byte {
	return n.Encode([]byte{})
}

// EncodeToString legendary insane encode function
// slow because it allocates new string on every call
// use Encode to reuse already created buffer and gain more performance
func (n *Node) EncodeToString() string {
	return toString(n.Encode([]byte{}))
}

// Encode legendary insane encode function
// uses already created byte buffer to place json data so
// mem allocations may occur only if buffer isn't long enough
// use it for performance
func (n *Node) Encode(out []byte) []byte {
	out = out[:0]
	s := 0
	curNode := n
	topNode := n

	if curNode.next != nil {
		if curNode.next.Type == objectEnd {
			return append(out, "{}"...)
		}
		if curNode.next.Type == arrayEnd {
			return append(out, "[]"...)
		}
	}

	goto encodeSkip
encode:
	out = append(out, ","...)
encodeSkip:
	switch curNode.Type {
	case Object:
		if curNode.next.Type == objectEnd {
			out = append(out, "{}"...)
			curNode = curNode.next.next
			goto popSkip
		}
		topNode = curNode
		out = append(out, '{')
		curNode = curNode.next
		if curNode.Type == Field {
			out = escapeString(out, curNode.value)
			out = append(out, ':')
		} else {
			out = append(out, curNode.value...)
		}
		curNode = curNode.next
		s++
		goto encodeSkip
	case Array:
		if curNode.next.Type == arrayEnd {
			out = append(out, "[]"...)
			curNode = curNode.next.next
			goto popSkip
		}
		topNode = curNode
		out = append(out, '[')
		curNode = curNode.next
		s++
		goto encodeSkip
	case Number:
		out = append(out, curNode.value...)
	case String:
		out = escapeString(out, curNode.value)
	case escapedString:
		out = append(out, curNode.value...)
	case False:
		out = append(out, "false"...)
	case True:
		out = append(out, "true"...)
	case Null:
		out = append(out, "null"...)
	}
pop:
	curNode = curNode.next
popSkip:
	if topNode.Type == Array {
		if curNode.Type == arrayEnd {
			out = append(out, "]"...)
			topNode = topNode.parent
			s--
			if s == 0 {
				return out
			}
			goto pop
		}
		goto encode
	} else if topNode.Type == Object {
		if curNode.Type == objectEnd {
			out = append(out, "}"...)
			topNode = topNode.parent
			s--
			if s == 0 {
				return out
			}
			goto pop
		}
		out = append(out, ","...)
		if curNode.Type == Field {
			out = escapeString(out, curNode.value)
			out = append(out, ':')
		} else {
			out = append(out, curNode.value...)
		}
		curNode = curNode.next
		goto encodeSkip
	} else {
		return out
	}
}

// Dig legendary insane dig function
func (n *Node) Dig(path ...string) *Node {
	if n == nil {
		return nil
	}
	if len(path) == 0 {
		return n
	}
	node := n
	pathField := path[0]

	pathDepth := len(path)
	depth := 0
get:
	if node.Type == Array {
		goto getArray
	}

	if node.data.flags&FlagFieldMap != FlagFieldMap && len(node.data.values) > 18 {
		if node.data.fields == nil {
			fields := make(map[string]int, len(node.data.values))
			node.data.fields = &fields
		} else {
			for field := range *node.data.fields {
				delete(*node.data.fields, field)
			}
		}

		for index, field := range node.data.values {
			if field.Type == escapedField {
				field.unescapeField()
			}
			(*node.data.fields)[field.value] = index
		}
		node.data.flags |= FlagFieldMap
	}

	if node.data.flags&FlagFieldMap == FlagFieldMap {
		index, has := (*node.data.fields)[pathField]
		if !has {
			return nil
		}

		depth++
		if depth == pathDepth {
			result := node.data.values[index].next
			result.data.dirtySeq = node.data.dirtySeq
			result.data.index = index

			return result
		}

		pathField = path[depth]
		node = node.data.values[index].next
		goto get
	}

	for index, field := range node.data.values {
		if field.Type == escapedField {
			field.unescapeField()
		}

		if field.value == pathField {
			depth++
			if depth == pathDepth {
				result := field.next
				result.data.dirtySeq = node.data.dirtySeq
				result.data.index = index

				return result
			}
			pathField = path[depth]
			node = field.next
			goto get
		}
	}
	return nil
getArray:
	index, err := strconv.Atoi(path[depth])
	if err != nil || index < 0 || index >= len(node.data.values) {
		return nil
	}
	depth++
	if depth == pathDepth {
		result := node.data.values[index]
		result.data.dirtySeq = node.data.dirtySeq
		result.data.index = index

		return result
	}
	pathField = path[depth]
	node = node.data.values[index]
	goto get
}

func (d *decoder) getNode() *Node {
	node := d.nodePool[d.nodes]
	d.nodes++
	if d.nodes > len(d.nodePool)-16 {
		d.expandPool()
	}

	return node
}

func (n *Node) DigStrict(path ...string) (*StrictNode, error) {
	result := n.Dig(path...)
	if result == nil {
		return nil, ErrNotFound
	}

	return result.InStrictMode(), nil
}

func (n *Node) AddField(name string) *Node {
	if n == nil || n.Type != Object {
		return nil
	}

	value := n.Dig(name)
	if value != nil {
		return value
	}

	decoder := n.data.decoder

	newNull := decoder.getNode()
	newNull.Type = Null
	newNull.next = n.data.end
	newNull.data.decoder = decoder
	newNull.parent = n

	newField := decoder.getNode()
	newField.Type = Field
	newField.next = newNull
	newField.data.decoder = decoder
	newField.parent = n
	newField.value = name

	l := len(n.data.values)
	last := n
	if l > 0 {
		last = n.data.values[l-1].next
		if last.Type == Array || last.Type == Object {
			last = last.data.end
		}
	}
	last.next = newField

	n.data.values = append(n.data.values, newField)

	if n.data.flags&FlagFieldMap == FlagFieldMap {
		(*n.data.fields)[name] = l
	}

	return newNull
}

func (n *Node) AppendElement() *Node {
	if n == nil || n.Type != Array {
		return nil
	}

	decoder := n.data.decoder

	newNull := decoder.getNode()
	newNull.Type = Null
	newNull.next = n.data.end
	newNull.data.decoder = decoder
	newNull.parent = n

	l := len(n.data.values)
	last := n
	if l > 0 {
		last = n.data.values[l-1]
		if last.Type == Array || last.Type == Object {
			last = last.data.end
		}
	}
	last.next = newNull

	n.data.values = append(n.data.values, newNull)

	return newNull
}

func (n *Node) InsertElement(pos int) *Node {
	if n == nil || n.Type != Array {
		return nil
	}

	l := len(n.data.values)
	if pos < 0 || pos > l {
		return nil
	}

	prev := n
	if pos > 0 {
		prev = n.data.values[pos-1]
		if prev.Type == Array || prev.Type == Object {
			prev = prev.data.end
		}
	}

	decoder := n.data.decoder

	newNull := decoder.getNode()
	newNull.Type = Null
	newNull.next = n.data.end
	newNull.data.decoder = decoder
	newNull.parent = n

	prev.next = newNull
	if pos != l {
		newNull.next = n.data.values[pos]
	} else {
		newNull.next = n.data.end
	}

	leftPart := n.data.values[:pos]
	rightPart := n.data.values[pos:]

	n.data.values = make([]*Node, 0, 0)
	n.data.values = append(n.data.values, leftPart...)
	n.data.values = append(n.data.values, newNull)
	n.data.values = append(n.data.values, rightPart...)

	return newNull
}

// Suicide legendary insane suicide function
func (n *Node) Suicide() {
	if n == nil {
		return
	}

	owner := n.parent

	// root is immortal, sorry
	if owner == nil {
		return
	}

	workingIndex := n.actualizeIndex()
	// already deleted?
	if workingIndex == -1 {
		return
	}

	// mark owner as dirty
	owner.data.dirtySeq++

	switch owner.Type {
	case Object:
		lastIndex := len(owner.data.values) - 1
		deletingField := owner.data.values[workingIndex]
		if lastIndex == 0 {
			owner.next = owner.data.end
			owner.data.values = owner.data.values[:0]

			if owner.data.flags&FlagFieldMap == FlagFieldMap {
				delete(*owner.data.fields, deletingField.value)
			}

			return
		}

		movingField := owner.data.values[lastIndex]
		owner.data.values[workingIndex] = movingField
		owner.data.values = owner.data.values[:len(owner.data.values)-1]

		if workingIndex == 0 {
			owner.next = movingField
		} else {
			prevVal := owner.data.values[workingIndex-1].next
			if prevVal.Type == Object || prevVal.Type == Array {
				prevVal.data.end.next = movingField
			} else {
				prevVal.next = movingField
			}
		}

		if lastIndex != 0 {
			prevVal := owner.data.values[lastIndex-1].next
			if prevVal.Type == Object || prevVal.Type == Array {
				prevVal.data.end.next = owner.data.end
			} else {
				prevVal.next = owner.data.end
			}
		} else {
			owner.next = owner.data.end
		}

		nend := n
		if n.Type == Object || n.Type == Array {
			nend = n.data.end
		}

		if movingField != nend.next {
			if movingField.next.Type == Object || movingField.next.Type == Array {
				movingField.next.data.end.next = nend.next
			} else {
				movingField.next.next = nend.next
			}
		}

		if owner.data.flags&FlagFieldMap == FlagFieldMap {
			delete(*owner.data.fields, deletingField.value)
			if workingIndex != lastIndex {
				(*owner.data.fields)[movingField.value] = workingIndex
			}
		}
	case Array:
		deletingEl := n
		owner.data.values = append(owner.data.values[:workingIndex], owner.data.values[workingIndex+1:]...)

		var prev *Node
		if workingIndex == 0 {
			prev = owner
		} else {
			prev = owner.data.values[workingIndex-1]
			if prev.Type == Object || prev.Type == Array {
				prev = prev.data.end
			}
		}

		if deletingEl.Type == Object || deletingEl.Type == Array {
			prev.next = deletingEl.data.end.next
		} else {
			prev.next = deletingEl.next
		}

	default:
		panic("insane json really goes outta its mind")
	}
}

func (n *Node) tryDropLinks() {
	if n.Type != Object && n.Type != Array {
		return
	}

	index := n.actualizeIndex()
	if index == -1 {
		return
	}

	next := n.parent.data.end
	if index != len(n.parent.data.values)-1 {
		next = n.parent.data.values[index+1]
	}

	n.next = next
}

func (n *Node) actualizeIndex() int {
	owner := n.parent
	if owner == nil {
		return -1
	}

	// check if owner isn't dirty so nothing to do
	if n.data.dirtySeq != -1 && n.data.dirtySeq == owner.data.dirtySeq {

		return n.data.index
	}

	index := n.findSelf()
	n.data.index = index
	if owner.data.dirtySeq == -1 {
		owner.data.dirtySeq = 0
	}
	n.data.dirtySeq = owner.data.dirtySeq

	return index
}

func (n *Node) findSelf() int {
	owner := n.parent
	if owner == nil {
		return -1
	}

	index := -1
	if owner.Type == Array {
		for i, node := range owner.data.values {
			if node == n {
				index = i
				break
			}
		}
	} else {
		for i, node := range owner.data.values {
			if node.next == n {
				index = i
				break
			}
		}
	}
	return index
}

// ******************** //
//      MUTATIONS       //
// ******************** //

func (n *Node) MergeWith(node *Node) *Node {
	if n == nil || node == nil {
		return n
	}
	if !n.IsObject() || !node.IsObject() {
		return n
	}

	for _, child := range node.data.values {
		child.unescapeField()
		childField := child.AsString()
		x := n.AddField(childField)
		x.MutateToNode(child.next)
	}

	return n
}

// MutateToNode it isn't safe function, if you create node cycle, encode() may freeze
func (n *Node) MutateToNode(node *Node) *Node {
	if n == nil || node == nil {
		return n
	}

	n.tryDropLinks()

	curNext := n.next
	if n.Type == Object || n.Type == Array {
		curNext = n.data.end.next
	}

	if node.Type == Object || node.Type == Array {
		node.data.end.next = curNext
	} else {
		node.next = curNext
	}

	n.Type = node.Type
	n.value = node.value
	if node.Type == Object || node.Type == Array {
		n.next = node.next
		n.data.end = node.data.end
		n.data.end.parent = n.parent
		n.data.flags &= ^FlagFieldMap // reset field mapping
		n.data.values = append(n.data.values[:0], node.data.values...)
		for _, child := range node.data.values {
			child.parent = n
			if node.Type == Object {
				child.next.parent = n
			}
		}
	}

	return n
}

func (n *Node) MutateToJSON(json string) *Node {
	if n == nil {
		return n
	}

	node, err := n.data.decoder.decode(json, false)
	if err != nil {
		return n
	}

	return n.MutateToNode(node)
}

func (n *Node) MutateToField(value string) *Node {
	if n.Type != Field {
		return n
	}

	n.value = value

	return n
}

func (n *Node) MutateToInt(value int) *Node {
	if n.Type == Field {
		return n
	}
	n.tryDropLinks()

	n.Type = Number
	n.value = strconv.Itoa(value)

	return n
}

func (n *Node) MutateToFloat(value float64) *Node {
	if n.Type == Field {
		return n
	}
	n.tryDropLinks()

	n.Type = Number
	n.value = strconv.FormatFloat(value, 'f', -1, 64)

	return n
}

func (n *Node) MutateToBool(value bool) *Node {
	if n.Type == Field {
		return n
	}
	n.tryDropLinks()

	if value {
		n.Type = True
	} else {
		n.Type = False
	}

	return n
}

func (n *Node) MutateToNull(value bool) *Node {
	if n.Type == Field {
		return n
	}
	n.tryDropLinks()

	n.Type = Null

	return n
}

func (n *Node) MutateToString(value string) *Node {
	if n.Type == Field {
		return nil
	}
	n.tryDropLinks()

	n.Type = String
	n.value = value

	return n
}

func (n *Node) MutateToEscapedString(value string) *Node {
	if n.Type == Field {
		return nil
	}
	n.tryDropLinks()

	n.Type = escapedString
	n.value = value

	return n
}

func (n *Node) MutateToObject() *Node {
	if n.Type == Field {
		return n
	}
	n.tryDropLinks()

	n.Type = Object

	decoder := n.data.decoder

	objEnd := decoder.getNode()
	objEnd.Type = objectEnd
	objEnd.next = n.next
	objEnd.data.decoder = decoder
	objEnd.parent = n

	n.next = objEnd

	return n
}

func (n *Node) AsField(field string) *Node {
	if n == nil || n.Type != Object {
		return nil
	}

	value := n.Dig(field)
	if value == nil {
		return nil
	}

	return n.data.values[value.data.index]
}

func (n *StrictNode) AsField(field string) (*Node, error) {
	if n == nil || n.Type != Object {
		return nil, ErrNotObject
	}

	value := n.Dig(field)
	if value == nil {
		return nil, ErrNotFound
	}

	return n.data.values[value.data.index], nil
}

func (n *Node) AsFields() []*Node {
	if n == nil {
		return make([]*Node, 0, 0)
	}

	if n.Type != Object {
		return n.data.values[:0]
	}

	for _, node := range n.data.values {
		if node.Type == escapedField {
			node.unescapeField()
		}
	}

	return n.data.values
}

func (n *StrictNode) AsFields() ([]*Node, error) {
	if n.Type != Object {
		return nil, ErrNotObject
	}

	for _, node := range n.data.values {
		if node.Type == escapedField {
			node.unescapeField()
		}
	}

	return n.data.values, nil
}

func (n *Node) AsFieldValue() *Node {
	if n == nil || n.Type != Field {
		return nil
	}

	return n.next
}

func (n *StrictNode) AsFieldValue() (*Node, error) {
	if n == nil || n.Type != Field {
		return nil, ErrNotField
	}

	return n.next, nil
}

func (n *Node) AsArray() []*Node {
	if n == nil {
		return make([]*Node, 0, 0)
	}
	if n.Type != Array {
		return n.data.values[:0]
	}

	return n.data.values
}

func (n *StrictNode) AsArray() ([]*Node, error) {
	if n == nil || n.Type != Array {
		return nil, ErrNotArray
	}

	return n.data.values, nil
}

func (n *Node) unescapeStr() {
	value := n.value
	n.value = unescapeStr(value[1 : len(value)-1])
	n.Type = String
}

func (n *Node) unescapeField() {
	if n.Type == Field {
		return
	}

	value := n.value
	i := strings.LastIndexByte(value, '"')
	n.value = unescapeStr(value[1:i])
	n.Type = Field
}

func (n *Node) AsString() string {
	if n == nil {
		return ""
	}

	switch n.Type {
	case String:
		return n.value
	case escapedString:
		n.unescapeStr()
		return n.value
	case Number:
		return n.value
	case True:
		return "true"
	case False:
		return "false"
	case Null:
		return "null"
	case Field:
		return n.value
	case escapedField:
		panic("insane json really goes outta its mind")
	default:
		return ""
	}
}

func (n *Node) AsBytes() []byte {
	return toByte(n.AsString())
}

func (n *StrictNode) AsBytes() ([]byte, error) {
	s, err := n.AsString()
	if err != nil {
		return nil, err
	}

	return toByte(s), nil
}

func (n *StrictNode) AsString() (string, error) {
	if n.Type == escapedField {
		panic("insane json really goes outta its mind")
	}

	if n.Type == escapedString {
		n.unescapeStr()
	}

	if n == nil || n.Type != String {
		return "", ErrNotString
	}

	return n.value, nil
}

func (n *Node) AsEscapedString() string {
	if n == nil {
		return ""
	}

	switch n.Type {
	case String:
		return toString(escapeString(make([]byte, 0, 0), n.value))
	case escapedString:
		return n.value
	case Number:
		return n.value
	case True:
		return "true"
	case False:
		return "false"
	case Null:
		return "null"
	case Field:
		return n.value
	case escapedField:
		panic("insane json really goes outta its mind")
	default:
		return ""
	}
}

func (n *StrictNode) AsEscapedString() (string, error) {
	if n.Type == escapedField {
		panic("insane json really goes outta its mind")
	}

	if n == nil || n.Type != String {
		return "", ErrNotString
	}

	if n.Type == escapedString {
		return n.value, nil
	}

	return toString(escapeString(make([]byte, 0, 0), n.value)), nil
}

func (n *Node) AsBool() bool {
	if n == nil {
		return false
	}

	switch n.Type {
	case String:
		return n.value == "true"
	case escapedString:
		n.unescapeStr()
		return n.value == "true"
	case Number:
		return n.value != "0"
	case True:
		return true
	case False:
		return false
	case Null:
		return false
	case Field:
		return n.value == "true"
	case escapedField:
		panic("insane json really goes outta its mind")
	default:
		return false
	}
}

func (n *StrictNode) AsBool() (bool, error) {
	if n == nil || (n.Type != True && n.Type != False) {
		return false, ErrNotBool
	}

	return n.Type == True, nil
}

func (n *Node) AsInt() int {
	if n == nil {
		return 0
	}

	switch n.Type {
	case String:
		return int(math.Round(decodeFloat64(n.value)))
	case escapedString:
		n.unescapeStr()
		return int(math.Round(decodeFloat64(n.value)))
	case Number:
		return int(math.Round(decodeFloat64(n.value)))
	case True:
		return 1
	case False:
		return 0
	case Null:
		return 0
	case Field:
		return int(math.Round(decodeFloat64(n.value)))
	case escapedField:
		panic("insane json really goes outta its mind")
	default:
		return 0
	}
}

func (n *StrictNode) AsInt() (int, error) {
	if n == nil || n.Type != Number {
		return 0, ErrNotNumber
	}
	num := decodeInt64(n.value)
	if num == 0 && n.value != "0" {
		return 0, ErrNotNumber
	}
	return int(num), nil
}

func (n *Node) AsFloat() float64 {
	switch n.Type {
	case String:
		return decodeFloat64(n.value)
	case escapedString:
		n.unescapeStr()
		return decodeFloat64(n.value)
	case Number:
		return decodeFloat64(n.value)
	case True:
		return 1
	case False:
		return 0
	case Null:
		return 0
	case Field:
		return decodeFloat64(n.value)
	case escapedField:
		panic("insane json really goes outta its mind")
	default:
		return 0
	}
}

func (n *StrictNode) AsFloat() (float64, error) {
	if n == nil || n.Type != Number {
		return 0, ErrNotNumber
	}

	return decodeFloat64(n.value), nil
}

func (n *Node) IsObject() bool {
	return n != nil && n.Type == Object
}

func (n *Node) IsArray() bool {
	return n != nil && n.Type == Array
}

func (n *Node) IsNumber() bool {
	return n != nil && n.Type == Number
}

func (n *Node) IsString() bool {
	return n != nil && (n.Type == String || n.Type == escapedString)
}

func (n *Node) IsTrue() bool {
	return n != nil && n.Type == True
}

func (n *Node) IsFalse() bool {
	return n != nil && n.Type == False
}

func (n *Node) IsNull() bool {
	return n != nil && n.Type == Null
}

func (n *Node) IsField() bool {
	return n != nil && n.Type == Field
}

func (n *Node) IsNil() bool {
	return n == nil
}

func (n *Node) InStrictMode() *StrictNode {
	n.data.err.Node = n
	return n.data.err
}

// ******************** //
//       DECODER        //
// ******************** //

func (d *decoder) initPool() {
	d.nodePool = make([]*Node, StartNodePoolSize, StartNodePoolSize)
	for i := 0; i < StartNodePoolSize; i++ {
		d.nodePool[i] = &Node{data: &data{decoder: d}}
	}
}

func (d *decoder) expandPool() []*Node {
	c := cap(d.nodePool)
	for i := 0; i < c; i++ {
		d.nodePool = append(d.nodePool, &Node{data: &data{decoder: d}})
	}

	return d.nodePool
}

func getFromPool() *decoder {
	decoderPoolMu.Lock()
	defer decoderPoolMu.Unlock()

	decoderPoolIndex++

	if decoderPoolIndex > len(decoderPool)-1 || decoderPool[decoderPoolIndex] == nil {
		decoder := &decoder{id: decoderPoolIndex}
		decoder.initPool()
		decoderPool = append(decoderPool, decoder)
	}

	return decoderPool[decoderPoolIndex]
}

func backToPool(d *decoder) {
	decoderPoolMu.Lock()
	defer decoderPoolMu.Unlock()

	cur := d.id

	decoderPool[cur] = decoderPool[decoderPoolIndex]
	decoderPool[cur].id = cur

	decoderPoolIndex--
}

func Spawn() *Root {
	root, _ := getFromPool().decodeHeadless("{}", true)
	return root
}

func DecodeBytes(jsonBytes []byte) (*Root, error) {
	return Spawn().data.decoder.decodeHeadless(toString(jsonBytes), true)
}

func DecodeString(json string) (*Root, error) {
	return Spawn().data.decoder.decodeHeadless(json, true)
}

// DecodeBytes clear root and decode new JSON
// useful for reusing root to decode multiple times and reduce allocations
func (r *Root) DecodeBytes(jsonBytes []byte) error {
	if r == nil {
		return ErrRootIsNil
	}
	_, err := r.data.decoder.decodeHeadless(toString(jsonBytes), false)

	return err
}

// DecodeString clear root and decode new JSON
// useful for reusing root to decode multiple times and reduce allocations
func (r *Root) DecodeString(json string) error {
	if r == nil {
		return ErrRootIsNil
	}
	_, err := r.data.decoder.decodeHeadless(json, false)

	return err
}

// DecodeBytesAdditional doesn't clean root, uses root's node pool to decode JSON
func (r *Root) DecodeBytesAdditional(jsonBytes []byte) (*Node, error) {
	if r == nil {
		return nil, ErrRootIsNil
	}

	return r.data.decoder.decode(toString(jsonBytes), false)
}

// DecodeStringAdditional doesn't clean root, uses root's node pool to decode JSON
func (r *Root) DecodeStringAdditional(json string) (*Node, error) {
	if r == nil {
		return nil, ErrRootIsNil
	}

	return r.data.decoder.decode(json, false)
}

func Release(root *Root) {
	if root == nil {
		return
	}

	backToPool(root.data.decoder)
}

func toString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func toByte(s string) []byte {
	header := (*reflect.StringHeader)(unsafe.Pointer(&s))
	slice := reflect.SliceHeader{
		Data: header.Data,
		Len:  header.Len,
		Cap:  header.Len,
	}

	return *(*[]byte)(unsafe.Pointer(&slice))
}

// this code copied from really cool and fast https://github.com/valyala/fastjson
func unescapeStr(s string) string {
	n := strings.IndexByte(s, '\\')
	if n < 0 {
		return s
	}

	b := toByte(s)
	b = b[:n]
	s = s[n+1:]
	for len(s) > 0 {
		ch := s[0]
		s = s[1:]
		switch ch {
		case '"':
			b = append(b, '"')
		case '\\':
			b = append(b, '\\')
		case '/':
			b = append(b, '/')
		case 'b':
			b = append(b, '\b')
		case 'f':
			b = append(b, '\f')
		case 'n':
			b = append(b, '\n')
		case 'r':
			b = append(b, '\r')
		case 't':
			b = append(b, '\t')
		case 'u':
			if len(s) < 4 {
				b = append(b, "\\u"...)
				break
			}
			xs := s[:4]
			x, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				b = append(b, "\\u"...)
				break
			}
			s = s[4:]
			if !utf16.IsSurrogate(rune(x)) {
				b = append(b, string(rune(x))...)
				break
			}

			if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			x1, err := strconv.ParseUint(s[2:6], 16, 16)
			if err != nil {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			r := utf16.DecodeRune(rune(x), rune(x1))
			b = append(b, string(r)...)
			s = s[6:]
		default:
			b = append(b, '\\', ch)
		}
		n = strings.IndexByte(s, '\\')
		if n < 0 {
			b = append(b, s...)
			break
		}
		b = append(b, s[:n]...)
		s = s[n+1:]
	}
	return toString(b)
}

func escapeString(out []byte, st string) []byte {
	if !shouldBeEscaped(st) {
		out = append(out, '"')
		out = append(out, st...)
		out = append(out, '"')
		return out
	}

	out = append(out, '"')
	s := toByte(st)
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' && b != '<' && b != '>' && b != '&' {
				i++
				continue
			}
			if start < i {
				out = append(out, s[start:i]...)
			}
			switch b {
			case '\\', '"':
				out = append(out, '\\')
				out = append(out, b)
			case '\n':
				out = append(out, "\\n"...)
			case '\r':
				out = append(out, "\\r"...)
			case '\t':
				out = append(out, "\\t"...)
			default:
				out = append(out, "\\u00"...)
				out = append(out, hex[b>>4])
				out = append(out, hex[b&0xf])
			}
			i++
			start = i
			continue
		}

		c, size := utf8.DecodeRune(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				out = append(out, s[start:i]...)
			}
			out = append(out, "\\ufffd"...)
			i += size
			start = i
			continue
		}

		if c == '\u2028' || c == '\u2029' {
			if start < i {
				out = append(out, s[start:i]...)
			}
			out = append(out, "\\u202"...)
			out = append(out, hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		out = append(out, s[start:]...)
	}
	out = append(out, '"')

	return out
}

func shouldBeEscaped(s string) bool {
	if strings.IndexByte(s, '"') >= 0 || strings.IndexByte(s, '\\') >= 0 {
		return true
	}

	l := len(s)
	for i := 0; i < l; i++ {
		if s[i] < 0x20 {
			return true
		}
	}

	return false
}

func decodeInt64(s string) int64 {
	l := len(s)
	if l == 0 {
		return 0
	}

	o := 0
	m := s[0] == '-'
	if m {
		s = s[1:]
		l--
	}

	num := int64(0)
	for o < l {
		c := uint(s[o] - '0')
		if c > 9 {
			return 0
		}

		num = num*10 + int64(c)
		o++
		if o <= 18 {
			continue
		}

		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0
		}
		num = x
		break
	}

	if m {
		return -num
	} else {
		return num
	}
}

// this code copied from really cool and fast https://github.com/valyala/fastjson
func decodeFloat64(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	i := uint(0)
	minus := s[0] == '-'
	if minus {
		i++
		if i >= uint(len(s)) {
			return 0
		}
	}

	d := uint64(0)
	j := i
	for i < uint(len(s)) {
		if s[i] >= '0' && s[i] <= '9' {
			d = d*10 + uint64(s[i]-'0')
			i++
			if i > 18 {
				// The integer part may be out of range for uint64.
				// Fall back to slow parsing.
				f, err := strconv.ParseFloat(s, 64)
				if err != nil && !math.IsInf(f, 0) {
					return 0
				}
				return f
			}
			continue
		}
		break
	}
	if i <= j {
		return 0
	}
	f := float64(d)
	if i >= uint(len(s)) {
		// Fast path - just integer.
		if minus {
			f = -f
		}
		return f
	}

	if s[i] == '.' {
		// Parse fractional part.
		i++
		if i >= uint(len(s)) {
			return 0
		}
		fr := uint64(0)
		j := i
		for i < uint(len(s)) {
			if s[i] >= '0' && s[i] <= '9' {
				fr = fr*10 + uint64(s[i]-'0')
				i++
				if i-j > 18 {
					// The fractional part may be out of range for uint64.
					// Fall back to standard parsing.
					f, err := strconv.ParseFloat(s, 64)
					if err != nil && !math.IsInf(f, 0) {
						return 0
					}
					return f
				}
				continue
			}
			break
		}
		if i <= j {
			return 0
		}
		f += float64(fr) / math.Pow10(int(i-j))
		if i >= uint(len(s)) {
			// Fast path - parsed fractional number.
			if minus {
				f = -f
			}
			return f
		}
	}
	if s[i] == 'e' || s[i] == 'E' {
		// Parse exponent part.
		i++
		if i >= uint(len(s)) {
			return 0
		}
		expMinus := false
		if s[i] == '+' || s[i] == '-' {
			expMinus = s[i] == '-'
			i++
			if i >= uint(len(s)) {
				return 0
			}
		}
		exp := int16(0)
		j := i
		for i < uint(len(s)) {
			if s[i] >= '0' && s[i] <= '9' {
				exp = exp*10 + int16(s[i]-'0')
				i++
				if exp > 300 {
					// The exponent may be too big for float64.
					// Fall back to standard parsing.
					f, err := strconv.ParseFloat(s, 64)
					if err != nil && !math.IsInf(f, 0) {
						return 0
					}
					return f
				}
				continue
			}
			break
		}
		if i <= j {
			return 0
		}
		if expMinus {
			exp = -exp
		}
		f *= math.Pow10(int(exp))
		if i >= uint(len(s)) {
			if minus {
				f = -f
			}
			return f
		}
	}
	return 0
}

var out = make([]byte, 0, 0)
var root = Spawn()

func Fuzz(data []byte) int {
	err := root.DecodeBytes(data)
	if err != nil {
		return -1
	}

	fields := []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
		"11", "21", "31", "41", "51", "61", "71", "81", "91", "101",
		"111", "211", "311", "411", "511", "611", "711", "811", "911", "1011",
	}
	jsons := []string{
		"1", "2", "3", "4", "5",
		`{"a":"b","c":"d"}`,
		`{"5":"5","l":[3,4]}`,
		`{"a":{"5":"5","l":[3,4]},"c":"d"}`,
		`{"a":"b","c":"d"}`,
		`{"5":"5","l":[3,4]}`,
		`{"a":"b","c":{"5":"5","l":[3,4]}}`,
		`{"a":{"somekey":"someval", "xxx":"yyy"},"c":"d"}`,
		`{"5":"5","l":[3,4]}`,
		`["a","b","c","d"]`,
		`{"5":"5","l":[3,4]}`,
		`[{"5":"5","l":[3,4]},"b","c","d"]`,
		`["a","b","c","d"]`,
		`{"5":"5","l":[3,4]}`,
		`["a","b","c",{"5":"5","l":[3,4]}]`,
	}

	node := root.Node

	for i := 0; i < 40; i++ {
		for j := 0; j < 100; j++ {
			if node.IsObject() {
				fields := node.AsFields()
				if len(fields) == 0 {
					break
				}
				node = node.Dig(fields[rand.Int()%len(fields)].AsString())
				continue
			}
			if node.IsArray() {
				fields := node.AsArray()
				if len(fields) == 0 {
					break
				}
				node = node.Dig(strconv.Itoa(rand.Int() % len(fields)))
				continue
			}

			node.AddField(fields[rand.Int()%len(fields)]).MutateToJSON(jsons[rand.Int()%len(jsons)])
			break
		}
		for j := 0; j < 200; j++ {
			if node.IsObject() {
				fields := node.AsFields()
				if len(fields) == 0 {
					break
				}
				node = node.Dig(fields[rand.Int()%len(fields)].AsString())
				continue
			}
			if node.IsArray() {
				fields := node.AsArray()
				if len(fields) == 0 {
					break
				}
				node = node.Dig(strconv.Itoa(rand.Int() % len(fields)))
				continue
			}

			node.Suicide()
			node = root.Node
			for ; node != nil; node = node.next {
			}
			node = root.Node
			break
		}
	}

	root.Encode(out)

	return 1
}

func insaneErr(err error, json string, offset int) error {
	a := offset - 20
	b := offset + 20
	if a < 0 {
		a = 0
	}
	if b > len(json)-1 {
		b = len(json) - 1
		if b < 0 {
			b = 0
		}
	}

	pointer := ""
	for x := 0; x < offset-a+len(err.Error())+len(" near "); x++ {
		pointer += " "
	}
	pointer += "^"
	str := ""
	if a != b {
		str = strings.ReplaceAll(json[a:b], "\n", " ")
	}
	str = strings.ReplaceAll(str, "\r", " ")
	str = strings.ReplaceAll(str, "\t", " ")

	return errors.New(fmt.Sprintf("%s near `%s`\n%s", err.Error(), str, pointer))
}
