# resources/list

create file pkg/mcp/resources.go, it will have a go type definition for and model context protocol resources/list message request and response.  include functions for marshaling and unmarshaling the request and response types

here are examples of the request and response

- Request
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "resources/list",
  "params": {}
}

```

- Response
```json
{
    "jsonrpc": "2.0",
    "id": "1",
    "result": {
        "resources": [
            {
            "uri": "file:///logs/app.log",
            "name": "Application Logs",
            "mimeType": "text/plain",
            "lastModified": "2025-04-07T09:30:00Z"
            }
        ]
    }
}
```


in pkg/mcp/resources.go, add json type definitions for the resources/read request and response, according to this examples:
- Request
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "resources/read",
  "uri": "file:///logs/app.log"
}
```

- Response

```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "contents": [
    {
      "uri": "file:///logs/app.log",
      "mimeType": "text/plain",
      "text": "2025-04-07 09:00:00 INFO Server started",
      "base64": null
    }
  ]
}
```