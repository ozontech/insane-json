package insaneJSON

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	bigJSON = `{"_id":"5d57e438c48c6e5d4ca83b29","index":0,"guid":"ac784884-a6a3-4987-bb0f-d19e09462677","isActive":false,"balance":"$1,961.04","picture":"http://placehold.it/32x32","age":33,"eyeColor":"blue","name":"McleodMendez","gender":"male","compafny":"PARCOE","com1pany":"PARCOE","co2mpfany":"PARCOE","co222fmpany":"PARCOE","co222empany":"PARCOE","co2222fmpany":"PARCOE","co2222mpany":"PARCOE","compfany":"PARCOE","co2mpany":"PARCOE","compwfany":"PARCOE","compfweany":"PARCOE","comfwepany":"PARCOE","compefefany":"PARCOE","comfeqpany":"PARCOE","comfwefvwepany":"PARCOE","comvfqfqewfpany":"PARCOE","compweewany":"PARCOE","wff":"PARCOE","comqvvpany":"PARCOE","comvqwevpany":"PARCOE","compvany":"PARCOE","compvqeany":"PARCOE","compfanvy":"PARCOE","comspany":"PARCOE","compaany":"PARCOE","compaaqny":"PARCOE","compaqny":"PARCOE","_id1":"5d57e438c48c6e5d4ca83b29","index1":0,"guid1":"ac784884-a6a3-4987-bb0f-d19e09462677","isActive1":false,"balance1":"$1,961.04","picture1":"http://placehold.it/32x32","age1":33,"eyeColor1":"blue","name1":"McleodMendez","gender1":"male","compafny1":"PARCOE","com1pany1":"PARCOE","co2mpfany1":"PARCOE","co222fmpany1":"PARCOE","co222empany1":"PARCOE","co2222fmpany1":"PARCOE","co2222mpany1":"PARCOE","compfany1":"PARCOE","co2mpany1":"PARCOE","compwfany1":"PARCOE","compfweany1":"PARCOE","comfwepany1":"PARCOE","compefefany1":"PARCOE","comfeqpany1":"PARCOE","comfwefvwepany1":"PARCOE","comvfqfqewfpany1":"PARCOE","compweewany1":"PARCOE","wff1":"PARCOE","comqvvpany1":"PARCOE","comvqwevpany1":"PARCOE","compvany1":"PARCOE","compvqeany1":"PARCOE","compfanvy1":"PARCOE","comspany1":"PARCOE","compaany1":"PARCOE","compaaqny1":"PARCOE","compaqny1":"PARCOE"}`
)

func TestDecodeLight(t *testing.T) {
	json := `{"_id":"5d53006246df0b962b787d11","index":"0","guid":"80d75945-6251-46a2-b6c9-a10094beed6e","isActive":"false","balance":"$2,258.24","picture":"http://placehold.it/32x32","age":"34","eyeColor":"brown","company":"NIMON","email":"anne.everett@nimon.name","phone":"+1(946)560-2227","address":"815EmpireBoulevard,Blue,Nevada,5617","about":"Proidentoccaecateulaborislaboreofficialaborumvelitanimnulla.Laboreametoccaecataliquaminimlaboreadenimdolorelaborum.Eiusmodesseeiusmodaliquacillumullamcodonisivelitesseincididunt.Ininestessereprehenderitirureaniminsit.","registered":"Friday,May27,20165:05AM","latitude":"-5.922381","longitude":"-49.143968","greeting":"Hello,Anne!Youhave7unreadmessages.","favoriteFruit":"banana"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")
	assert.True(t, root.IsObject(), "wrong first node")
}

func TestDecodeQuote(t *testing.T) {
	json := `{"log":"{\"ts\":\"2019-10-04T15:54:22.312412503Z\",\"service\":\"oms-go-broker\",\"message\":\"\\u003e0\\","stream":"stdout","time":"2019-10-04T15:54:22.313584867Z"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")
	assert.True(t, root.IsObject(), "wrong first node")
}

func TestDecodeQuoteTriple(t *testing.T) {
	json := `{"log":"{\"ts\":\"2019-10-04T15:54:22.312412503Z\",\"service\":\"oms-go-broker\",\"message\":\"\\u003e0\\\"","stream":"stdout","time":"2019-10-04T15:54:22.313584867Z"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")
	assert.True(t, root.IsObject(), "wrong first node")
}

func TestDecodeReusing(t *testing.T) {
	json := `{"_id":"5d53006246df0b962b787d11","index":"0","guid":"80d75945-6251-46a2-b6c9-a10094beed6e","isActive":"false","balance":"$2,258.24","picture":"http://placehold.it/32x32","age":"34","eyeColor":"brown","company":"NIMON","email":"anne.everett@nimon.name","phone":"+1(946)560-2227","address":"815EmpireBoulevard,Blue,Nevada,5617","about":"Proidentoccaecateulaborislaboreofficialaborumvelitanimnulla.Laboreametoccaecataliquaminimlaboreadenimdolorelaborum.Eiusmodesseeiusmodaliquacillumullamcodonisivelitesseincididunt.Ininestessereprehenderitirureaniminsit.","registered":"Friday,May27,20165:05AM","latitude":"-5.922381","longitude":"-49.143968","greeting":"Hello,Anne!Youhave7unreadmessages.","favoriteFruit":"banana"}`
	root := Spawn()
	err := root.DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")
	assert.True(t, root.IsObject(), "wrong first node")
}

func TestDecodeAdditional(t *testing.T) {
	jsonA := `{"_id":"5d53006246df0b962b787d11"}`
	root, err := DecodeString(jsonA)
	defer Release(root)
	assert.NotNil(t, root, "node shouldn't be nil")
	assert.NoError(t, err, "error while decoding")

	jsonB := `[0, 1, 2]`
	node, err := root.DecodeStringAdditional(jsonB)
	assert.NotNil(t, node, "node shouldn't be nil")
	assert.NoError(t, err, "error while decoding")

	assert.Equal(t, jsonA, root.EncodeToString(), "wrong first node")
	assert.Equal(t, len(jsonA), root.ByteLength(), "wrong byte length")

	assert.Equal(t, 1, node.Dig("1").AsInt(), "wrong node value")
}

func TestDecodeManyObjects(t *testing.T) {
	json := `{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":{"no_key2":"100","somefield":"ok"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"},"no_key1":"100"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")
	assert.True(t, root.IsObject(), "wrong first node")
	assert.True(t, root.AsFields()[0].IsField(), "wrong second node")
}

func TestDecodeArray(t *testing.T) {
	json := `{"first":["s1","s2","s3"],"second":[{"s4":true},{"s5":false}]}`
	node, err := DecodeString(json)
	defer Release(node)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, node, "node shouldn't be nil")

	array := node.Dig("first").AsArray()
	assert.NotNil(t, array, "array shouldn't be empty")
	assert.Equal(t, 3, len(array), "wrong array length")

	assert.Equal(t, "s1", array[0].AsString(), "wrong node value")
	assert.Equal(t, "s2", array[1].AsString(), "wrong node value")
	assert.Equal(t, "s3", array[2].AsString(), "wrong node value")

	array = node.Dig("second").AsArray()
	assert.NotNil(t, array, "array shouldn't be empty")
	assert.NotNil(t, 2, len(array), "wrong array length")
	arrayNode := array[0]
	assert.True(t, arrayNode.IsObject(), "wrong node value")
	assert.True(t, arrayNode.Dig("s4").AsBool(), "wrong node value")

	arrayNode = array[1]
	assert.True(t, arrayNode.IsObject(), "wrong node value")
	assert.Equal(t, false, arrayNode.Dig("s5").AsBool(), "wrong node value")
}

func TestDecodeArrayElements(t *testing.T) {
	json := `["first","second","third"]`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, "first", root.AsArray()[0].AsString(), "wrong array element value")
	assert.Equal(t, "second", root.AsArray()[1].AsString(), "wrong array element value")
	assert.Equal(t, "third", root.AsArray()[2].AsString(), "wrong array element value")
}

func TestDecodeTrueFalseNull(t *testing.T) {
	json := `{"true":true,"false":false,"null":null}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, true, root.Dig("true").AsBool(), "wrong node value")
	assert.Equal(t, false, root.Dig("false").AsBool(), "wrong node value")
	assert.Equal(t, true, root.Dig("null").IsNull(), "wrong node value")
}

func TestDecodeNumber(t *testing.T) {
	json := `{"first": 100,"second": 1e20,"third": 1e+1}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, 100, root.Dig("first").AsInt(), "wrong node value")
	assert.Equal(t, 1e20, root.Dig("second").AsFloat(), "wrong node value")
	assert.Equal(t, 10, root.Dig("third").AsInt(), "wrong node value")
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

		// flagObject
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
			assert.NotNil(t, err, "where should be an error decoding %s", test.json)
			assert.True(t, strings.Contains(err.Error(), test.err.Error()), "wrong err %s, expected=%s, got=%s", test.json, test.err.Error(), err.Error())
		} else {
			assert.NoError(t, err, "where shouldn't be an error %s", test.json)
			root.EncodeToByte()
		}
		Release(root)
	}
}

func TestEncode(t *testing.T) {
	json := `{"key_a":{"key_a_a":["v1","vv1"],"key_a_b":[],"key_a_c":"v3"},"key_b":{"key_b_a":["v3","v31"],"key_b_b":{}}}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, json, root.EncodeToString(), "wrong encoding")
	assert.Equal(t, len(json), root.ByteLength(), "wrong byte length")
}

func TestString(t *testing.T) {
	json := `["hello \\ \" op \\ \" op op","shit"]`

	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, `hello \ " op \ " op op`, root.Dig("0").AsString(), "wrong node value")
	assert.Equal(t, "shit", root.Dig("1").AsString(), "wrong node value")

	assert.Equal(t, json, root.EncodeToString(), "wrong encoding")
	assert.Equal(t, len(json), root.ByteLength(), "wrong byte length")
}

func TestField(t *testing.T) {
	json := `{"hello \\ \" op \\ \" op op":"shit"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, "shit", root.Dig(`hello \ " op \ " op op`).AsString(), "wrong node value")

	assert.Equal(t, json, root.EncodeToString(), "wrong encoding")
	assert.Equal(t, len(json), root.ByteLength(), "wrong byte length")
}

func TestInsane(t *testing.T) {
	test := loadJSON("insane", [][]string{})

	root, err := DecodeBytes(test.json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	encoded := root.EncodeToByte()
	assert.Equal(t, 465158, len(encoded), "wrong encoding")
}

func TestDig(t *testing.T) {
	root, err := DecodeString(bigJSON)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, "ac784884-a6a3-4987-bb0f-d19e09462677", root.Dig("guid").AsString(), "wrong encoding")
}

func TestDigEmpty(t *testing.T) {
	root, err := DecodeString(`{"":""}`)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, "", root.Dig("").AsString(), "wrong node value")
}

func TestDigDeep(t *testing.T) {
	test := loadJSON("heavy", [][]string{})

	root, err := DecodeBytes(test.json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	value := root.Dig("first", "second", "third", "fourth", "fifth", "ok")
	assert.NotNil(t, value, "Can't find field")
	assert.Equal(t, "ok", value.AsString(), "wrong encoding")
}

func TestDigNil(t *testing.T) {
	json := `["first","second","third"]`
	root, _ := DecodeString(json)
	defer Release(root)

	value := root.Dig("one", "$f_8)9").AsString()
	assert.Equal(t, "", value, "wrong array element value")
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
		root, err := DecodeString(test.json)
		assert.NoError(t, err, "where shouldn't be an error")
		for _, field := range test.fields {
			root.AddField(field)
			assert.True(t, root.Dig(field).IsNull(), "wrong node type")
		}
		assert.Equal(t, test.result, root.EncodeToString(), "wrong encoding")
		assert.Equal(t, len(test.result), root.ByteLength(), "wrong byte length")
		Release(root)
	}
}

func TestAddElement(t *testing.T) {
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
		root, err := DecodeString(test.json)
		assert.NoError(t, err, "where shouldn't be an error")
		for index := 0; index < test.count; index++ {
			root.AddElement()
			l := len(root.AsArray())
			assert.True(t, root.Dig(strconv.Itoa(l-1)).IsNull(), "wrong node type")
		}
		assert.Equal(t, test.result, root.EncodeToString(), "wrong encoding")
		assert.Equal(t, len(test.result), root.ByteLength(), "wrong byte length")
		Release(root)
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
		root, err := DecodeString(test.json)
		assert.NoError(t, err, "where shouldn't be an error")
		root.InsertElement(test.pos1)
		assert.True(t, root.Dig(strconv.Itoa(test.pos1)).IsNull(), "wrong node type")
		root.InsertElement(test.pos2)
		assert.True(t, root.Dig(strconv.Itoa(test.pos2)).IsNull(), "wrong node type")

		assert.Equal(t, test.result, root.EncodeToString(), "wrong encoding")
		assert.Equal(t, len(test.result), root.ByteLength(), "wrong byte length")
		Release(root)
	}
}

// it simply shouldn't crush
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

	// hell yeah, we are deterministic
	rand.Seed(666)
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
		assert.NoError(t, err, "err should be nil")
		for range root.AsArray() {
			root.Dig("0").Suicide()
		}
		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `[]`, root.EncodeToString(), "array should be empty")
		assert.Equal(t, len(`[]`), root.ByteLength(), "wrong byte length")
		Release(root)

		root, err = DecodeString(json)
		assert.NoError(t, err, "err should be nil")
		l := len(root.AsArray())
		for i := range root.AsArray() {
			root.Dig(strconv.Itoa(l - i - 1)).Suicide()
		}

		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `[]`, root.EncodeToString(), "array should be empty")
		assert.Equal(t, len(`[]`), root.ByteLength(), "wrong byte length")
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
		assert.NoError(t, err, "err should be nil")
		fields := root.AsFields()
		for _, field := range fields {
			root.Dig(field.AsString()).Suicide()
		}
		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `{}`, root.EncodeToString(), "array should be empty")
		assert.Equal(t, len(`{}`), root.ByteLength(), "wrong byte length")
		Release(root)

		root, err = DecodeString(json)
		assert.NoError(t, err, "err should be nil")
		fields = root.AsFields()
		l := len(fields)
		for i := range fields {
			root.Dig(fields[l-i-1].AsString()).Suicide()
		}
		for _, field := range fields {
			root.Dig(field.AsString()).Suicide()
		}
		assert.Equal(t, 0, len(root.AsArray()), "array should be empty")
		assert.Equal(t, `{}`, root.EncodeToString(), "array should be empty")
		assert.Equal(t, len(`{}`), root.ByteLength(), "wrong byte length")
		Release(root)
	}
}

func TestMergeWith(t *testing.T) {
	jsonA := `{"1":"1","2":"2"}`
	root, err := DecodeString(jsonA)
	defer Release(root)

	assert.NotNil(t, root, "node shouldn't be nil")
	assert.NoError(t, err, "error while decoding")

	jsonB := `{"1":"1","3":"3","4":"4"}`
	node, err := root.DecodeStringAdditional(jsonB)
	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, node, "node shouldn't be nil")

	root.MergeWith(node)

	assert.Equal(t, `{"1":"1","2":"2","3":"3","4":"4"}`, root.EncodeToString(), "wrong first node")
	assert.Equal(t, len(`{"1":"1","2":"2","3":"3","4":"4"}`), root.ByteLength(), "wrong byte length")
}

func TestMergeWithComplex(t *testing.T) {
	jsonA := `{"1":{"1":"1"}}`
	root, err := DecodeString(jsonA)
	defer Release(root)

	assert.NotNil(t, root, "node shouldn't be nil")
	assert.NoError(t, err, "error while decoding")

	jsonB := `{"1":1,"2":{"2":"2"}}`
	node, err := root.DecodeStringAdditional(jsonB)
	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, node, "node shouldn't be nil")

	root.MergeWith(node)

	assert.Equal(t, `{"1":1,"2":{"2":"2"}}`, root.EncodeToString(), "wrong first node")
	assert.Equal(t, len(`{"1":1,"2":{"2":"2"}}`), root.ByteLength(), "wrong byte length")
}

func TestMutateToJSON(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		mutation    string
		dig         []string
		result      string
		checkDig    [][]string
		checkValues []int
	}{
		{
			json:        `{"a":"b","c":"4"}`,
			dig:         []string{"a"},
			mutation:    `5`,
			result:      `{"a":5,"c":"4"}`,
			checkDig:    [][]string{{"a"}, {"c"}},
			checkValues: []int{5, 4},
		},
		{
			json:        `{"a":"1","c":"4"}`,
			dig:         []string{"c"},
			mutation:    `6`,
			result:      `{"a":"1","c":6}`,
			checkDig:    [][]string{{"a"}, {"c"}},
			checkValues: []int{1, 6},
		},
		{
			json:        `{"a":"b","c":"d"}`,
			dig:         []string{"a"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `{"a":{"5":"5","l":[3,4]},"c":"d"}`,
			checkDig:    [][]string{{"a", "l", "0"}, {"a", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			json:        `{"a":"b","c":"d"}`,
			dig:         []string{"c"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `{"a":"b","c":{"5":"5","l":[3,4]}}`,
			checkDig:    [][]string{{"c", "l", "0"}, {"c", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			json:        `{"a":{"somekey":"someval", "xxx":"yyy"},"c":"d"}`,
			dig:         []string{"a"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `{"a":{"5":"5","l":[3,4]},"c":"d"}`,
			checkDig:    [][]string{{"a", "l", "0"}, {"a", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			json:        `["a","b","c","d"]`,
			dig:         []string{"0"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `[{"5":"5","l":[3,4]},"b","c","d"]`,
			checkDig:    [][]string{{"0", "l", "0"}, {"0", "l", "1"}},
			checkValues: []int{3, 4},
		},
		{
			json:        `["a","b","c","d"]`,
			dig:         []string{"3"},
			mutation:    `{"5":"5","l":[3,4]}`,
			result:      `["a","b","c",{"5":"5","l":[3,4]}]`,
			checkDig:    [][]string{{"3", "l", "0"}, {"3", "l", "1"}},
			checkValues: []int{3, 4},
		},
	}

	for _, test := range tests {
		root, err := DecodeString(test.json)
		assert.NoError(t, err, "error while decoding")

		mutatingNode := root.Dig(test.dig...)
		mutatingNode.MutateToJSON(root, test.mutation)
		for i, dig := range test.checkDig {
			assert.Equal(t, test.checkValues[i], root.Dig(dig...).AsInt(), "wrong node value")
			if len(dig) > 1 {
				assert.Equal(t, test.checkValues[i], mutatingNode.Dig(dig[1:]...).AsInt(), "wrong node value")
			}
		}

		assert.Equal(t, test.result, root.EncodeToString(), "wrong result json")
		assert.Equal(t, len(test.result), root.ByteLength(), "wrong byte length")

		Release(root)
	}
}

func TestMutateToObject(t *testing.T) {
	tests := []struct {
		json   string
		dig    []string
		result string
	}{
		{
			json:   `{"a":"b","c":"d"}`,
			dig:    []string{"a"},
			result: `{"a":{},"c":"d"}`,
		},
		{
			json:   `{"a":"b","c":"d"}`,
			dig:    []string{"c"},
			result: `{"a":"b","c":{}}`,
		},
		{
			json:   `["a","b","c","d"]`,
			dig:    []string{"1"},
			result: `["a",{},"c","d"]`,
		},
		{
			json:   `["a","b","c","d"]`,
			dig:    []string{"0"},
			result: `[{},"b","c","d"]`,
		},
		{
			json:   `["a","b","c","d"]`,
			dig:    []string{"3"},
			result: `["a","b","c",{}]`,
		},
		{
			json:   `"string"`,
			dig:    []string{},
			result: `{}`,
		},
		{
			json:   `[]`,
			dig:    []string{},
			result: `{}`,
		},
		{
			json:   `true`,
			dig:    []string{},
			result: `{}`,
		},
		{
			json:   `{"a":{"a":{"a":{"a":"b","c":"d"},"c":"d"},"c":"d"},"c":"d"}`,
			dig:    []string{"a", "a", "a", "a"},
			result: `{"a":{"a":{"a":{"a":{},"c":"d"},"c":"d"},"c":"d"},"c":"d"}`,
		},
	}

	for _, test := range tests {
		root, err := DecodeString(test.json)
		assert.NoError(t, err, "error while decoding")

		mutatingNode := root.Dig(test.dig...).MutateToObject()
		assert.True(t, mutatingNode.IsObject(), "wrong node type")

		o := root.Dig(test.dig...)
		assert.True(t, o.IsObject(), "wrong node type")

		o.AddField("test").MutateToString("ok")
		assert.Equal(t, "ok", o.Dig("test").AsString(), "wrong result json")
		o.Dig("test").Suicide()

		assert.Equal(t, test.result, root.EncodeToString(), "wrong result json")
		assert.Equal(t, len(test.result), root.ByteLength(), "wrong byte length")

		Release(root)
	}
}

func TestMutateToArray(t *testing.T) {
	tests := []struct {
		json   string
		dig    []string
		result string
	}{
		{
			json:   `{"a":"b","c":"d"}`,
			dig:    []string{"a"},
			result: `{"a":[],"c":"d"}`,
		},
		{
			json:   `{"a":"b","c":"d"}`,
			dig:    []string{"c"},
			result: `{"a":"b","c":[]}`,
		},
		{
			json:   `["a","b","c","d"]`,
			dig:    []string{"1"},
			result: `["a",[],"c","d"]`,
		},
		{
			json:   `["a","b","c","d"]`,
			dig:    []string{"0"},
			result: `[[],"b","c","d"]`,
		},
		{
			json:   `["a","b","c","d"]`,
			dig:    []string{"3"},
			result: `["a","b","c",[]]`,
		},
		{
			json:   `"string"`,
			dig:    []string{},
			result: `[]`,
		},
		{
			json:   `[]`,
			dig:    []string{},
			result: `[]`,
		},
		{
			json:   `true`,
			dig:    []string{},
			result: `[]`,
		},
		{
			json:   `{"a":{"a":{"a":{"a":"b","c":"d"},"c":"d"},"c":"d"},"c":"d"}`,
			dig:    []string{"a", "a", "a", "a"},
			result: `{"a":{"a":{"a":{"a":[],"c":"d"},"c":"d"},"c":"d"},"c":"d"}`,
		},
	}

	for _, test := range tests {
		root, err := DecodeString(test.json)
		assert.NoError(t, err, "error while decoding")

		mutatingNode := root.Dig(test.dig...).MutateToArray()
		assert.True(t, mutatingNode.IsArray(), "wrong node type")

		o := root.Dig(test.dig...)
		assert.True(t, o.IsArray(), "wrong node type")

		o.AddElement().MutateToString("ok")
		assert.Equal(t, "ok", o.Dig("0").AsString(), "wrong result json")
		o.Dig("0").Suicide()

		assert.Equal(t, test.result, root.EncodeToString(), "wrong result json")
		assert.Equal(t, len(test.result), root.ByteLength(), "wrong byte length")

		Release(root)
	}
}

func TestMutateCollapse(t *testing.T) {
	tests := []struct {
		json        string
		dig         []string
		mutation    int
		result      string
		checkDig    [][]string
		checkValues []int
	}{
		{
			json:        `{"a":"b","b":["k","k","l","l"],"m":"m"}`,
			dig:         []string{"b"},
			mutation:    15,
			result:      `{"a":"b","b":15,"m":"m"}`,
			checkDig:    [][]string{{"b"}},
			checkValues: []int{15},
		},
		{
			json:        `{"a":"b","b":{"k":"k","l":"l"},"m":"m"}`,
			dig:         []string{"b"},
			mutation:    15,
			result:      `{"a":"b","b":15,"m":"m"}`,
			checkDig:    [][]string{{"b"}},
			checkValues: []int{15},
		},
	}

	for _, test := range tests {
		root, err := DecodeString(test.json)
		assert.NoError(t, err, "error while decoding")

		mutatingNode := root.Dig(test.dig...)
		mutatingNode.MutateToInt(test.mutation)
		for i := range test.checkDig {
			assert.Equal(t, test.checkValues[i], root.Dig(test.checkDig[i]...).AsInt(), "wrong node value")
			assert.Equal(t, test.checkValues[i], mutatingNode.Dig(test.checkDig[i][1:]...).AsInt(), "wrong node value")
		}

		assert.Equal(t, test.result, root.EncodeToString(), "wrong result json")
		assert.Equal(t, len(test.result), root.ByteLength(), "wrong byte length")
		Release(root)
	}
}

func TestMutateToInt(t *testing.T) {
	root, err := DecodeString(`{"a":"b"}`)
	defer Release(root)
	assert.NoError(t, err, "error while decoding")

	root.Dig("a").MutateToInt(5)
	assert.Equal(t, 5, root.Dig("a").AsInt(), "wrong node value")

	assert.Equal(t, `{"a":5}`, root.EncodeToString(), "wrong result json")
	assert.Equal(t, len(`{"a":5}`), root.ByteLength(), "wrong byte length")
}

func TestMutateToFloat(t *testing.T) {
	root, err := DecodeString(`{"a":"b"}`)
	defer Release(root)
	assert.NoError(t, err, "error while decoding")

	root.Dig("a").MutateToFloat(5.6)
	assert.Equal(t, 5.6, root.Dig("a").AsFloat(), "wrong node value")
	assert.Equal(t, 6, root.Dig("a").AsInt(), "wrong node value")

	assert.Equal(t, `{"a":5.6}`, root.EncodeToString(), "wrong result json")
	assert.Equal(t, len(`{"a":5.6}`), root.ByteLength(), "wrong byte length")
}

func TestMutateToString(t *testing.T) {
	root, err := DecodeString(`{"a":"b"}`)
	defer Release(root)
	assert.NoError(t, err, "error while decoding")

	root.Dig("a").MutateToString("insane")
	assert.Equal(t, "insane", root.Dig("a").AsString(), "wrong node value")

	assert.Equal(t, `{"a":"insane"}`, root.EncodeToString(), "wrong result json")
	assert.Equal(t, len(`{"a":"insane"}`), root.ByteLength(), "wrong byte length")
}

func TestMutateToField(t *testing.T) {
	jsons := []string{
		`{"unique":"some_val"}`,
		`{"a":"a","b":"b","c":"c","x1":"x1","a1":"a1","b1":"b1","c1":"c1","x2":"x2","a2":"a2","b2":"b2","c2":"c2","x12":"x12","a12":"a12","b12":"b12","c12":"c12","a121":"a121","b121":"b121","c121":"c121","unique":"some_val"}`,
	}

	for _, json := range jsons {
		root, err := DecodeString(json)
		assert.NoError(t, err, "error while decoding")

		root.DigField("unique").MutateToField("mutated")

		assert.Equal(t, "", root.Dig("unique").AsString(), "wrong node value for %s", json)
		assert.Equal(t, "some_val", root.Dig("mutated").AsString(), "wrong node value for %s", json)
		assert.Equal(t, strings.ReplaceAll(json, "unique", "mutated"), root.EncodeToString(), "wrong result json for %s", json)
		assert.Equal(t, len(strings.ReplaceAll(json, "unique", "mutated")), root.ByteLength(), "wrong byte length for %s", json)

		Release(root)
	}
}

func TestDigField(t *testing.T) {
	root, err := DecodeString(`{"a":"b"}`)
	defer Release(root)
	assert.NoError(t, err, "error while decoding")

	root.DigField("a").MutateToField("insane")
	assert.Equal(t, "b", root.Dig("insane").AsString(), "wrong node value")

	assert.Equal(t, `{"insane":"b"}`, root.EncodeToString(), "wrong result json")
	assert.Equal(t, len(`{"insane":"b"}`), root.ByteLength(), "wrong byte length")
}

func TestWhitespace(t *testing.T) {
	json := `{
 "one"   :   "one",
 "two"   :   "two"
}`

	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, "one", root.Dig("one").AsString(), "wrong field name")
	assert.Equal(t, "two", root.Dig("two").AsString(), "wrong field name")
}

func TestObjectManyFieldsSuicide(t *testing.T) {
	root, err := DecodeString(bigJSON)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	fields := make([]string, 0, 0)
	for _, field := range root.AsFields() {
		fields = append(fields, string(field.AsBytes()))
	}

	for _, field := range root.AsFields() {
		root.Dig(field.AsString()).Suicide()
	}

	for _, field := range fields {
		assert.Nil(t, root.Dig(field), "node should'n be findable")
	}

	assert.Equal(t, `{}`, root.EncodeToString(), "wrong result json")
	assert.Equal(t, len(`{}`), root.ByteLength(), "wrong byte length")
}

func TestObjectManyFieldsAddSuicide(t *testing.T) {
	root, err := DecodeString("{}")
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	fields := 30
	for i := 0; i < fields; i++ {
		root.AddField(strconv.Itoa(i)).MutateToString(strconv.Itoa(i))
	}

	for i := 0; i < fields; i++ {
		assert.NotNil(t, root.Dig(strconv.Itoa(i)), "node should be findable")
		assert.Equal(t, strconv.Itoa(i), root.Dig(strconv.Itoa(i)).AsString(), "wrong node value")
	}

	for i := 0; i < fields; i++ {
		root.Dig(strconv.Itoa(i)).Suicide()
	}

	for i := 0; i < fields; i++ {
		assert.Nil(t, root.Dig(strconv.Itoa(i), "node should be findable"))
	}

	assert.Equal(t, `{}`, root.EncodeToString(), "wrong result json")
	assert.Equal(t, len(`{}`), root.ByteLength(), "wrong byte length")
}

func TestObjectFields(t *testing.T) {
	json := `{"first": "1","second":"2","third":"3"}`
	root, err := DecodeString(json)
	defer Release(root)

	assert.NoError(t, err, "error while decoding")
	assert.NotNil(t, root, "node shouldn't be nil")

	assert.Equal(t, `first`, root.AsFields()[0].AsString(), "wrong field name")
	assert.Equal(t, `second`, root.AsFields()[1].AsString(), "wrong field name")
	assert.Equal(t, `third`, root.AsFields()[2].AsString(), "wrong field name")
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
		{s: `OZON\vshpilevoy`},
	}

	out := make([]byte, 0, 0)
	for _, test := range tests {
		out = escapeString(out[:0], test.s)
		assert.Equal(t, string(strconv.AppendQuote(nil, test.s)), string(out), "wrong escaping")
		size := newBytesCounter().addEscapedString(test.s).total()
		assert.Equal(t, len(string(strconv.AppendQuote(nil, test.s))), size, "wrong string escaping by bytes counter")
	}
}

func TestIndex(t *testing.T) {
	node := Node{}

	index := 5
	node.setIndex(5)

	assert.Equal(t, index, node.getIndex(), "wrong index")
}
