create file pkg/mcp/prompts.go, it will have a go type definition for and model context protocol prompts/list message request and response.  include functions for marshaling and unmarshaling the request and response types

here are examples of the request and response
- Request
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "prompts/list"
}
```
- Response

```json
{
  "prompts": [
    {
      "name": "git-commit",
      "description": "Generate Git commit messages",
      "arguments": [
        {"name": "changes", "required": true}
      ]
    }
  ]
}
```

in pkg/mcp/prompts.go, add a go type definition for a model context protocol prompts/get request and response. here are examples:
- Request

```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "prompts/get",
  "params": {
    "name": "generate_commit_message",
    "arguments": {
      "changes": "- Fixed a bug in the login flow\n- Improved performance of the dashboard"
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
    "messages": [
      {
        "role": "system",
        "content": {
          "type": "text",
          "text": "You are an AI assistant that generates concise and professional Git commit messages."
        }
      },
      {
        "role": "user",
        "content": {
          "type": "text",
          "text": "Generate a commit message for the following changes:\n\n- Fixed a bug in the login flow\n- Improved performance of the dashboard"
        }
      }
    ]
  }
}

```

