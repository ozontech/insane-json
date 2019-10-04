# Insane JSON
Lighting fast and simple JSON decode/encode library for GO

## Key features
To be filled

## Usage
```go
    // ==== DECODE API ====
    root, err = insaneJSON.DecodeString(jsonString)        // from string
    root, err = insaneJSON.DecodeBytes(jsonBytes)          // from byte slice
    defer insaneJSON.Release(root)                         // place root back to pool 

    // ==== GET API ====
    code = root.Dig("response", "code").AsInt()            // int from objects
    body = root.Dig("response", "body").AsString()         // string from objects

    keys = []string{"items", "3", "name"} 
    thirdItemName = root.Dig(keys...).AsString()           // string from objects and array

    // ==== CHECK API ====
    isObject = root.Dig("response").IsObject()             // is value object?
    isInt = root.Dig("response", "code").IsInt()           // is value null?
    isArray = root.Dig("items").IsArray()                  // is value array?

    // ==== DELETE API ====
    root.Dig("response", "code").Suicide()                 // delete object field
    root.Dig("items", "3").Suicide()                       // delete array element
    anyDugNode.Suicide()                                   // delete any previously dug node

    // ==== MODIFY API ====
    root.Dig("response", "code").MutateToString("OK")      // convert to string
    root.Dig("items", "3").MutateToObject()                // convert to empty object

    item = `{"name":"book","weight":1000}`
    err = root.Dig("items", "3").MutateToJSON(item)        // convert to parsed JSON  

    // ==== OBJECT API ====
    response = root.Dig("response")                        // get object
    fields = response.AsFields()                           // get object fields

    for _, field = range(fields) {                         
        fmt.Println(field.AsField())                       // print all object fields 
    }

    for _, field = range(fields) {                         
        response.Dig(field.AsField()).Suicide()            // remove all fields            
    }

    for _, field = range(fields) {                         
        field.Suicide()                                    // simpler way to remove all fields
    }
    
    header="Content-Encoding: gzip"
    response.AddField("header").MutateToString(header)     // add new field and set value 

    // ==== ARRAY API ====
    items = root.Dig("items")                              // get array
    elements = items.AsArray()                             // get array elements

    for _, element = range(elements) {                     
        fmt.Println(element.AsString())                    // print all array elements    
    }

    for _, element = range(elements) {                     
        element.Suicide()                                  // remove all elements
    }

    item = `{"name":"book","weight":1000}`
    err = items.AddElement().MutateToJSON(item)            // add new element and set value 

    // ==== ENCODE API ====
    To be filled

    // ==== STRICT API ====
    items = root.Dig("items").InStrictMode()               // convert value to strict mode
    items, err = root.DigStrict("items")                   // or get strict value directly
         
    o, err = items.AsObject()                              // now value has api with error handling  
    name, err = items.Dig("5").Dig("name").AsInt           // err won't be nil since name is a string 

    // ==== POOL API ====
    root, err = insaneJSON.DecodeString(json)              // get a root from the pool and place decoded json into it 
    emptyRoot = insaneJSON.Spawn()                         // get an empty root from the pool            

    root.DecodeString(emptyRoot, anotherJson)              // reuse a root to decode another JSONs

    insaneJSON.Release(root)                               // place roots back to the pool
    insaneJSON.Release(emptyRoot)                           
```

## Benchmarks
To be filled