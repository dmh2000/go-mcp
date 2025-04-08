# initialize

create file pkg/mcp/initialize.go. Include a go type definition for and model context protocol messages initialize for both the requests and responses.  Refer to "design/schema.json" for the json type definitions. in design/schema.json, the types are:

- Initialize
  - InitializeRequest
  - InitializeResult

===================================================================

in pkg/mcp/initialize.go, add a function that marshals and unmarshals a initialize request and result. here are examples. note that the "id" field can be a number or a string

Request
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "roots": {
        "listChanged": true
      },
      "sampling": {}
    },
    "clientInfo": {
      "name": "ExampleClient",
      "version": "1.0.0"
    }
  }
}


Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "logging": {},
      "prompts": {
        "listChanged": true
      },
      "resources": {
        "subscribe": true,
        "listChanged": true
      },
      "tools": {
        "listChanged": true
      }
    },
    "serverInfo": {
      "name": "ExampleServer",
      "version": "1.0.0"
    }
  }
}



===================================================================

create a file, pkg/mcp/initialize_test.go, that tests marshalling and unmarshaling initialization requests and responses 



/add
 ./pkg/mcp/prompts_test.go
 ./pkg/mcp/resources.go
 ./pkg/mcp/tools.go
 ./pkg/mcp/prompts.go
 ./pkg/mcp/tools_test.go
 ./pkg/mcp/resources_test.go
 ./pkg/mcp/types.go
 