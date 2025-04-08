create file pkg/mcp/error.go. Include a go type definition for and model context protocol 'error' response to any request, in the event there is a problem with the request. refer to JSONRPCError in design/schema.json

- error
  - JSONRPCError

===================================================================

in pkg/mcp/error.go, add a function that marshals and unmarshals an 'error' response. here are examples. note that the "id" field can be a number or a string

- Example 1: Invalid Parameters
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "error": {
    "code": -32602,
    "message": "Invalid parameters",
    "data": {
      "expectedSchema": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "age": { "type": "integer" }
        },
        "required": ["name", "age"]
      },
      "receivedParams": {
        "name": 123
      }
    }
  }
}
```

- Example 2: Method Not Found
```json
{
  "jsonrpc": "2.0",
  "id": "2",
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": {
      "requestedMethod": "/tools/unknownTool"
    }
  }
}
```

- Example 3: Internal Server Error
```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32603,
    "message": "Internal server error",
    "data": {
      "details": "Unexpected null pointer exception in tool execution."
    }
  }
}
```
===================================================================

create a file, pkg/mcp/error_test.go, that tests marshalling and unmarshaling an 'error' response 