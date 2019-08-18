package insaneJSON

import (
	"fmt"
	"github.com/valyala/fastjson/fastfloat"
	"io/ioutil"
	"math/rand"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fastjson"
)

type test struct {
	json []byte
	name string

	digFields []string
}

func loadTest(name string, getFields []string) *test {
	content, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", name))
	if err != nil {
		panic(err.Error())
	}

	return &test{json: content, name: name, digFields: getFields}
}

func getBenchmarks() []*test {
	tests := make([]*test, 0, 0)
	tests = append(tests, loadTest("light-ws", []string{}))
	tests = append(tests, loadTest("many-objects", []string{}))
	tests = append(tests, loadTest("heavy", []string{}))
	tests = append(tests, loadTest("insane", []string{}))
	return tests
}

func TestDecodeLight(t *testing.T) {
	test := loadTest("light", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")
	assert.Equal(t, Object, node.Type, "Wrong first node")
}

func TestDecodeManyObjects(t *testing.T) {
	test := loadTest("many-objects", []string{})

	node, err := DecodeBytes(test.json)
	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")
	assert.Equal(t, Object, node.Type, "Wrong first node")
	assert.Equal(t, Field, node.AsFields()[0].Type, "Wrong second node")
}

func TestDecodeArray(t *testing.T) {
	test := loadTest("array", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	array := node.Dig("first").AsArray()
	assert.NotNil(t, array, "Array is empty")
	assert.Equal(t, 3, len(array), "Array has wrong length")

	assert.Equal(t, "s1", array[0].AsString(), "Wrong value")
	assert.Equal(t, "s2", array[1].AsString(), "Wrong value")
	assert.Equal(t, "s3", array[2].AsString(), "Wrong value")

	array = node.Dig("second").AsArray()
	assert.NotNil(t, array, "Array is empty")
	assert.NotNil(t, 2, len(array), "Array has wrong length")
	arrayNode := array[0]
	assert.Equal(t, Object, arrayNode.Type, "Wrong value")
	assert.Equal(t, true, arrayNode.Dig("s4").AsBool(), "Wrong value")

	arrayNode = array[1]
	assert.Equal(t, Object, arrayNode.Type, "Wrong value")
	assert.Equal(t, false, arrayNode.Dig("s5").AsBool(), "Wrong value")
}

func TestDecodeArrayFields(t *testing.T) {
	test := loadTest("array-values", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "first", node.AsArray()[0].AsString(), "Wrong array value")
	assert.Equal(t, "second", node.AsArray()[1].AsString(), "Wrong array value")
	assert.Equal(t, "third", node.AsArray()[2].AsString(), "Wrong array value")
}

func TestDecodeTrueFalseNull(t *testing.T) {
	test := loadTest("true-false-null", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, true, node.Dig("true").AsBool(), "Wrong value")
	assert.Equal(t, false, node.Dig("false").AsBool(), "Wrong value")
	assert.Equal(t, true, node.Dig("null").IsNull(), "Wrong value")
}

func TestDecodeNumber(t *testing.T) {
	test := loadTest("number", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, 100, node.Dig("first").AsInt(), "Wrong value")
	assert.Equal(t, 1e20, node.Dig("second").AsFloat(), "Wrong value")
}

func TestDecodeErr(t *testing.T) {
	tests := []struct {
		json string
		err  error
	}{
		// common
		{json: ``, err: ErrEmptyJSON},
		{json: `"`, err: ErrUnexpectedEndOfString},
		{json: `l`, err: ErrExpectedValue},
		{json: `t`, err: ErrUnexpectedEndOfTrue},
		{json: `f`, err: ErrUnexpectedEndOfFalse},
		{json: `n`, err: ErrUnexpectedEndOfNull},

		// array
		{json: `[`, err: ErrExpectedValue},
		{json: `[ `, err: ErrExpectedValue},
		{json: `[ "`, err: ErrUnexpectedEndOfString},
		{json: `[ t`, err: ErrUnexpectedEndOfTrue},
		{json: `[ f`, err: ErrUnexpectedEndOfFalse},
		{json: `[ n`, err: ErrUnexpectedEndOfNull},
		{json: `[[0`, err: ErrUnexpectedJSONEnding},
		{json: `[,`, err: ErrExpectedValue},
		{json: `[e[]00]`, err: ErrExpectedComma},
		{json: `[,e[]00]`, err: ErrExpectedValue},

		// object
		{json: ` {`, err: ErrExpectedObjectField},
		{json: `{`, err: ErrExpectedObjectField},
		{json: `{ `, err: ErrExpectedObjectField},
		{json: `{ f`, err: ErrExpectedObjectField},
		{json: `{{`, err: ErrExpectedObjectField},
		{json: `{"`, err: ErrUnexpectedEndOfObjectField},
		{json: `{l`, err: ErrExpectedObjectField},
		{json: `{ l`, err: ErrExpectedObjectField},
		{json: `{""`, err: ErrExpectedObjectFieldSeparator},
		{json: `{""  `, err: ErrExpectedObjectFieldSeparator},
		{json: `{"":`, err: ErrExpectedValue},
		{json: `{"": `, err: ErrExpectedValue},
		{json: `{"" :`, err: ErrExpectedValue},
		{json: `{"":"`, err: ErrUnexpectedEndOfString},
		{json: `{"": "`, err: ErrUnexpectedEndOfString},
		{json: `{"":0`, err: ErrUnexpectedJSONEnding},
		{json: `{,`, err: ErrExpectedObjectField},
		{json: `{"":0[]00[]00]`, err: ErrExpectedComma},
		{json: `{"":0""[]00[]00]`, err: ErrExpectedComma},
		{json: `{,"":0""[]00[]00]`, err: ErrExpectedObjectField},

		// endings
		{json: `1.0jjj`, err: ErrUnexpectedJSONEnding},
		{json: `{}}`, err: ErrUnexpectedJSONEnding},
		{json: `[].`, err: ErrUnexpectedJSONEnding},
		{json: `"sssss".`, err: ErrUnexpectedJSONEnding},
		{json: `truetrue.`, err: ErrUnexpectedJSONEnding},
		{json: `falsenull`, err: ErrUnexpectedJSONEnding},
		{json: `null:`, err: ErrUnexpectedJSONEnding},


		// ok
		{json: `0`, err: nil},
		{json: `1.0`, err: nil},
		{json: `"string"`, err: nil},
		{json: `true`, err: nil},
		{json: `false`, err: nil},
		{json: `null`, err: nil},
		{json: `{}`, err: nil},
		{json: `[]`, err: nil},
		{json: `[[	],0]`, err: nil},
		{json: `{"":{"l":[30]},"c":""}`, err: nil},
		{json: `{"a":{"6":"5","l":[3,4]},"c":"d"}`, err: nil},
	}

	for _, test := range tests {
		root, err := DecodeString(test.json)
		if test.err != nil {
			assert.NotNil(t, err, "Where should be an error decoding %s", test.json)
		} else {
			assert.Nil(t, err, "Where shouldn't be an error %s", test.json)
			root.Encode()
		}
		assert.Equal(t, test.err, err, "Wrong err %s", test.json)
		Release(root)
	}
}

func TestEncode(t *testing.T) {
	test := loadTest("tricky", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	json := make([]byte, 0, 500000)
	json = node.EncodeNoAlloc(json)
	assert.Equal(t, string(test.json), string(json), "Wrong encoding")
}

func TestString(t *testing.T) {
	test := loadTest("string", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, `hello \ " op \ " op op`, node.Dig("0").AsString(), "Wrong value")
	assert.Equal(t, "shit", node.Dig("1").AsString(), "Wrong value")

	json := make([]byte, 0, 500000)
	json = node.EncodeNoAlloc(json)
	assert.Equal(t, string(test.json), string(json), "Wrong encoding")
}

func TestField(t *testing.T) {
	test := loadTest("field", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "shit", node.Dig(`hello \ " op \ " op op`).AsString(), "Wrong value")

	json := make([]byte, 0, 500000)
	json = node.EncodeNoAlloc(json)
	assert.Equal(t, string(test.json), string(json), "Wrong encoding")
}

func TestInsane(t *testing.T) {
	test := loadTest("insane", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	json := make([]byte, 0, 500000)
	json = node.EncodeNoAlloc(json)
	assert.Equal(t, 465158, len(json), "Wrong encoding")
}

func TestDig(t *testing.T) {
	test := loadTest("many-fields", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "ac784884-a6a3-4987-bb0f-d19e09462677", node.Dig("guid").AsString(), "Wrong encoding")
}

func TestDigEmpty(t *testing.T) {
	node, err := DecodeString(`{"":""}`)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "", node.Dig("").AsString(), "wrong value")
}

func TestDigDeep(t *testing.T) {
	test := loadTest("heavy", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	value := node.Dig("first", "second", "third", "fourth", "fifth", "ok")
	assert.NotNil(t, value, "Can't find field")
	assert.Equal(t, "ok", value.AsString(), "Wrong encoding")
}

func TestDigNil(t *testing.T) {
	test := loadTest("array-values", []string{})

	node, _ := DecodeBytes(test.json)
	defer Release(node)

	value := node.Dig("one", "$f_8)9").AsString()
	assert.Equal(t, "", value, "Wrong array value")
}

func TestAddField(t *testing.T) {
	tests := []struct {
		json   string
		fields []string
		result string
	}{
		{json: `{}`, fields: []string{"a"}, result: `{"a":null}`},
		{json: `{}`, fields: []string{"a", "b", "c"}, result: `{"a":null,"b":null,"c":null}`},
		{json: `{}`, fields: []string{"a", "b", "c", "d"}, result: `{"a":null,"b":null,"c":null,"d":null}`},
		{json: `{"a":"a"}`, fields: []string{"b", "c", "d"}, result: `{"a":"a","b":null,"c":null,"d":null}`},
		{json: `{"x":{"a":"a","e":"e"}}`, fields: []string{"b", "c", "d"}, result: `{"x":{"a":"a","e":"e"},"b":null,"c":null,"d":null}`},
		{json: `{"x":["a","a"]}`, fields: []string{"b", "c", "d"}, result: `{"x":["a","a"],"b":null,"c":null,"d":null}`},
	}

	for _, test := range tests {
		node, err := DecodeString(test.json)
		assert.Nil(t, err, "Where shouldn't be an error")
		for _, field := range test.fields {
			node.AddField(field)
			assert.Equal(t, Null, node.Dig(field).Type, "Wrong type")
		}
		assert.Equal(t, test.result, string(node.Encode()), "Wrong json")
		Release(node)
	}
}

func TestAppendElement(t *testing.T) {
	tests := []struct {
		json   string
		count  int
		result string
	}{
		{json: `[]`, count: 1, result: `[null]`},
		{json: `[]`, count: 3, result: `[null,null,null]`},
		{json: `[]`, count: 4, result: `[null,null,null,null]`},
		{json: `["a"]`, count: 3, result: `["a",null,null,null]`},
		{json: `["a","a"]`, count: 3, result: `["a","a",null,null,null]`},
		{json: `[{"a":"a"}]`, count: 3, result: `[{"a":"a"},null,null,null]`},
		{json: `[["a","a"]]`, count: 3, result: `[["a","a"],null,null,null]`},
	}

	for _, test := range tests {
		node, err := DecodeString(test.json)
		assert.Nil(t, err, "Where shouldn't be an error")
		for index := 0; index < test.count; index++ {
			node.AppendElement()
			l := len(node.AsArray())
			assert.Equal(t, Null, node.Dig(strconv.Itoa(l - 1)).Type, "Wrong type")
		}
		assert.Equal(t, test.result, string(node.Encode()), "Wrong json")
		Release(node)
	}
}

func TestInsertElement(t *testing.T) {
	tests := []struct {
		json   string
		pos1   int
		pos2   int
		result string
	}{
		{json: `[]`, pos1: 0, pos2: 0, result: `[null,null]`},
		{json: `[]`, pos1: 0, pos2: 1, result: `[null,null]`},
		{json: `[{"a":"a"},{"a":"a"},{"a":"a"}]`, pos1: 0, pos2: 0, result: `[null,null,{"a":"a"},{"a":"a"},{"a":"a"}]`},
		{json: `[[],[],[]]`, pos1: 0, pos2: 1, result: `[null,null,[],[],[]]`},
		{json: `[[],[],[]]`, pos1: 0, pos2: 2, result: `[null,[],null,[],[]]`},
		{json: `[[],[],[]]`, pos1: 2, pos2: 3, result: `[[],[],null,null,[]]`},
		{json: `[[],[],[]]`, pos1: 2, pos2: 2, result: `[[],[],null,null,[]]`},
		{json: `[[],[],[]]`, pos1: 3, pos2: 3, result: `[[],[],[],null,null]`},
		{json: `[[],[],[]]`, pos1: 3, pos2: 4, result: `[[],[],[],null,null]`},
	}

	for _, test := range tests {
		node, err := DecodeString(test.json)
		assert.Nil(t, err, "Where shouldn't be an error")
		node.InsertElement(test.pos1)
		assert.Equal(t, Null, node.Dig(strconv.Itoa(test.pos1)).Type, "Wrong type")
		node.InsertElement(test.pos2)
		assert.Equal(t, Null, node.Dig(strconv.Itoa(test.pos2)).Type, "Wrong type")

		assert.Equal(t, test.result, string(node.Encode()), "Wrong json")
		Release(node)
	}
}

func TestFuzz(t *testing.T) {
	tests := []string{
		`[]`,
		`[0,1,2,3]`,
		`[{},{},{},{}]`,
		`[{"1":"1"},{"1":"1"},{"1":"1"},{"1":"1"}]`,
		`[["1","1"],["1","1"],["1","1"],["1","1"]]`,
		`[[],0]`,
		`["a",{"6":"5","l":[3,4]},"c","d"]`,
		`{}`,
		`{"a":null}`,
		`{"a":null,"b":null,"c":null}`,
		`{"a":null,"b":null,"c":null,"d":null}`,
		`{"a":"a"}`,
		`{"a":"a","b":null,"c":null,"d":null}`,
		`{"x":{"a":"a","e":"e"}}`,
		`{"x":{"a":"a","e":"e"},"b":null,"c":null,"d":null}`,
		`{"x":["a","a"]}`,
		`{"x":["a","a"],"b":null,"c":null,"d":null}`,
		`[null]`,
		`[null,null,null]`,
		`[null,null,null,null]`,
		`["a",null,null,null]`,
		`["a","a"]`,
		`["a","a",null,null,null]`,
		`[{"a":"a"}]`,
		`[{"a":"a"},null,null,null]`,
		`[["a","a"]]`,
		`[["a","a"],null,null,null]`,
		`[]`,
		`[0,1,2,3]`,
		`[{},{},{},{}]`,
		`[{"1":"1"},{"1":"1"},{"1":"1"},{"1":"1"}]`,
		`[["1","1"],["1","1"],["1","1"],["1","1"]]`,
		`[[],0]`,
		`["a",{"6":"5","l":[3,4]},"c","d"]`,
	}

	rand.Seed(0)
	for _, test := range tests {
		Fuzz([]byte(test))
	}
}

func TestArraySuicide(t *testing.T) {
	tests := []string{
		`[]`,
		`[0,1,2,3]`,
		`[{},{},{},{}]`,
		`[{"1":"1"},{"1":"1"},{"1":"1"},{"1":"1"}]`,
		`[["1","1"],["1","1"],["1","1"],["1","1"]]`,
		`[[],0]`,
		`["a",{"6":"5","l":[3,4]},"c","d"]`,
	}

	for _, json := range tests {
		root, err := DecodeString(json)
		assert.Nil(t, err, "err should be nil")
		for range root.AsArray() {
			root.Dig("0").Suicide()
		}
		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `[]`, string(root.Encode()), "array should be empty")
		Release(root)

		root, err = DecodeString(json)
		assert.Nil(t, err, "err should be nil")
		l := len(root.AsArray())
		for i := range root.AsArray() {
			root.Dig(strconv.Itoa(l - i - 1)).Suicide()
		}

		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `[]`, string(root.Encode()), "array should be empty")
		Release(root)
	}
}

func TestObjectSuicide(t *testing.T) {
	tests := []string{
		`{}`,
		`{"0":"0","1":"1","2":"2","3":"3"}`,
		`{"1":{},"2":{},"3":{},"4":{}}`,
		`{"1":{"1":"1"},"2":{"1":"1"},"3":{"1":"1"},"4":{"1":"1"}}`,
		`{"1":["1","1"],"2":["1","1"],"3":["1","1"],"4":["1","1"]}`,
		`{"1":[],"2":0}`,
		`{"a":{"6":"5","l":[3,4]},"c":"d"}`,
		`{"":{"":"","":""}}`,
		`{"x":{"a":"a","e":"e"}}`,
	}

	for _, json := range tests {
		root, err := DecodeString(json)
		assert.Nil(t, err, "err should be nil")
		fields := root.AsFields()
		for _, field := range fields {
			root.Dig(field.AsString()).Suicide()
		}
		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `{}`, string(root.Encode()), "array should be empty")
		Release(root)

		root, err = DecodeString(json)
		assert.Nil(t, err, "err should be nil")
		fields = root.AsFields()
		l := len(fields)
		for i := range fields {
			root.Dig(fields[l-i-1].AsString()).Suicide()
		}
		for _, field := range fields {
			root.Dig(field.AsString()).Suicide()
		}
		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `{}`, string(root.Encode()), "array should be empty")
		Release(root)
	}
}

func TestMutateToJSON(t *testing.T) {
	tests := []struct {
		name        string
		sourceJSON  string
		mutateJSON  string
		mutateDig   []string
		resultJSON  string
		checkDig    [][]string
		checkValues []int
	}{
		{
			sourceJSON:  `{"a":"b","c":"d"}`,
			mutateJSON:  `{"5":"5","l":[3,4]}`,
			mutateDig:   []string{"a"},
			resultJSON:  `{"a":{"5":"5","l":[3,4]},"c":"d"}`,
			checkDig:    [][]string{{"a", "l", "0"}, {"a", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			sourceJSON:  `{"a":"b","c":"d"}`,
			mutateJSON:  `{"5":"5","l":[3,4]}`,
			mutateDig:   []string{"c"},
			resultJSON:  `{"a":"b","c":{"5":"5","l":[3,4]}}`,
			checkDig:    [][]string{{"c", "l", "0"}, {"c", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			sourceJSON:  `{"a":{"somekey":"someval", "xxx":"yyy"},"c":"d"}`,
			mutateJSON:  `{"5":"5","l":[3,4]}`,
			mutateDig:   []string{"a"},
			resultJSON:  `{"a":{"5":"5","l":[3,4]},"c":"d"}`,
			checkDig:    [][]string{{"a", "l", "0"}, {"a", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			sourceJSON:  `["a","b","c","d"]`,
			mutateJSON:  `{"5":"5","l":[3,4]}`,
			mutateDig:   []string{"0"},
			resultJSON:  `[{"5":"5","l":[3,4]},"b","c","d"]`,
			checkDig:    [][]string{{"0", "l", "0"}, {"0", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			sourceJSON:  `["a","b","c","d"]`,
			mutateJSON:  `{"5":"5","l":[3,4]}`,
			mutateDig:   []string{"3"},
			resultJSON:  `["a","b","c",{"5":"5","l":[3,4]}]`,
			checkDig:    [][]string{{"3", "l", "0"}, {"3", "l", "1"}},
			checkValues: []int{3, 4},
		},
	}
	for _, test := range tests {
		node, err := DecodeBytes([]byte(test.sourceJSON))
		assert.Nil(t, err, "Error while decoding")

		mutatingNode := node.Dig(test.mutateDig...)
		mutatingNode.MutateToJSON(test.mutateJSON)
		for i := range test.checkDig {
			assert.Equal(t, test.checkValues[i], node.Dig(test.checkDig[i]...).AsInt(), "Wrong value")
			assert.Equal(t, test.checkValues[i], mutatingNode.Dig(test.checkDig[i][1:]...).AsInt(), "Wrong value")
		}

		encoded := node.Encode()
		assert.Equal(t, test.resultJSON, string(encoded), "Wrong result json")
	}
}

func TestMutateToObject(t *testing.T) {
	tests := []struct {
		sourceJSON string
		mutateDig  []string
		resultJSON string
	}{
		{
			sourceJSON: `{"a":"b","c":"d"}`,
			mutateDig:  []string{"a"},
			resultJSON: `{"a":{},"c":"d"}`,
		},
		{
			sourceJSON: `{"a":"b","c":"d"}`,
			mutateDig:  []string{"c"},
			resultJSON: `{"a":"b","c":{}}`,
		},
		{
			sourceJSON: `["a","b","c","d"]`,
			mutateDig:  []string{"1"},
			resultJSON: `["a",{},"c","d"]`,
		},
		{
			sourceJSON: `["a","b","c","d"]`,
			mutateDig:  []string{"0"},
			resultJSON: `[{},"b","c","d"]`,
		},
		{
			sourceJSON: `["a","b","c","d"]`,
			mutateDig:  []string{"3"},
			resultJSON: `["a","b","c",{}]`,
		},
		{
			sourceJSON: `"string"`,
			mutateDig:  []string{},
			resultJSON: `{}`,
		},
		{
			sourceJSON: `[]`,
			mutateDig:  []string{},
			resultJSON: `{}`,
		},
		{
			sourceJSON: `true`,
			mutateDig:  []string{},
			resultJSON: `{}`,
		},
		{
			sourceJSON: `{"a":{"a":{"a":{"a":"b","c":"d"},"c":"d"},"c":"d"},"c":"d"}`,
			mutateDig:  []string{"a", "a", "a", "a"},
			resultJSON: `{"a":{"a":{"a":{"a":{},"c":"d"},"c":"d"},"c":"d"},"c":"d"}`,
		},
	}
	for _, test := range tests {
		node, err := DecodeBytes([]byte(test.sourceJSON))
		assert.Nil(t, err, "Error while decoding")

		mutatingNode := node.Dig(test.mutateDig...).MutateToObject()
		assert.Equal(t, Object, mutatingNode.Type, "Wrong type")
		assert.Equal(t, Object, node.Dig(test.mutateDig...).Type, "Wrong type")

		encoded := node.Encode()
		assert.Equal(t, test.resultJSON, string(encoded), "Wrong result json")
	}
}

func TestMutateCollapse(t *testing.T) {
	tests := []struct {
		name        string
		sourceJSON  string
		mutateValue int
		mutateDig   []string
		resultJSON  string
		checkDig    [][]string
		checkValues []int
	}{
		{
			sourceJSON:  `{"a":"b","b":["k","k","l","l"],"m":"m"}`,
			mutateValue: 15,
			mutateDig:   []string{"b"},
			resultJSON:  `{"a":"b","b":15,"m":"m"}`,
			checkDig:    [][]string{{"b"}},
			checkValues: []int{15},
		},
		{
			sourceJSON:  `{"a":"b","b":{"k":"k","l":"l"},"m":"m"}`,
			mutateValue: 15,
			mutateDig:   []string{"b"},
			resultJSON:  `{"a":"b","b":15,"m":"m"}`,
			checkDig:    [][]string{{"b"}},
			checkValues: []int{15},
		},
	}

	for _, test := range tests {
		node, err := DecodeBytes([]byte(test.sourceJSON))
		assert.Nil(t, err, "Error while decoding")

		mutatingNode := node.Dig(test.mutateDig...)
		mutatingNode.MutateToInt(test.mutateValue)
		for i := range test.checkDig {
			assert.Equal(t, test.checkValues[i], node.Dig(test.checkDig[i]...).AsInt(), "Wrong value")
			assert.Equal(t, test.checkValues[i], mutatingNode.Dig(test.checkDig[i][1:]...).AsInt(), "Wrong value")
		}

		encoded := node.Encode()
		assert.Equal(t, test.resultJSON, string(encoded), "Wrong result json")
	}
}

func TestMutateToInt(t *testing.T) {
	node, err := DecodeBytes([]byte(`{"a":"b"}`))
	assert.Nil(t, err, "Error while decoding")

	node.Dig("a").MutateToInt(5)
	assert.Equal(t, 5, node.Dig("a").AsInt(), "Wrong value")

	encoded := node.Encode()
	assert.Equal(t, `{"a":5}`, string(encoded), "Wrong result json")
}

func TestMutateToFloat(t *testing.T) {
	node, err := DecodeBytes([]byte(`{"a":"b"}`))
	assert.Nil(t, err, "Error while decoding")

	node.Dig("a").MutateToFloat(5.6)
	assert.Equal(t, 5.6, node.Dig("a").AsFloat(), "Wrong value")
	assert.Equal(t, 6, node.Dig("a").AsInt(), "Wrong value")

	encoded := node.Encode()
	assert.Equal(t, `{"a":5.6}`, string(encoded), "Wrong result json")
}

func TestMutateToString(t *testing.T) {
	node, err := DecodeBytes([]byte(`{"a":"b"}`))
	assert.Nil(t, err, "Error while decoding")

	node.Dig("a").MutateToString("insane")
	assert.Equal(t, "insane", node.Dig("a").AsString(), "Wrong value")

	encoded := node.Encode()
	assert.Equal(t, `{"a":"insane"}`, string(encoded), "Wrong result json")
}

func TestMutateToField(t *testing.T) {
	node, err := DecodeBytes([]byte(`{"a":"b"}`))
	assert.Nil(t, err, "Error while decoding")

	node.AsFields()[0].MutateToField("insane")
	assert.Equal(t, "b", node.Dig("insane").AsString(), "Wrong value")

	encoded := node.Encode()
	assert.Equal(t, `{"insane":"b"}`, string(encoded), "Wrong result json")
}

func TestAsField(t *testing.T) {
	node, err := DecodeBytes([]byte(`{"a":"b"}`))
	assert.Nil(t, err, "Error while decoding")

	node.AsField("a").MutateToField("insane")
	assert.Equal(t, "b", node.Dig("insane").AsString(), "Wrong value")

	encoded := node.Encode()
	assert.Equal(t, `{"insane":"b"}`, string(encoded), "Wrong result json")
}

func TestWhitespace(t *testing.T) {
	test := loadTest("whitespace", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "one", node.Dig("one").AsString(), "Wrong field name")
	assert.Equal(t, "two", node.Dig("two").AsString(), "Wrong field name")
}

func TestObjectManyFieldsSuicide(t *testing.T) {
	test := loadTest("many-fields", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	fields := make([]string, 0, 0)
	for _, field := range node.AsFields() {
		fields = append(fields, string([]byte(field.AsString())))
	}

	for _, field := range node.AsFields() {
		node.Dig(field.AsString()).Suicide()
	}

	for _, field := range fields {
		assert.Nil(t, node.Dig(field), "node should'n be findable")
	}

	assert.Equal(t, `{}`, string(node.Encode()), "Wrong result json")
}

func TestObjectManyFieldsAddSuicide(t *testing.T) {
	node, err := DecodeString("{}")
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	fields := 30
	for i := 0; i < fields; i++ {
		node.AddField(strconv.Itoa(i)).MutateToString(strconv.Itoa(i))
	}

	for i := 0; i < fields; i++ {
		assert.NotNil(t, node.Dig(strconv.Itoa(i)), "node should be findable")
		assert.Equal(t, strconv.Itoa(i), node.Dig(strconv.Itoa(i)).AsString(), "wrong value")
	}

	for i := 0; i < fields; i++ {
		node.Dig(strconv.Itoa(i)).Suicide()
	}

	for i := 0; i < fields; i++ {
		assert.Nil(t, node.Dig(strconv.Itoa(i), "node should be findable"))
	}

	assert.Equal(t, `{}`, string(node.Encode()), "wrong result json")
}

func TestObjectFields(t *testing.T) {
	test := loadTest("object-fields", []string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, `"first":`, node.data.values[0].value, "Wrong field name")
	assert.Equal(t, `"second":`, node.data.values[1].value, "Wrong field name")
	assert.Equal(t, `"third":`, node.data.values[2].value, "Wrong field name")
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		s string
		n int64
	}{
		{s: "", n: 0},
		{s: " ", n: 0},
		{s: "xxx", n: 0},
		{s: "-xxx", n: 0},
		{s: "1xxx", n: 0},
		{s: "-", n: 0},
		{s: "111 ", n: 0},
		{s: "1-1", n: 0},
		{s: "s1", n: 0},
		{s: "0", n: 0},
		{s: "-0", n: 0},
		{s: "5", n: 5},
		{s: "-5", n: -5},
		{s: " 0", n: 0},
		{s: " 5", n: 0},
		{s: "333", n: 333},
		{s: "-333", n: -333},
		{s: "1111111111", n: 1111111111},
		{s: "987654321", n: 987654321},
		{s: "123456789", n: 123456789},
		{s: "9223372036854775807", n: 9223372036854775807},
		{s: "-9223372036854775807", n: -9223372036854775807},
		{s: "9999999999999999999", n: 0},
		{s: "99999999999999999999", n: 0},
		{s: "-9999999999999999999", n: 0},
		{s: "-99999999999999999999", n: 0},
	}

	for _, test := range tests {
		x := decodeInt64(test.s)

		assert.Equal(t, test.n, x, "wrong number")
	}
}

func BenchmarkDecode(b *testing.B) {
	runtime.SetCPUProfileRate(100000000)
	for _, benchmark := range getBenchmarks() {
		b.Run("insane-"+benchmark.name, func(b *testing.B) {
			root := Spawn()
			b.SetBytes(int64(len(benchmark.json)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = DecodeBytesReusing(root, benchmark.json)
			}
			Release(root)
		})
		b.Run("fast-"+benchmark.name, func(b *testing.B) {
			parser := fastjson.Parser{}
			b.SetBytes(int64(len(benchmark.json)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = parser.ParseBytes(benchmark.json)
			}
		})
	}
}

func BenchmarkEncode(b *testing.B) {
	for _, benchmark := range getBenchmarks() {
		b.Run("insane-"+benchmark.name, func(b *testing.B) {
			root, _ := DecodeBytes(benchmark.json)
			s := make([]byte, 0, 500000)
			b.SetBytes(int64(len(benchmark.json)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = root.EncodeNoAlloc(s[:0])
			}
			Release(root)
		})

		b.Run("fastjson", func(b *testing.B) {
			parser := fastjson.Parser{}
			c, _ := parser.ParseBytes(benchmark.json)
			s := make([]byte, 0, 500000)
			b.SetBytes(int64(len(benchmark.json)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = c.MarshalTo(s[:0])
			}
		})
	}
}

func BenchmarkDig(b *testing.B) {
	tests := make([]*test, 0, 0)
	tests = append(tests, loadTest("light-ws", []string{"about"}))
	tests = append(tests, loadTest("many-objects", []string{"somefield", "somefield", "somefield", "somefield", "somefield", "somefield", "somefield"}))
	tests = append(tests, loadTest("heavy", []string{"first", "second", "third", "fourth", "fifth"}))
	tests = append(tests, loadTest("many-fields", []string{"compfanvy"}))
	tests = append(tests, loadTest("few-fields", []string{"compfanvy"}))

	for _, benchmark := range tests {
		b.Run("insane-"+benchmark.name, func(b *testing.B) {
			root, _ := DecodeBytes(benchmark.json)
			b.SetBytes(1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				root.Dig(benchmark.digFields...)
			}
			Release(root)
		})

		b.Run("fastjson", func(b *testing.B) {
			parser := fastjson.Parser{}
			c, _ := parser.ParseBytes(benchmark.json)
			b.SetBytes(1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.Get(benchmark.digFields...)
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		s string
	}{
		{s: `"`},
		{s: `	`},
		{s: `"""\\\\\"""\'\"				\\\""|"|"|"|\\'\dasd'		|"|\\\\'\\\|||\\'"`},
	}

	out := make([]byte, 0, 0)
	for _, test := range tests {
		out = escapeStringNg(out[:0], test.s)
		assert.Equal(t, string(strconv.AppendQuote(nil, test.s)), string(out), "wrong escaping")
	}

}

func BenchmarkDecodeInt(b *testing.B) {
	tests := []struct {
		s string
		n int64
	}{
		{s: "", n: 0},
		{s: " ", n: 0},
		{s: "xxx", n: 0},
		{s: "-xxx", n: 0},
		{s: "1xxx", n: 0},
		{s: "-", n: 0},
		{s: "111 ", n: 0},
		{s: "1-1", n: 0},
		{s: "s1", n: 0},
		{s: "0", n: 0},
		{s: "-0", n: 0},
		{s: "5", n: 5},
		{s: "-5", n: -5},
		{s: " 0", n: 0},
		{s: " 5", n: 0},
		{s: "333", n: 333},
		{s: "-333", n: -333},
		{s: "1111111111", n: 1111111111},
		{s: "987654321", n: 987654321},
		{s: "123456789", n: 123456789},
		{s: "9223372036854775807", n: 9223372036854775807},
		{s: "-9223372036854775807", n: -9223372036854775807},
		{s: "9999999999999999999", n: 0},
		{s: "99999999999999999999", n: 0},
		{s: "-9999999999999999999", n: 0},
		{s: "-99999999999999999999", n: 0},
	}

	b.Run("insane", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range tests {
				decodeInt64(test.s)
			}
		}
	})

	b.Run("fastjson", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, test := range tests {
				fastfloat.ParseInt64BestEffort(test.s)
			}
		}
	})
}

func BenchmarkEscapeString(b *testing.B) {
	tests := []struct {
		s string
	}{
		{s: `"""\\\\\"""\'\"				\\\""|"|"|"|\\'\dasd'		|"|\\\\'\\\|||\\'"`},
		//{s: `sfsafwefqwueibfiquwbfiuqwebfiuqwbfiquwbfqiwbfoqiwuefboqiweubfoqiwebfiqowebufu`},
	}

	b.Run("insane", func(b *testing.B) {
		out := make([]byte, 0, 0)
		for i := 0; i < b.N; i++ {
			for _, test := range tests {
				out = escapeStringNg(out[:0], test.s)
			}
		}
	})

	b.Run("fastjson", func(b *testing.B) {
		out := make([]byte, 0, 0)
		for i := 0; i < b.N; i++ {
			for _, test := range tests {
				out = escapeString(out[:0], test.s)
			}
		}
	})
}
