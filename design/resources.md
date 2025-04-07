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