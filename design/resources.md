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



CP-SERVER: 2025/04/09 16:28:39 main.go:36: MCP Server starting...
MCP-SERVER: 2025/04/09 16:28:39 main.go:37: Logging to file: ./mcp-server.log
MCP-SERVER: 2025/04/09 16:28:39 main.go:46: Server initialized, starting run loop.
MCP-SERVER: 2025/04/09 16:28:39 server.go:101: Server Run() started.
MCP-SERVER: 2025/04/09 16:28:39 server.go:64: Message type : method=initialize
MCP-SERVER: 2025/04/09 16:28:39 server.go:175: Receive : {"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"sampling":{},"roots":{"listChanged":true}},"clientInfo":{"name":"mcp-inspector","version":"0.8.2"}}}
MCP-SERVER: 2025/04/09 16:28:39 handlers.go:62: Initialize params were not RawMessage initially, re-marshalled: {"capabilities":{"roots":{"listChanged":true},"sampling":{}},"clientInfo":{"name":"mcp-inspector","version":"0.8.2"},"protocolVersion":"2024-11-05"}
MCP-SERVER: 2025/04/09 16:28:39 handlers.go:78: Received Initialize Request (ID: 0): ClientInfo={Name:mcp-inspector Version:0.8.2}, ProtocolVersion=2024-11-05, Caps={Experimental:map[] Roots:0xc0000b25e7 Sampling:map[]}
MCP-SERVER: 2025/04/09 16:28:39 handlers.go:121: Prepared Initialize Response (ID: 0): ServerInfo={Name:GoMCPExampleServer Version:0.1.0}, ProtocolVersion=2024-11-05, Caps={Experimental:map[] Logging:map[] Prompts:<nil> Resources:0xc0000b2648 Tools:<nil>}
MCP-SERVER: 2025/04/09 16:28:39 server.go:290: Send   : {"jsonrpc":"2.0","result":{"capabilities":{"resources":{}},"instructions":"Welcome to the Go MCP Example Server! The 'random_data' resource is available via resources/read.","protocolVersion":"2024-11-05","serverInfo":{"name":"GoMCPExampleServer","version":"0.1.0"}},"id":0}
MCP-SERVER: 2025/04/09 16:28:39 server.go:192: Initialize response sent
MCP-SERVER: 2025/04/09 16:28:39 server.go:64: Message type : method=notifications/initialized
MCP-SERVER: 2025/04/09 16:28:39 server.go:175: Receive : {"jsonrpc":"2.0","method":"notifications/initialized"}
MCP-SERVER: 2025/04/09 16:28:39 server.go:207: Warning: Received duplicate 'notifications/initialized' notification after already initialized. Ignoring.
