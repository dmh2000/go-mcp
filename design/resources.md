# resources

create file pkg/mcp/resources.go. Include a go type definition for and model context protocol messages resources/list and resources/read, for both the requests and responses.  include functions for marshaling and unmarshaling the request and response types. Refer to "design/schema.json" for the json type definitions. in schema.json, the types are:

- resources/list
  - ListResourcesRequest
  - ListResourcesResult
- resources/read  
  - ReadResourceRequest
  - ReadResourceResponse


in pkg/mcp/resources.go, add a function that marshals and unmarshals a resources list request and result like these examples. note that the "id" field can be a number or a string

Request
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "resources/list"
}

Response
{
  "jsonrpc": "2.0",
  "id": "1",
  "resources": [
    {
      "uri": "file:///logs/app.log",
      "name": "Application Logs",
      "mimeType": "text/plain",
      "lastModified": "2025-04-07T09:30:00Z"
    }
  ]
}


in pkg/mcp/resources.go, add a function that marshals and unmarhsls a resources/read request and result, like these examples:

request
{
  "jsonrpc": "2.0",
  "id": "2",
  "method": "resources/read",
  "params": {
    "uri": "file:///logs/app.log"
  }
}

response
{
  "jsonrpc": "2.0",
  "id": "2",
  "contents": [
    {
      "uri": "file:///logs/app.log",
      "mimeType": "text/plain",
      "text": "2025-04-07 09:00:00 INFO Server started",
      "base64": null
    }
  ]
}

create a file, pkg/mcp/resources_test.go, that uses the standard golang test
framework. it will test the functions in pkg/mcp/resources.go for proper marshalling and unmarshaling


I want to change mcp-server/handlers.go to delegate the actual processing to a separate file. modify handleReadRe source so that it checks which resource is being accessed based on the uri. when it determines the resource it needs to process, call a function in a separate file. in this case, when handleReadResource determines the uri is for data://random_data, perform the processing in an already existing file "random.go". this will make it more straightforward to add more resources in the future.  


now the mcp-server needs to support 'resource templates'. refer to design/schema.json for their structure.  here are examples of the request and response for the random_data uri. make any changes required so the resource read of data://random_data?{length} handles the length argument.

Add a handler for resources/templates/list and add support for data://random_data?{length} 
request:
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "resources/templates/list"
}

response:
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "resourceTemplates": [
      {
        "uriTemplate": "data://random_data?{length}",
        "name": "random_data",
        "description": "Returns a string of random ASCII characters. Use URI like 'data://random_data?length=N' in resources/read, where N is the desired length.",
        "mimeType": "text/plain"
      },
    ]
  }
}

{
  "contents": [
    {
      "uri": "data://random_data?length=8",
      "mimeType": "text/plain",
      "text": "q7O|7;{L",
    }
  ]
}

// resource/read

{
  "jsonrpc": "2.0",
  "method": "resources/read",
  "params": {
    "resource_id": "abc123"
  },
  "id": 1
}

{
  "jsonrpc": "2.0",
  "result": {
    "resource_id": "abc123",
    "name": "Sample Resource",
    "type": "document",
    "content": "This is the content of the resource."
  },
  "id": 1
}

==========================================================
i want to add support for a resources/read request and response in mcp-server. you will probably need to modify handlers.go to register the handler. create new file mcp-server/resources.go and add the mcp handler there. then in directory mcp-server/resources add a function that executes the actual request. refer to design/schema.json for the syntax of a resource/read command. Here are examples of the mcp request and response that should be implemented:
request
```json
{
  "jsonrpc": "2.0",
  "method": "resources/read",
  "params": {
    "uri": "file:///documents/example.txt"
  },
  "id": 42
}


```

response
```json
{
  "jsonrpc": "2.0",
  "result": {
    "contents": "This is the content of example.txt"
  },
  "id": 42
}
```