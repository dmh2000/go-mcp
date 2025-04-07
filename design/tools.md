# tools/list

create file pkg/mcp/tools.go, it will have a go type definition for and model context protocol tools/list message request and response.  include functions for marshaling and unmarshaling the request and response types

here are examples of the request and response

tools/list request

```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "tools/list",
  "params": {}
}

```

tools/list response

```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "tools": [
        {
        "name": "random_string",
        "description": "return a random string of characters",
        "inputSchema": {
            "type": "object",
            "properties": {
            "length": {"type": "number"},
            },
            "required": ["length"]
        }
    }
  ]
  }
}
```


in pkg/mcp/tools.go, add json type definitions for the tools/call request and response, according to this examples:
- Request
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "tools/call",
  "params": {
    "name": "calculate_sum",
    "arguments": {
      "a": 5,
      "b": 3
    }
  }
}

```
- Response
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "8"
      }
    ]
  }
}

"{\"jsonrpc\":\"2.0\",\"id\":123,\"result\":{\"tools\":[{\"name\":\"tool1\",\"description\":\"desc1\",\"inputSchema\":{\"type\":\"string\"}}]}}"