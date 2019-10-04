package insaneJSON

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	manyFields = `{"_id":"5d57e438c48c6e5d4ca83b29","index":0,"guid":"ac784884-a6a3-4987-bb0f-d19e09462677","isActive":false,"balance":"$1,961.04","picture":"http://placehold.it/32x32","age":33,"eyeColor":"blue","name":"McleodMendez","gender":"male","compafny":"PARCOE","com1pany":"PARCOE","co2mpfany":"PARCOE","co222fmpany":"PARCOE","co222empany":"PARCOE","co2222fmpany":"PARCOE","co2222mpany":"PARCOE","compfany":"PARCOE","co2mpany":"PARCOE","compwfany":"PARCOE","compfweany":"PARCOE","comfwepany":"PARCOE","compefefany":"PARCOE","comfeqpany":"PARCOE","comfwefvwepany":"PARCOE","comvfqfqewfpany":"PARCOE","compweewany":"PARCOE","wff":"PARCOE","comqvvpany":"PARCOE","comvqwevpany":"PARCOE","compvany":"PARCOE","compvqeany":"PARCOE","compfanvy":"PARCOE","comspany":"PARCOE","compaany":"PARCOE","compaaqny":"PARCOE","compaqny":"PARCOE","_id1":"5d57e438c48c6e5d4ca83b29","index1":0,"guid1":"ac784884-a6a3-4987-bb0f-d19e09462677","isActive1":false,"balance1":"$1,961.04","picture1":"http://placehold.it/32x32","age1":33,"eyeColor1":"blue","name1":"McleodMendez","gender1":"male","compafny1":"PARCOE","com1pany1":"PARCOE","co2mpfany1":"PARCOE","co222fmpany1":"PARCOE","co222empany1":"PARCOE","co2222fmpany1":"PARCOE","co2222mpany1":"PARCOE","compfany1":"PARCOE","co2mpany1":"PARCOE","compwfany1":"PARCOE","compfweany1":"PARCOE","comfwepany1":"PARCOE","compefefany1":"PARCOE","comfeqpany1":"PARCOE","comfwefvwepany1":"PARCOE","comvfqfqewfpany1":"PARCOE","compweewany1":"PARCOE","wff1":"PARCOE","comqvvpany1":"PARCOE","comvqwevpany1":"PARCOE","compvany1":"PARCOE","compvqeany1":"PARCOE","compfanvy1":"PARCOE","comspany1":"PARCOE","compaany1":"PARCOE","compaaqny1":"PARCOE","compaqny1":"PARCOE"}`
)

func TestDecodeLight(t *testing.T) {
	json := `{"_id":"5d53006246df0b962b787d11","index":"0","guid":"80d75945-6251-46a2-b6c9-a10094beed6e","isActive":"false","balance":"$2,258.24","picture":"http://placehold.it/32x32","age":"34","eyeColor":"brown","company":"NIMON","email":"anne.everett@nimon.name","phone":"+1(946)560-2227","address":"815EmpireBoulevard,Blue,Nevada,5617","about":"Proidentoccaecateulaborislaboreofficialaborumvelitanimnulla.Laboreametoccaecataliquaminimlaboreadenimdolorelaborum.Eiusmodesseeiusmodaliquacillumullamcodonisivelitesseincididunt.Ininestessereprehenderitirureaniminsit.","registered":"Friday,May27,20165:05AM","latitude":"-5.922381","longitude":"-49.143968","greeting":"Hello,Anne!Youhave7unreadmessages.","favoriteFruit":"banana"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, root, "Node is nil")
	assert.Equal(t, Object, root.Type, "Wrong first node")
}

func TestDecodeReusing(t *testing.T) {
	json := `{"_id":"5d53006246df0b962b787d11","index":"0","guid":"80d75945-6251-46a2-b6c9-a10094beed6e","isActive":"false","balance":"$2,258.24","picture":"http://placehold.it/32x32","age":"34","eyeColor":"brown","company":"NIMON","email":"anne.everett@nimon.name","phone":"+1(946)560-2227","address":"815EmpireBoulevard,Blue,Nevada,5617","about":"Proidentoccaecateulaborislaboreofficialaborumvelitanimnulla.Laboreametoccaecataliquaminimlaboreadenimdolorelaborum.Eiusmodesseeiusmodaliquacillumullamcodonisivelitesseincididunt.Ininestessereprehenderitirureaniminsit.","registered":"Friday,May27,20165:05AM","latitude":"-5.922381","longitude":"-49.143968","greeting":"Hello,Anne!Youhave7unreadmessages.","favoriteFruit":"banana"}`
	root := Spawn()
	err := root.DecodeString(json)
	defer Release(root)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, root, "Node is nil")
	assert.Equal(t, Object, root.Type, "Wrong first node")
}

func TestDecodeAdditional(t *testing.T) {
	jsonA := `{"_id":"5d53006246df0b962b787d11"}`
	root, err := DecodeString(jsonA)
	assert.NotNil(t, root, "Node is nil")
	assert.Nil(t, err, "Error while decoding")

	jsonB := `[0, 1, 2]`
	node, err := root.DecodeStringAdditional(jsonB)
	assert.NotNil(t, node, "Node is nil")
	assert.Nil(t, err, "Error while decoding")

	assert.Equal(t, jsonA, root.EncodeToString(), "Wrong first node")
	assert.Equal(t, 1, node.Dig("1").AsInt(), "Wrong value")
}

func TestDecodeManyObjects(t *testing.T) {
	json := `{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":"ok"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, root, "Node is nil")
	assert.Equal(t, Object, root.Type, "Wrong first node")
	assert.Equal(t, Field, root.AsFields()[0].Type, "Wrong second node")
}

func TestDecodeArray(t *testing.T) {
	json := `{"first":["s1","s2","s3"],"second":[{"s4":true},{"s5":false}]}`
	node, err := DecodeString(json)
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
	json := `["first","second","third"]`
	root, err := DecodeString(json)
	defer Release(root)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, root, "Node is nil")

	assert.Equal(t, "first", root.AsArray()[0].AsString(), "Wrong array value")
	assert.Equal(t, "second", root.AsArray()[1].AsString(), "Wrong array value")
	assert.Equal(t, "third", root.AsArray()[2].AsString(), "Wrong array value")
}

func TestDecodeTrueFalseNull(t *testing.T) {
	json := `{"true":true,"false":false,"null":null}`
	node, err := DecodeString(json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, true, node.Dig("true").AsBool(), "Wrong value")
	assert.Equal(t, false, node.Dig("false").AsBool(), "Wrong value")
	assert.Equal(t, true, node.Dig("null").IsNull(), "Wrong value")
}

func TestDecodeNumber(t *testing.T) {
	json := `{"first": 100,"second": 1e20}`
	node, err := DecodeString(json)
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
			root.EncodeToByte()
		}
		assert.Equal(t, test.err, err, "Wrong err %s", test.json)
		Release(root)
	}
}

func TestEncode(t *testing.T) {
	json := `{"key_a":{"key_a_a":["v1","vv1"],"key_a_b":[],"key_a_c":"v3"},"key_b":{"key_b_a":["v3","v31"],"key_b_b":{}}}`
	node, err := DecodeString(json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	encoded := node.EncodeToByte()
	assert.Equal(t, json, string(encoded), "Wrong encoding")
}

func TestString(t *testing.T) {
	json := `["hello \\ \" op \\ \" op op","shit"]`

	node, err := DecodeString(json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, `hello \ " op \ " op op`, node.Dig("0").AsString(), "Wrong value")
	assert.Equal(t, "shit", node.Dig("1").AsString(), "Wrong value")

	encoded := node.EncodeToByte()
	assert.Equal(t, json, string(encoded), "Wrong encoding")
}

func TestField(t *testing.T) {
	json := `{"hello \\ \" op \\ \" op op":"shit"}`
	node, err := DecodeString(json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "shit", node.Dig(`hello \ " op \ " op op`).AsString(), "Wrong value")

	encoded := node.EncodeToByte()
	assert.Equal(t, json, string(encoded), "Wrong encoding")
}

func TestInsane(t *testing.T) {
	test := loadJSON("insane", [][]string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	encoded := node.EncodeToByte()
	assert.Equal(t, 465158, len(encoded), "Wrong encoding")
}

func TestDig(t *testing.T) {
	root, err := DecodeString(manyFields)
	defer Release(root)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, root, "Node is nil")

	assert.Equal(t, "ac784884-a6a3-4987-bb0f-d19e09462677", root.Dig("guid").AsString(), "Wrong encoding")
}

func TestDigEmpty(t *testing.T) {
	node, err := DecodeString(`{"":""}`)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "", node.Dig("").AsString(), "wrong value")
}

func TestDigDeep(t *testing.T) {
	test := loadJSON("heavy", [][]string{})

	node, err := DecodeBytes(test.json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	value := node.Dig("first", "second", "third", "fourth", "fifth", "ok")
	assert.NotNil(t, value, "Can't find field")
	assert.Equal(t, "ok", value.AsString(), "Wrong encoding")
}

func TestDigNil(t *testing.T) {
	json := `["first","second","third"]`
	root, _ := DecodeString(json)
	defer Release(root)

	value := root.Dig("one", "$f_8)9").AsString()
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
		assert.Equal(t, test.result, string(node.EncodeToByte()), "Wrong json")
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
		assert.Equal(t, test.result, string(node.EncodeToByte()), "Wrong json")
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

		assert.Equal(t, test.result, string(node.EncodeToByte()), "Wrong json")
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
		assert.Equal(t, `[]`, string(root.EncodeToByte()), "array should be empty")
		Release(root)

		root, err = DecodeString(json)
		assert.Nil(t, err, "err should be nil")
		l := len(root.AsArray())
		for i := range root.AsArray() {
			root.Dig(strconv.Itoa(l - i - 1)).Suicide()
		}

		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `[]`, string(root.EncodeToByte()), "array should be empty")
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
		assert.Equal(t, `{}`, string(root.EncodeToByte()), "array should be empty")
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
		assert.Equal(t, `{}`, string(root.EncodeToByte()), "array should be empty")
		Release(root)
	}
}

func TestMergeWith(t *testing.T) {
	jsonA := `{"1":"1","2":"2"}`
	root, err := DecodeString(jsonA)
	assert.NotNil(t, root, "Node is nil")
	assert.Nil(t, err, "Error while decoding")

	jsonB := `{"1":"1","3":"3","4":"4"}`
	node, err := root.DecodeStringAdditional(jsonB)
	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	root.MergeWith(node)

	assert.Equal(t, `{"1":"1","2":"2","3":"3","4":"4"}`, root.EncodeToString(), "Wrong first node")
}

func TestMergeWithComplex(t *testing.T) {
	jsonA := `{"1":{"1":"1"}}`
	root, err := DecodeString(jsonA)
	assert.NotNil(t, root, "Node is nil")
	assert.Nil(t, err, "Error while decoding")

	jsonB := `{"1":1,"2":{"2":"2"}}`
	node, err := root.DecodeStringAdditional(jsonB)
	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	root.MergeWith(node)

	assert.Equal(t, `{"1":1,"2":{"2":"2"}}`, root.EncodeToString(), "Wrong first node")
}

func TestMutateToJSON(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		mutation    string
		dig         []string
		result      string
		checkDig    [][]string
		checkValues []int
	}{
		{
			source:      `{"a":"b","c":"4"}`,
			dig:         []string{"a"},
			mutation:    `5`,
			result:      `{"a":5,"c":"4"}`,
			checkDig:    [][]string{{"a"}, {"c"}},
			checkValues: []int{5, 4},
		},
		{
			source:      `{"a":"1","c":"4"}`,
			dig:         []string{"c"},
			mutation:    `6`,
			result:      `{"a":"1","c":6}`,
			checkDig:    [][]string{{"a"}, {"c"}},
			checkValues: []int{1, 6},
		},
		{
			source:      `{"a":"b","c":"d"}`,
			dig:         []string{"a"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `{"a":{"5":"5","l":[3,4]},"c":"d"}`,
			checkDig:    [][]string{{"a", "l", "0"}, {"a", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			source:      `{"a":"b","c":"d"}`,
			dig:         []string{"c"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `{"a":"b","c":{"5":"5","l":[3,4]}}`,
			checkDig:    [][]string{{"c", "l", "0"}, {"c", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			source:      `{"a":{"somekey":"someval", "xxx":"yyy"},"c":"d"}`,
			dig:         []string{"a"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `{"a":{"5":"5","l":[3,4]},"c":"d"}`,
			checkDig:    [][]string{{"a", "l", "0"}, {"a", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			source:      `["a","b","c","d"]`,
			dig:         []string{"0"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `[{"5":"5","l":[3,4]},"b","c","d"]`,
			checkDig:    [][]string{{"0", "l", "0"}, {"0", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			source:      `["a","b","c","d"]`,
			dig:         []string{"3"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `["a","b","c",{"5":"5","l":[3,4]}]`,
			checkDig:    [][]string{{"3", "l", "0"}, {"3", "l", "1"}},
			checkValues: []int{3, 4},
		},
	}
	for _, test := range tests {
		node, err := DecodeString(test.source)
		assert.Nil(t, err, "Error while decoding")

		mutatingNode := node.Dig(test.dig...)
		mutatingNode.MutateToJSON(test.mutation)
		for i, dig := range test.checkDig {
			assert.Equal(t, test.checkValues[i], node.Dig(dig...).AsInt(), "Wrong value")
			if len(dig) > 1 {
				assert.Equal(t, test.checkValues[i], mutatingNode.Dig(dig[1:]...).AsInt(), "Wrong value")
			}
		}

		encoded := node.EncodeToByte()
		assert.Equal(t, test.result, string(encoded), "Wrong result json")
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
		node, err := DecodeString(test.sourceJSON)
		assert.Nil(t, err, "Error while decoding")

		mutatingNode := node.Dig(test.mutateDig...).MutateToObject()
		assert.Equal(t, Object, mutatingNode.Type, "Wrong type")
		assert.Equal(t, Object, node.Dig(test.mutateDig...).Type, "Wrong type")

		encoded := node.EncodeToByte()
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
		node, err := DecodeString(test.sourceJSON)
		assert.Nil(t, err, "Error while decoding")

		mutatingNode := node.Dig(test.mutateDig...)
		mutatingNode.MutateToInt(test.mutateValue)
		for i := range test.checkDig {
			assert.Equal(t, test.checkValues[i], node.Dig(test.checkDig[i]...).AsInt(), "Wrong value")
			assert.Equal(t, test.checkValues[i], mutatingNode.Dig(test.checkDig[i][1:]...).AsInt(), "Wrong value")
		}

		encoded := node.EncodeToByte()
		assert.Equal(t, test.resultJSON, string(encoded), "Wrong result json")
	}
}

func TestMutateToInt(t *testing.T) {
	node, err := DecodeString(`{"a":"b"}`)
	assert.Nil(t, err, "Error while decoding")

	node.Dig("a").MutateToInt(5)
	assert.Equal(t, 5, node.Dig("a").AsInt(), "Wrong value")

	encoded := node.EncodeToByte()
	assert.Equal(t, `{"a":5}`, string(encoded), "Wrong result json")
}

func TestMutateToFloat(t *testing.T) {
	node, err := DecodeString(`{"a":"b"}`)
	assert.Nil(t, err, "Error while decoding")

	node.Dig("a").MutateToFloat(5.6)
	assert.Equal(t, 5.6, node.Dig("a").AsFloat(), "Wrong value")
	assert.Equal(t, 6, node.Dig("a").AsInt(), "Wrong value")

	encoded := node.EncodeToByte()
	assert.Equal(t, `{"a":5.6}`, string(encoded), "Wrong result json")
}

func TestMutateToString(t *testing.T) {
	node, err := DecodeString(`{"a":"b"}`)
	assert.Nil(t, err, "Error while decoding")

	node.Dig("a").MutateToString("insane")
	assert.Equal(t, "insane", node.Dig("a").AsString(), "Wrong value")

	encoded := node.EncodeToByte()
	assert.Equal(t, `{"a":"insane"}`, string(encoded), "Wrong result json")
}

func TestMutateToField(t *testing.T) {
	node, err := DecodeString(`{"a":"b"}`)
	assert.Nil(t, err, "Error while decoding")

	node.AsFields()[0].MutateToField("insane")
	assert.Equal(t, "b", node.Dig("insane").AsString(), "Wrong value")

	encoded := node.EncodeToByte()
	assert.Equal(t, `{"insane":"b"}`, string(encoded), "Wrong result json")
}

func TestAsField(t *testing.T) {
	node, err := DecodeString(`{"a":"b"}`)
	assert.Nil(t, err, "Error while decoding")

	node.AsField("a").MutateToField("insane")
	assert.Equal(t, "b", node.Dig("insane").AsString(), "Wrong value")

	encoded := node.EncodeToByte()
	assert.Equal(t, `{"insane":"b"}`, string(encoded), "Wrong result json")
}

func TestWhitespace(t *testing.T) {
	json := `{
  "one"   :   "one",
  "two"   :   "two"
}`

	node, err := DecodeString(json)
	defer Release(node)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, node, "Node is nil")

	assert.Equal(t, "one", node.Dig("one").AsString(), "Wrong field name")
	assert.Equal(t, "two", node.Dig("two").AsString(), "Wrong field name")
}

func TestObjectManyFieldsSuicide(t *testing.T) {
	root, err := DecodeString(manyFields)
	defer Release(root)

	assert.Nil(t, err, "Error while decoding")
	assert.NotNil(t, root, "Node is nil")

	fields := make([]string, 0, 0)
	for _, field := range root.AsFields() {
		fields = append(fields, string([]byte(field.AsString())))
	}

	for _, field := range root.AsFields() {
		root.Dig(field.AsString()).Suicide()
	}

	for _, field := range fields {
		assert.Nil(t, root.Dig(field), "node should'n be findable")
	}

	assert.Equal(t, `{}`, string(root.EncodeToByte()), "Wrong result json")
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

	assert.Equal(t, `{}`, string(node.EncodeToByte()), "wrong result json")
}

func TestObjectFields(t *testing.T) {
	json := `{"first": "1","second":"2","third":"3"}`
	node, err := DecodeString(json)
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
		out = escapeString(out[:0], test.s)
		assert.Equal(t, string(strconv.AppendQuote(nil, test.s)), string(out), "wrong escaping")
	}
}
