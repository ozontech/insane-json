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

// 0-11  bits – node type
// 12-35 bits – node index
// 36-59 bits – dirty sequence
// 60    bit  – map usage
type hellBits uint64

const (
	hellBitObject        hellBits = 1 << 0
	hellBitEnd           hellBits = 1 << 1
	hellBitArray         hellBits = 1 << 2
	hellBitArrayEnd      hellBits = 1 << 3
	hellBitString        hellBits = 1 << 4
	hellBitEscapedString hellBits = 1 << 5
	hellBitNumber        hellBits = 1 << 6
	hellBitTrue          hellBits = 1 << 7
	hellBitFalse         hellBits = 1 << 8
	hellBitNull          hellBits = 1 << 9
	hellBitField         hellBits = 1 << 10
	hellBitEscapedField  hellBits = 1 << 11
	hellBitTypeFilter    hellBits = 1<<11 - 1

	hellBitUseMap       hellBits = 1 << 60
	hellBitsUseMapReset          = 1<<64 - 1 - hellBitUseMap

	hellBitsDirtyFilter          = 0x0FFFFFF000000000
	hellBitsDirtyReset  hellBits = 0xF000000FFFFFFFFF
	hellBitsDirtyStep   hellBits = 1 << 36

	hellBitsIndexFilter hellBits = 0x0000000FFFFFF000
	hellBitsIndexReset  hellBits = 0xFFFFFFF000000FFF
	hellBitsIndexStep   hellBits = 1 << 12

	hex = "0123456789abcdef"
)

var (
	StartNodePoolSize = 128
	MapUseThreshold   = 16

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

func init() {
	numbersMap['.'] = 1
	numbersMap['-'] = 1
	numbersMap['e'] = 1
	numbersMap['E'] = 1
	numbersMap['+'] = 1
}

/*
Node Is a building block of the decoded JSON. There is seven basic nodes:
	1. Object
	2. Array
	3. String
	4. Number
	5. True
	6. False
	7. Null
And a special one – Field, which represents the field(key) on an objects.
It allows to easily change field's name, checkout MutateToField() function.
*/
type Node struct {
	bits   hellBits
	data   string
	next   *Node
	parent *Node
	nodes  []*Node
	fields *map[string]int
}

/*
Root is a top Node of decoded JSON. It holds decoder, current JSON data and pool of Nodes.
Node pool is used to reduce memory allocations and GC time.
Checkout ReleaseMem()/ReleasePoolMem()/ReleaseBufMem() to clear pools.
Root can be reused to decode another JSON using DecodeBytes()/DecodeString().
Also Root can decode additional JSON using DecodeAdditionalBytes()/DecodeAdditionalString().
*/
type Root struct {
	*Node
	decoder *decoder
}

/*
StrictNode implements API with error handling.
Transform any Node with MutateToStrict(), Mutate*()/As*() functions will return an error
*/
type StrictNode struct {
	*Node
}

type decoder struct {
	id        int
	buf       []byte
	root      Root
	nodePool  []*Node
	nodeCount int
}

/*
ReleaseMem sends node pool and internal buffer to GC.
Useful to reduce memory usage after decoding big JSON.
*/
func (r *Root) ReleaseMem() {
	r.ReleasePoolMem()
	r.ReleaseBufMem()
}

/*
ReleasePoolMem sends node pool to GC.
Useful to reduce memory usage after decoding big JSON.
*/
func (r *Root) ReleasePoolMem() {
	r.decoder.initPool()
}

/*
ReleaseBufMem sends internal buffer to GC.
Useful to reduce memory usage after decoding big JSON.
*/
func (r *Root) ReleaseBufMem() {
	r.decoder.buf = make([]byte, 0, 0)
}

/*
BuffCap returns current size of internal buffer.
*/
func (r *Root) BuffCap() int {
	return cap(r.decoder.buf)
}

/*
PoolSize returns how many Node objects is in the pool right now.
*/
func (r *Root) PoolSize() int {
	return len(r.decoder.nodePool)
}

// ******************** //
//      MAIN SHIT       //
// ******************** //

// decode is a legendary function for decoding JSONs
func (d *decoder) decode(json string, shouldReset bool) (*Node, error) {
	if shouldReset {
		d.nodeCount = 0
		d.buf = d.buf[:0]
	}
	o := len(d.buf)

	d.buf = append(d.buf, json...)
	json = toString(d.buf)

	l := len(json)
	if l == 0 {
		return nil, insaneErr(ErrEmptyJSON, cp(json), o)
	}

	nodePool := d.nodePool
	nodePoolLen := len(nodePool)
	nodes := d.nodeCount

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
		return nil, insaneErr(ErrUnexpectedJSONEnding, cp(json), o)
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

		curNode.bits = hellBitEnd
		curNode.parent = topNode

		topNode.next = nodePool[nodes]
		topNode = topNode.parent

		goto pop
	}

	if c != ',' {
		if len(topNode.nodes) > 0 {
			return nil, insaneErr(ErrExpectedComma, cp(json), o)
		}
		o--
	} else {
		if len(topNode.nodes) == 0 {
			return nil, insaneErr(ErrExpectedObjectField, cp(json), o)
		}
		if o == l {
			return nil, insaneErr(ErrUnexpectedJSONEnding, cp(json), o)
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
		return nil, insaneErr(ErrExpectedObjectField, cp(json), o)
	}

	t = o - 1
	for {
		x = strings.IndexByte(json[o:], '"')
		o += x + 1
		if x < 0 {
			return nil, insaneErr(ErrUnexpectedEndOfObjectField, cp(json), o)
		}

		if x == 0 || json[o-2] != '\\' {
			break
		}

		// untangle fucking escaping hell
		z := o - 3
		for json[z] == '\\' {
			z--
		}
		if (o-z)%2 == 0 {
			break
		}

	}
	if o == l {
		return nil, insaneErr(ErrExpectedObjectFieldSeparator, cp(json), o)
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
		return nil, insaneErr(ErrExpectedObjectFieldSeparator, cp(json), o)
	}
	if o == l {
		return nil, insaneErr(ErrExpectedValue, cp(json), o)
	}
	curNode.bits = hellBitEscapedField
	curNode.data = json[t:o]
	curNode.parent = topNode
	topNode.nodes = append(topNode.nodes, curNode)

	goto decode
decodeArray:
	if o == l {
		return nil, insaneErr(ErrUnexpectedJSONEnding, cp(json), o)
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

		curNode.bits = hellBitArrayEnd
		curNode.parent = topNode

		topNode.next = nodePool[nodes]
		topNode = topNode.parent

		goto pop
	}

	if c != ',' {
		if len(topNode.nodes) > 0 {
			return nil, insaneErr(ErrExpectedComma, cp(json), o)
		}
		o--
	} else {
		if len(topNode.nodes) == 0 {
			return nil, insaneErr(ErrExpectedValue, cp(json), o)
		}
		if o == l {
			return nil, insaneErr(ErrUnexpectedJSONEnding, cp(json), o)
		}
	}

	topNode.nodes = append(topNode.nodes, nodePool[nodes])
decode:
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
			return nil, insaneErr(ErrExpectedObjectField, cp(json), o)
		}

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.bits = hellBitObject
		curNode.nodes = curNode.nodes[:0]
		curNode.parent = topNode

		topNode = curNode
		if nodes >= nodePoolLen-1 {
			nodePool = d.expandPool()
			nodePoolLen = len(nodePool)
		}
		goto decodeObject
	case '[':
		if o == l {
			return nil, insaneErr(ErrExpectedValue, cp(json), o)
		}
		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.bits = hellBitArray
		curNode.nodes = curNode.nodes[:0]
		curNode.parent = topNode

		topNode = curNode
		if nodes >= nodePoolLen-1 {
			nodePool = d.expandPool()
			nodePoolLen = len(nodePool)
		}
		goto decodeArray
	case '"':
		t = o
		for {
			x := strings.IndexByte(json[t:], '"')
			t += x + 1
			if x < 0 {
				return nil, insaneErr(ErrUnexpectedEndOfString, cp(json), o)
			}
			if x == 0 || json[t-2] != '\\' {
				break
			}

			// untangle fucking escaping hell
			z := t - 3
			for json[z] == '\\' {
				z--
			}
			if (t-z)%2 == 0 {
				break
			}
		}

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.bits = hellBitEscapedString
		curNode.data = json[o-1 : t]
		curNode.parent = topNode

		o = t
	case 't':
		if len(json) < o+3 || json[o:o+3] != "rue" {
			return nil, insaneErr(ErrUnexpectedEndOfTrue, cp(json), o)
		}
		o += 3

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.bits = hellBitTrue
		curNode.parent = topNode

	case 'f':
		if len(json) < o+4 || json[o:o+4] != "alse" {
			return nil, insaneErr(ErrUnexpectedEndOfFalse, cp(json), o)
		}
		o += 4

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.bits = hellBitFalse
		curNode.parent = topNode

	case 'n':
		if len(json) < o+3 || json[o:o+3] != "ull" {
			return nil, insaneErr(ErrUnexpectedEndOfNull, cp(json), o)
		}
		o += 3

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.bits = hellBitNull
		curNode.parent = topNode
	default:
		o--
		t = o
		for ; o != l && ((json[o] >= '0' && json[o] <= '9') || numbersMap[json[o]] == 1); o++ {
		}
		if t == o {
			return nil, insaneErr(ErrExpectedValue, cp(json), o)
		}

		curNode.next = nodePool[nodes]
		curNode = curNode.next
		nodes++

		curNode.bits = hellBitNumber
		curNode.data = json[t:o]
		curNode.parent = topNode
	}
pop:
	if topNode == nil {
		goto exit
	}

	if nodes >= nodePoolLen-1 {
		nodePool = d.expandPool()
		nodePoolLen = len(nodePool)
	}

	if topNode.bits&hellBitObject == hellBitObject {
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
			return nil, insaneErr(ErrUnexpectedJSONEnding, cp(json), o)
		}
	}

	root.next = nil
	curNode.next = nil
	d.nodeCount = nodes

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
	d.root.decoder = d

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

	if len(curNode.nodes) == 0 {
		if curNode.bits&hellBitObject == hellBitObject {
			return append(out, "{}"...)
		}
		if curNode.bits&hellBitArray == hellBitArray {
			return append(out, "[]"...)
		}
	}

	goto encodeSkip
encode:
	out = append(out, ","...)
encodeSkip:
	switch curNode.bits & hellBitTypeFilter {
	case hellBitObject:
		if len(curNode.nodes) == 0 {
			out = append(out, "{}"...)
			curNode = curNode.next
			goto popSkip
		}
		topNode = curNode
		out = append(out, '{')
		curNode = curNode.nodes[0]
		if curNode.bits&hellBitField == hellBitField {
			out = escapeString(out, curNode.data)
			out = append(out, ':')
		} else {
			out = append(out, curNode.data...)
		}
		curNode = curNode.next
		s++
		goto encodeSkip
	case hellBitArray:
		if len(curNode.nodes) == 0 {
			out = append(out, "[]"...)
			curNode = curNode.next
			goto popSkip
		}
		topNode = curNode
		out = append(out, '[')
		curNode = curNode.nodes[0]
		s++
		goto encodeSkip
	case hellBitNumber:
		out = append(out, curNode.data...)
	case hellBitString:
		out = escapeString(out, curNode.data)
	case hellBitEscapedString:
		out = append(out, curNode.data...)
	case hellBitFalse:
		out = append(out, "false"...)
	case hellBitTrue:
		out = append(out, "true"...)
	case hellBitNull:
		out = append(out, "null"...)
	}
pop:
	curNode = curNode.next
popSkip:
	if topNode.bits&hellBitArray == hellBitArray {
		if curNode.bits&hellBitArrayEnd == hellBitArrayEnd {
			out = append(out, "]"...)
			curNode = topNode
			topNode = topNode.parent
			s--
			if s == 0 {
				return out
			}
			goto pop
		}
		goto encode
	} else if topNode.bits&hellBitObject == hellBitObject {
		if curNode.bits&hellBitEnd == hellBitEnd {
			out = append(out, "}"...)
			curNode = topNode
			topNode = topNode.parent
			s--
			if s == 0 {
				return out
			}
			goto pop
		}
		out = append(out, ","...)
		if curNode.bits&hellBitField == hellBitField {
			out = escapeString(out, curNode.data)
			out = append(out, ':')
		} else {
			out = append(out, curNode.data...)
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

	maxDepth := len(path)
	if maxDepth == 0 {
		return n
	}

	node := n
	curField := path[0]
	curDepth := 0
get:
	if node.bits&hellBitArray == hellBitArray {
		goto getArray
	}

	if len(node.nodes) > MapUseThreshold {
		if node.bits&hellBitUseMap != hellBitUseMap {
			var m map[string]int
			if node.fields == nil {
				m = make(map[string]int, len(node.nodes))
				node.fields = &m
			} else {
				m = *node.fields
				for field := range m {
					delete(m, field)
				}
			}

			for index, field := range node.nodes {
				if field.bits&hellBitEscapedField == hellBitEscapedField {
					field.unescapeField()
				}
				m[field.data] = index
			}
			node.bits |= hellBitUseMap
		}

		if node.bits&hellBitUseMap == hellBitUseMap {
			index, has := (*node.fields)[curField]
			if !has {
				return nil
			}

			curDepth++
			if curDepth == maxDepth {
				result := (node.nodes)[index].next
				result.bits = result.bits&hellBitsDirtyReset | (node.bits & hellBitsDirtyFilter)
				result.setIndex(index)

				return result
			}

			curField = path[curDepth]
			node = (node.nodes)[index].next
			goto get
		}
	}

	for index, field := range node.nodes {
		if field.bits&hellBitEscapedField == hellBitEscapedField {
			field.unescapeField()
		}

		if field.data == curField {
			curDepth++
			if curDepth == maxDepth {
				result := field.next
				result.bits = result.bits&hellBitsDirtyReset | (node.bits & hellBitsDirtyFilter)
				result.setIndex(index)

				return result
			}
			curField = path[curDepth]
			node = field.next
			goto get
		}
	}
	return nil
getArray:
	index, err := strconv.Atoi(curField)
	if err != nil || index < 0 || index >= len(node.nodes) {
		return nil
	}
	curDepth++
	if curDepth == maxDepth {
		result := (node.nodes)[index]
		result.bits = result.bits&hellBitsDirtyReset | (node.bits & hellBitsDirtyFilter)
		result.setIndex(index)

		return result
	}
	curField = path[curDepth]
	node = (node.nodes)[index]
	goto get
}

func (d *decoder) getNode() *Node {
	node := d.nodePool[d.nodeCount]
	d.nodeCount++
	if d.nodeCount > len(d.nodePool)-16 {
		d.expandPool()
	}

	return node
}

func (n *Node) DigStrict(path ...string) (*StrictNode, error) {
	result := n.Dig(path...)
	if result == nil {
		return nil, ErrNotFound
	}

	return result.MutateToStrict(), nil
}

func (n *Node) AddField(name string) *Node {
	if n == nil || n.bits&hellBitObject != hellBitObject {
		return nil
	}

	node := n.Dig(name)
	if node != nil {
		return node
	}

	newNull := &Node{}
	newNull.bits = hellBitNull
	newNull.parent = n

	newField := n.getNode()
	newField.bits = hellBitField
	newField.next = newNull
	newField.parent = n
	newField.data = name

	l := len(n.nodes)
	if l > 0 {
		lastVal := (n.nodes)[l-1]
		newNull.next = lastVal.next.next
		lastVal.next.next = newField
	} else {
		// restore lost end
		newEnd := n.getNode()
		newEnd.bits = hellBitEnd
		newEnd.next = n.next
		newEnd.parent = n
		newNull.next = newEnd
	}
	n.nodes = append(n.nodes, newField)

	if n.bits&hellBitUseMap == hellBitUseMap {
		(*n.fields)[name] = l
	}

	return newNull
}

func (n *Node) AddElement() *Node {
	if n == nil || n.bits&hellBitArray != hellBitArray {
		return nil
	}

	newNull := n.getNode()
	newNull.bits = hellBitNull
	newNull.parent = n

	l := len(n.nodes)
	if l > 0 {
		lastVal := (n.nodes)[l-1]
		newNull.next = lastVal.next
		lastVal.next = newNull
	} else {
		// restore lost end
		newEnd := n.getNode()
		newEnd.bits = hellBitArrayEnd
		newEnd.next = n.next
		newEnd.parent = n
		newNull.next = newEnd
	}
	n.nodes = append(n.nodes, newNull)

	return newNull
}

func (n *Node) InsertElement(pos int) *Node {
	if n == nil || n.bits&hellBitArray != hellBitArray {
		return nil
	}

	l := len(n.nodes)
	if pos < 0 || pos > l {
		return nil
	}

	newNull := n.getNode()
	newNull.bits = hellBitNull
	newNull.parent = n

	if l == 0 {
		// restore lost end
		newEnd := n.getNode()
		newEnd.bits = hellBitArrayEnd
		newEnd.next = n.next
		newEnd.parent = n
		newNull.next = newEnd
	} else {
		if pos != l {
			newNull.next = n.nodes[pos]
		} else {
			newNull.next = n.nodes[pos-1].next
		}
	}

	if pos > 0 {
		n.nodes[pos-1].next = newNull
	}

	leftPart := n.nodes[:pos]
	rightPart := n.nodes[pos:]

	n.nodes = make([]*Node, 0, 0)
	n.nodes = append(n.nodes, leftPart...)
	n.nodes = append(n.nodes, newNull)
	n.nodes = append(n.nodes, rightPart...)

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

	delIndex := n.actualizeIndex()
	// already deleted?
	if delIndex == -1 {
		return
	}

	// mark owner as dirty
	owner.bits += hellBitsDirtyStep

	switch owner.bits & hellBitTypeFilter {
	case hellBitObject:
		moveIndex := len(owner.nodes) - 1
		delField := owner.nodes[delIndex]
		if moveIndex == 0 {
			owner.nodes = owner.nodes[:0]

			if owner.bits&hellBitUseMap == hellBitUseMap {
				delete(*owner.fields, delField.data)
			}

			return
		}

		lastField := owner.nodes[moveIndex]
		owner.nodes[delIndex] = lastField

		if delIndex != 0 {
			owner.nodes[delIndex-1].next.next = lastField
		}

		owner.nodes[moveIndex-1].next.next = owner.nodes[moveIndex].next.next
		if lastField != n.next {
			lastField.next.next = n.next
		}

		if owner.bits&hellBitUseMap == hellBitUseMap {
			delete(*owner.fields, delField.data)
			if delIndex != moveIndex {
				(*owner.fields)[lastField.data] = delIndex
			}
		}
		owner.nodes = owner.nodes[:len(owner.nodes)-1]

	case hellBitArray:
		if delIndex != 0 {
			owner.nodes[delIndex-1].next = n.next
		}
		owner.nodes = append(owner.nodes[:delIndex], owner.nodes[delIndex+1:]...)
	default:
		panic("insane json really goes outta its mind")
	}
}

func (n *Node) actualizeIndex() int {
	owner := n.parent
	if owner == nil {
		return -1
	}

	a := n.bits & hellBitsDirtyFilter
	b := owner.bits & hellBitsDirtyFilter

	//  if owner isn't dirty then nothing to do
	if a != 0 && a == b {
		return n.getIndex()
	}

	index := n.findSelf()
	n.setIndex(index)
	n.bits = n.bits&hellBitsDirtyReset | (owner.bits & hellBitsDirtyFilter)

	return index
}

func (n *Node) findSelf() int {
	owner := n.parent
	if owner == nil {
		return -1
	}

	index := -1
	if owner.bits&hellBitArray == hellBitArray {
		for i, node := range owner.nodes {
			if node == n {
				index = i
				break
			}
		}
	} else {
		for i, node := range owner.nodes {
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
	if n.bits&hellBitObject != hellBitObject || node.bits&hellBitObject != hellBitObject {
		return n
	}

	for _, child := range node.nodes {
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

	n.bits = node.bits
	n.data = node.data
	if node.bits&hellBitObject == hellBitObject || node.bits&hellBitArray == hellBitArray {
		n.bits &= hellBitsUseMapReset
		n.nodes = append((n.nodes)[:0], node.nodes...)
		for _, child := range node.nodes {
			child.parent = n
			if node.bits&hellBitObject == hellBitObject {
				child.next.parent = n
			}
		}
	}

	return n
}

func (n *Node) MutateToJSON(root *Root, json string) *Node {
	if n == nil {
		return n
	}

	node, err := root.decoder.decode(json, false)
	if err != nil {
		return n
	}

	return n.MutateToNode(node)
}

// MutateToField changes name of objects's field
// works only with Field nodes received by AsField()/AsFields()
// example:
// root, err := insaneJSON.DecodeString(`{"a":"a","b":"b"}`)
// root.AsField("a").MutateToField("new_name")
// root.Encode() will be {"new_name":"a","b":"b"}
func (n *Node) MutateToField(newFieldName string) *Node {
	if n == nil || n.bits&hellBitField != hellBitField {
		return n
	}

	parent := n.parent
	if parent.bits&hellBitUseMap == hellBitUseMap {
		x := (*parent.fields)[n.data]
		delete(*parent.fields, n.data)
		(*parent.fields)[newFieldName] = x
	}

	n.data = newFieldName

	return n
}

func (n *Node) MutateToInt(value int) *Node {
	if n == nil || n.bits&hellBitField == hellBitField {
		return n
	}

	n.bits = hellBitNumber
	n.data = strconv.Itoa(value)

	return n
}

func (n *Node) MutateToFloat(value float64) *Node {
	if n == nil || n.bits&hellBitField == hellBitField {
		return n
	}

	n.bits = hellBitNumber
	n.data = strconv.FormatFloat(value, 'f', -1, 64)

	return n
}

func (n *Node) MutateToBool(value bool) *Node {
	if n == nil || n.bits&hellBitField == hellBitField {
		return n
	}

	if value {
		n.bits = hellBitTrue
	} else {
		n.bits = hellBitFalse
	}

	return n
}

func (n *Node) MutateToNull(value bool) *Node {
	if n == nil || n.bits&hellBitField == hellBitField {
		return n
	}

	n.bits = hellBitNull

	return n
}

func (n *Node) MutateToString(value string) *Node {
	if n == nil || n.bits&hellBitField == hellBitField {
		return nil
	}

	n.bits = hellBitString
	n.data = value

	return n
}

func (n *Node) MutateToEscapedString(value string) *Node {
	if n == nil || n.bits&hellBitField == hellBitField {
		return nil
	}

	n.bits = hellBitEscapedString
	n.data = value

	return n
}

func (n *Node) MutateToObject() *Node {
	if n == nil || n.bits&hellBitField == hellBitField {
		return n
	}

	n.bits = hellBitObject
	n.nodes = n.nodes[:0]

	return n
}

func (n *Node) MutateToStrict() *StrictNode {
	return &StrictNode{n}
}

func (n *Node) DigField(path ...string) *Node {
	if n == nil || len(path) == 0 {
		return nil
	}

	node := n.Dig(path...)
	if node == nil {
		return nil
	}

	return n.nodes[node.getIndex()]
}

func (n *Node) AsFields() []*Node {
	if n == nil {
		return make([]*Node, 0, 0)
	}

	if n.bits&hellBitObject != hellBitObject {
		return (n.nodes)[:0]
	}

	for _, node := range n.nodes {
		if node.bits&hellBitEscapedField == hellBitEscapedField {
			node.unescapeField()
		}
	}

	return n.nodes
}

func (n *StrictNode) AsFields() ([]*Node, error) {
	if n.bits&hellBitObject != hellBitObject {
		return nil, ErrNotObject
	}

	for _, node := range n.nodes {
		if node.bits&hellBitEscapedField == hellBitEscapedField {
			node.unescapeField()
		}
	}

	return n.nodes, nil
}

func (n *Node) AsFieldValue() *Node {
	if n == nil || n.bits&hellBitField != hellBitField {
		return nil
	}

	return n.next
}

func (n *StrictNode) AsFieldValue() (*Node, error) {
	if n == nil || n.bits&hellBitField != hellBitField {
		return nil, ErrNotField
	}

	return n.next, nil
}

func (n *Node) AsArray() []*Node {
	if n == nil {
		return make([]*Node, 0, 0)
	}
	if n.bits&hellBitArray != hellBitArray {
		return (n.nodes)[:0]
	}

	return n.nodes
}

func (n *StrictNode) AsArray() ([]*Node, error) {
	if n == nil || n.bits&hellBitArray != hellBitArray {
		return nil, ErrNotArray
	}

	return n.nodes, nil
}

func (n *Node) unescapeStr() {
	value := n.data
	n.data = unescapeStr(value[1 : len(value)-1])
	n.bits = hellBitString
}

func (n *Node) unescapeField() {
	if n.bits&hellBitField == hellBitField {
		return
	}

	value := n.data
	i := strings.LastIndexByte(value, '"')
	n.data = unescapeStr(value[1:i])
	n.bits = hellBitField
}

func (n *Node) AsString() string {
	if n == nil {
		return ""
	}

	switch n.bits & hellBitTypeFilter {
	case hellBitString:
		return n.data
	case hellBitEscapedString:
		n.unescapeStr()
		return n.data
	case hellBitNumber:
		return n.data
	case hellBitTrue:
		return "true"
	case hellBitFalse:
		return "false"
	case hellBitNull:
		return "null"
	case hellBitField:
		return n.data
	case hellBitEscapedField:
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
	if n.bits&hellBitEscapedField == hellBitEscapedField {
		panic("insane json really goes outta its mind")
	}

	if n.bits&hellBitEscapedString == hellBitEscapedString {
		n.unescapeStr()
	}

	if n == nil || n.bits&hellBitString != hellBitString {
		return "", ErrNotString
	}

	return n.data, nil
}

func (n *Node) AsEscapedString() string {
	if n == nil {
		return ""
	}

	switch n.bits & hellBitTypeFilter {
	case hellBitString:
		return toString(escapeString(make([]byte, 0, 0), n.data))
	case hellBitEscapedString:
		return n.data
	case hellBitNumber:
		return n.data
	case hellBitTrue:
		return "true"
	case hellBitFalse:
		return "false"
	case hellBitNull:
		return "null"
	case hellBitField:
		return n.data
	case hellBitEscapedField:
		panic("insane json really goes outta its mind")
	default:
		return ""
	}
}

func (n *StrictNode) AsEscapedString() (string, error) {
	if n.bits&hellBitEscapedField == hellBitEscapedField {
		panic("insane json really goes outta its mind")
	}

	if n == nil || n.bits&hellBitString != hellBitString {
		return "", ErrNotString
	}

	if n.bits&hellBitEscapedString == hellBitEscapedString {
		return n.data, nil
	}

	return toString(escapeString(make([]byte, 0, 0), n.data)), nil
}

func (n *Node) AsBool() bool {
	if n == nil {
		return false
	}

	switch n.bits & hellBitTypeFilter {
	case hellBitString:
		return n.data == "true"
	case hellBitEscapedString:
		n.unescapeStr()
		return n.data == "true"
	case hellBitNumber:
		return n.data != "0"
	case hellBitTrue:
		return true
	case hellBitFalse:
		return false
	case hellBitNull:
		return false
	case hellBitField:
		return n.data == "true"
	case hellBitEscapedField:
		panic("insane json really goes outta its mind")
	default:
		return false
	}
}

func (n *StrictNode) AsBool() (bool, error) {
	if n == nil || (n.bits&hellBitTrue != hellBitTrue && n.bits&hellBitTrue != hellBitFalse) {
		return false, ErrNotBool
	}

	return n.bits&hellBitTrue == hellBitTrue, nil
}

func (n *Node) AsInt() int {
	if n == nil {
		return 0
	}

	switch n.bits & hellBitTypeFilter {
	case hellBitString:
		return int(math.Round(decodeFloat64(n.data)))
	case hellBitEscapedString:
		n.unescapeStr()
		return int(math.Round(decodeFloat64(n.data)))
	case hellBitNumber:
		return int(math.Round(decodeFloat64(n.data)))
	case hellBitTrue:
		return 1
	case hellBitFalse:
		return 0
	case hellBitNull:
		return 0
	case hellBitField:
		return int(math.Round(decodeFloat64(n.data)))
	case hellBitEscapedField:
		panic("insane json really goes outta its mind")
	default:
		return 0
	}
}

func (n *StrictNode) AsInt() (int, error) {
	if n == nil || n.bits&hellBitNumber != hellBitNumber {
		return 0, ErrNotNumber
	}
	num := decodeInt64(n.data)
	if num == 0 && n.data != "0" {
		return 0, ErrNotNumber
	}
	return int(num), nil
}

func (n *Node) AsFloat() float64 {
	switch n.bits & hellBitTypeFilter {
	case hellBitString:
		return decodeFloat64(n.data)
	case hellBitEscapedString:
		n.unescapeStr()
		return decodeFloat64(n.data)
	case hellBitNumber:
		return decodeFloat64(n.data)
	case hellBitTrue:
		return 1
	case hellBitFalse:
		return 0
	case hellBitNull:
		return 0
	case hellBitField:
		return decodeFloat64(n.data)
	case hellBitEscapedField:
		panic("insane json really goes outta its mind")
	default:
		return 0
	}
}

func (n *StrictNode) AsFloat() (float64, error) {
	if n == nil || n.bits&hellBitNumber != hellBitNumber {
		return 0, ErrNotNumber
	}

	return decodeFloat64(n.data), nil
}

func (n *Node) IsObject() bool {
	return n != nil && n.bits&hellBitObject == hellBitObject
}

func (n *Node) IsArray() bool {
	return n != nil && n.bits&hellBitArray == hellBitArray
}

func (n *Node) IsNumber() bool {
	return n != nil && n.bits&hellBitNumber == hellBitNumber
}

func (n *Node) IsString() bool {
	return n != nil && (n.bits&hellBitString == hellBitString || n.bits&hellBitEscapedString == hellBitEscapedString)
}

func (n *Node) IsTrue() bool {
	return n != nil && n.bits&hellBitTrue == hellBitTrue
}

func (n *Node) IsFalse() bool {
	return n != nil && n.bits&hellBitFalse == hellBitFalse
}

func (n *Node) IsNull() bool {
	return n != nil && n.bits&hellBitNull == hellBitNull
}

func (n *Node) IsField() bool {
	return n != nil && n.bits&hellBitField == hellBitField
}

func (n *Node) IsNil() bool {
	return n == nil
}

func (n *Node) TypeStr() string {
	if n == nil {
		return "nil"
	}

	switch n.bits & hellBitTypeFilter {
	case hellBitObject:
		return "hellBitObject"
	case hellBitEnd:
		return "hellBitObject end"
	case hellBitArray:
		return "hellBitArray"
	case hellBitArrayEnd:
		return "hellBitArray end"
	case hellBitField:
		return "field"
	case hellBitEscapedField:
		return "field escaped"
	case hellBitNull:
		return "null"
	case hellBitString:
		return "string"
	case hellBitEscapedString:
		return "string escaped"
	case hellBitNumber:
		return "number"
	case hellBitTrue:
		return "true"
	case hellBitFalse:
		return "false"
	default:
		return "unknown"
	}
}

// todo how can we use root's decoder here? to avoid allocs?
func (n *Node) getNode() *Node {
	return &Node{}
}

func (n *Node) setIndex(index int) {
	n.bits = (n.bits & hellBitsIndexReset) + hellBits(index)*hellBitsIndexStep
}

func (n *Node) getIndex() int {
	return int((n.bits & hellBitsIndexFilter) / hellBitsIndexStep)
}

// ******************** //
//       DECODER        //
// ******************** //

func (d *decoder) initPool() {
	d.nodePool = make([]*Node, StartNodePoolSize, StartNodePoolSize)
	for i := 0; i < StartNodePoolSize; i++ {
		d.nodePool[i] = &Node{}
	}
}

func (d *decoder) expandPool() []*Node {
	c := cap(d.nodePool)
	for i := 0; i < c; i++ {
		d.nodePool = append(d.nodePool, &Node{})
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
	return Spawn().decoder.decodeHeadless(toString(jsonBytes), true)
}

func DecodeString(json string) (*Root, error) {
	return Spawn().decoder.decodeHeadless(json, true)
}

// DecodeBytes clears Root and decodes new JSON. Useful for reusing Root to reduce allocations.
func (r *Root) DecodeBytes(jsonBytes []byte) error {
	if r == nil {
		return ErrRootIsNil
	}
	_, err := r.decoder.decodeHeadless(toString(jsonBytes), false)

	return err
}

// DecodeString clears Root and decodes new JSON. Useful for reusing Root to reduce allocations.
func (r *Root) DecodeString(json string) error {
	if r == nil {
		return ErrRootIsNil
	}
	_, err := r.decoder.decodeHeadless(json, false)

	return err
}

// DecodeBytesAdditional doesn't clean Root, uses Root node pool to decode JSON
func (r *Root) DecodeBytesAdditional(jsonBytes []byte) (*Node, error) {
	if r == nil {
		return nil, ErrRootIsNil
	}

	return r.decoder.decode(toString(jsonBytes), false)
}

// DecodeStringAdditional doesn't clean Root, uses Root node pool to decode JSON
func (r *Root) DecodeStringAdditional(json string) (*Node, error) {
	if r == nil {
		return nil, ErrRootIsNil
	}

	return r.decoder.decode(json, false)
}

func Release(root *Root) {
	if root == nil {
		return
	}

	backToPool(root.decoder)
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
	if !shouldEscape(st) {
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

func shouldEscape(s string) bool {
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

var fuzzRoot *Root = nil

func Fuzz(data []byte) int {
	if fuzzRoot == nil {
		fuzzRoot = Spawn()
	}
	err := fuzzRoot.DecodeBytes(data)
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

	node := fuzzRoot.Node

	for i := 0; i < 1; i++ {
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

			node.AddField(fields[rand.Int()%len(fields)]).MutateToJSON(fuzzRoot, jsons[rand.Int()%len(jsons)])
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
			node = fuzzRoot.Node
			break
		}
	}

	fuzzRoot.EncodeToString()

	return 1
}

type jsonCopy string

func cp(s string) jsonCopy {
	return jsonCopy(s)
}
func insaneErr(err error, json jsonCopy, offset int) error {
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
		str = strings.ReplaceAll(string(json[a:b]), "\n", " ")
	}
	str = strings.ReplaceAll(str, "\r", " ")
	str = strings.ReplaceAll(str, "\t", " ")

	return errors.New(fmt.Sprintf("%s near `%s`\n%s", err.Error(), str, pointer))
}
