# Insane JSON

Fast, zero-allocation JSON library for Go. Decode, navigate, mutate, and encode JSON without unmarshalling into Go structs. Designed for high-throughput pipelines where performance matters.

## Installation

```bash
go get github.com/ozontech/insane-json
```

## Quick Start

```go
root, err := insaneJSON.DecodeString(`{"name":"John","age":30}`)
if err != nil {
    panic(err)
}
defer insaneJSON.Release(root)

name := root.Dig("name").AsString()   // "John"
age := root.Dig("age").AsInt()        // 30

root.Dig("age").MutateToInt(31)
root.AddField("active").MutateToBool(true)

output := root.Encode(nil) // []byte: {"name":"John","age":31,"active":true}
```

## Examples

### Extracting fields from API response

```go
root, err := insaneJSON.DecodeBytes(responseBody)
if err != nil {
    return err
}
defer insaneJSON.Release(root)

status := root.Dig("response", "status").AsString()
code := root.Dig("response", "code").AsInt()
items := root.Dig("response", "data", "items")

if items.IsArray() {
    for _, item := range items.AsArray() {
        id := item.Dig("id").AsInt()
        name := item.Dig("name").AsString()
        fmt.Printf("id=%d name=%s\n", id, name)
    }
}
```

### Transforming JSON logs

```go
root, err := insaneJSON.DecodeBytes(logLine)
if err != nil {
    return err
}
defer insaneJSON.Release(root)

// add tracing info
root.AddField("trace_id").MutateToString(traceID)
root.AddField("processed_at").MutateToString(time.Now().Format(time.RFC3339))

// remove sensitive data
root.Dig("request", "headers", "Authorization").Suicide()
root.Dig("request", "body", "password").Suicide()

// rename field
root.DigField("level").MutateToField("log_level")

output = root.Encode(output[:0])
```

### Filtering array elements

```go
root, err := insaneJSON.DecodeString(`{"users":[{"name":"Alice","active":true},{"name":"Bob","active":false},{"name":"Carol","active":true}]}`)
if err != nil {
    return err
}
defer insaneJSON.Release(root)

users := root.Dig("users")
for _, user := range users.AsArray() {
    if !user.Dig("active").AsBool() {
        user.Suicide()
    }
}

fmt.Println(root.EncodeToString())
// {"users":[{"name":"Alice","active":true},{"name":"Carol","active":true}]}
```

### High-throughput processing with Root reuse

```go
root := insaneJSON.Spawn()
defer insaneJSON.Release(root)

buf := make([]byte, 0, 4096)

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    if err := root.DecodeBytes(scanner.Bytes()); err != nil {
        continue
    }

    root.AddField("source").MutateToString("pipeline-v2")

    buf = root.Encode(buf[:0])
    writer.Write(buf)
}
```

### Working with nested JSON

```go
root, err := insaneJSON.DecodeString(`{"a":{"b":{"c":"deep"}}}`)
if err != nil {
    return err
}
defer insaneJSON.Release(root)

// Dig traverses nested objects
value := root.Dig("a", "b", "c").AsString() // "deep"

// array elements accessed by string index
root2, _ := insaneJSON.DecodeString(`{"items":["zero","one","two"]}`)
defer insaneJSON.Release(root2)

second := root2.Dig("items", "1").AsString() // "one"
```

### Strict mode with error handling

```go
root, err := insaneJSON.DecodeString(`{"count":"not a number"}`)
if err != nil {
    return err
}
defer insaneJSON.Release(root)

node, err := root.DigStrict("count")
if err != nil {
    return err // insaneJSON.ErrNotFound
}

count, err := node.AsInt()
if err != nil {
    return err // insaneJSON.ErrNotNumber
}
```

### Merging objects

```go
root, _ := insaneJSON.DecodeString(`{"a":"1","b":"2"}`)
defer insaneJSON.Release(root)

patch, _ := root.DecodeStringAdditional(`{"b":"updated","c":"3"}`)

root.MergeWith(patch)
fmt.Println(root.EncodeToString())
// {"a":"1","b":"updated","c":"3"}
```

## API Overview

### Decode

| Function | Description |
|---|---|
| `DecodeString(json) (*Root, error)` | Decode JSON string, returns Root from pool |
| `DecodeBytes(json) (*Root, error)` | Decode JSON byte slice, returns Root from pool |
| `Spawn() *Root` | Get an empty Root from pool |
| `Release(root)` | Return Root to pool |
| `root.DecodeString(json) error` | Reuse Root to decode another JSON |
| `root.DecodeBytes(json) error` | Reuse Root to decode another JSON |
| `root.DecodeStringAdditional(json) (*Node, error)` | Decode JSON using Root's node pool without clearing |
| `root.DecodeBytesAdditional(json) (*Node, error)` | Decode JSON using Root's node pool without clearing |

### Navigate

| Function | Description |
|---|---|
| `node.Dig(path...) *Node` | Navigate to nested value. Returns nil if not found. You can also access elements by index. See [Working with nested JSON](#working-with-nested-json) |
| `node.DigStrict(path...) (*StrictNode, error)` | Same as Dig but returns error if not found |
| `node.AsFields() []*Node` | Get object field nodes |
| `node.AsArray() []*Node` | Get array element nodes |
| `node.AsFieldValue() *Node` | Get value node from field node |
| `node.DigField(path...) *Node` | Get field node (not value) at path |

### Read Values

| Function | Description |
|---|---|
| `node.AsString() string` | Get string value |
| `node.AsInt() int` | Get integer value |
| `node.AsInt64() int64` | Get int64 value |
| `node.AsUint64() uint64` | Get uint64 value |
| `node.AsFloat() float64` | Get float64 value |
| `node.AsBool() bool` | Get bool value |
| `node.AsBytes() []byte` | Get value as byte slice |
| `node.AsEscapedString() string` | Get JSON-escaped string value |

### Type Checks

| Function | Description |
|---|---|
| `node.IsObject() bool` | Is value an object? |
| `node.IsArray() bool` | Is value an array? |
| `node.IsString() bool` | Is value a string? |
| `node.IsNumber() bool` | Is value a number? |
| `node.IsTrue() bool` | Is value true? |
| `node.IsFalse() bool` | Is value false? |
| `node.IsNull() bool` | Is value null? |
| `node.IsNil() bool` | Is node nil? |

### Modify

| Function | Description |
|---|---|
| `node.MutateToString(v)` | Set value to string |
| `node.MutateToInt(v)` | Set value to int |
| `node.MutateToFloat(v)` | Set value to float64 |
| `node.MutateToBool(v)` | Set value to bool |
| `node.MutateToNull()` | Set value to null |
| `node.MutateToObject()` | Set value to empty object |
| `node.MutateToArray()` | Set value to empty array |
| `node.MutateToJSON(root, json)` | Set value to parsed JSON |
| `node.MutateToField(name)` | Rename object field |
| `node.MutateToNode(other)` | Copy another node's value |
| `node.Suicide()` | Remove node from parent |
| `node.AddField(name) *Node` | Add field to object, returns value node |
| `node.AddElement() *Node` | Append element to array |
| `node.InsertElement(pos) *Node` | Insert element at position |
| `node.MergeWith(other)` | Merge other object's fields into this one |

### Encode

| Function | Description |
|---|---|
| `node.Encode(buf) []byte` | Encode to byte slice, reusing buf |
| `node.EncodeToByte() []byte` | Encode to new byte slice |
| `node.EncodeToString() string` | Encode to string |

## Important Notes

### Pool and Lifecycle

Decoded nodes live inside a pool managed by the Root. After calling `Release(root)`, the Root and all its nodes are returned to the pool and **must not be used**. Accessing nodes after Release leads to undefined behavior.

```go
root, _ := insaneJSON.DecodeString(`{"a":"b"}`)
node := root.Dig("a")

insaneJSON.Release(root)

// BUG: node belongs to the released root, this is undefined behavior
fmt.Println(node.AsString())
```

Always use `defer insaneJSON.Release(root)` right after decode.

### Thread Safety

The top-level functions `DecodeString`, `DecodeBytes`, and `Spawn` are safe to call from multiple goroutines — they use `sync.Pool` internally.

However, a specific Root and its Nodes are **not thread-safe**. Do not share a Root between goroutines without synchronization. The typical pattern is one Root per goroutine:

```go
// correct: each goroutine gets its own Root
for _, data := range items {
    go func(d []byte) {
        root, err := insaneJSON.DecodeBytes(d)
        if err != nil {
            return
        }
        defer insaneJSON.Release(root)
        // work with root...
    }(data)
}
```

### Nil-safe Navigation

`Dig` on a nil node returns nil without panicking. This allows safe chaining:

```go
// even if "a" doesn't exist, this won't panic — returns 0
value := root.Dig("a", "b", "c").AsInt()
```

`As*` methods on nil nodes return zero values (`""`, `0`, `false`).

Use `DigStrict` when you need to distinguish "field not found" from "field is zero value":

```go
node, err := root.DigStrict("user", "email")
if err != nil {
    // field doesn't exist
}
email, err := node.AsString()
if err != nil {
    // field exists but is not a string
}
```

### Memory Management

For best performance, reuse Root objects instead of decoding into new ones:

```go
root := insaneJSON.Spawn()
defer insaneJSON.Release(root)

for _, msg := range messages {
    root.DecodeBytes(msg)   // reuses internal buffers
    process(root)
}
```

Use `root.ReleaseMem()` after processing an unusually large JSON to free internal buffers:

```go
root.DecodeBytes(hugeJSON)
process(root)
root.ReleaseMem() // release internal buffers to GC
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `insaneJSON.StartNodePoolSize` | 128 | Initial number of pre-allocated nodes per Root |
| `insaneJSON.MapUseThreshold` | 16 | Object field count above which Dig builds a hash map for O(1) lookup |
| `insaneJSON.DisableBeautifulErrors` | false | Set to true to skip formatting decode error messages for better performance |

```go
func init() {
    insaneJSON.StartNodePoolSize = 256
    insaneJSON.MapUseThreshold = 32
    insaneJSON.DisableBeautifulErrors = true
}
```
