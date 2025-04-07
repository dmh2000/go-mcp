These is a summary of the request/response implementation requirements for an MCP server using the STDIO transport. There are additional requirements for transport SSE subscriptions, not covered here.

# The server will have these core endpoints

- Initialization Endpoint (/initialize)
  - Handles the initial handshake between the client and server.
  - Accepts the protocol version and capabilities from the client.
  - Responds with the server's supported protocol version and capabilities.
  - results in a notification/initialized message with no response
  - supports the following requests:
    - initialize

## Tools Endpoint (/tools)
- Allows clients to call model-controlled functions (tools).
    - Accepts requests with specific parameters and returns results.
    - Example: A weather lookup tool could accept a city name and return weather data.
    - Requirements:   
      - Validate inputs against tool's inputSchema
      - Handle async operations with progress updates via SSE (N/A for stdio)
      - Return structured results (text, files, or structured data)

### /tools/list
        - Implementation: Server returns tool metadata including name, description, and JSON Schema for inputs
        - Purpose: List all available tools and their schemas
        - Parameters: None required

- Request
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "tools/list"
}
```

- Response
```json
{
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
```

### /tools/call
  - Purpose: execute the specific tool

- Request
```json
{
  "name": "calculate_sum",
  "arguments": {
    "a": 5,
    "b": 3
  }
}
```
- Response
```json
{
  "content": [
    {
      "type": "text",
      "text": "8"
    }
  ]
}
```
- Error Handling
  - Invalid Tool: Return {"error": {"code": -32601, "message": "Tool not found"}}
  - Invalid Arguments: Return {"error": {"code": -32602, "message": "Invalid params"}}
- Security Considerations
  - Validate all inputs against JSON Schema before execution
  - Implement OAuth scopes for sensitive operations
  - Use token-based authentication for requests

## Resources Endpoint (/resources)
 - Provides application-controlled data sources.
 - Functions as a REST-style GET endpoint for retrieving data without side effects.
 - Example: Accessing project files or database queries.

### resources/list
  - Purpose: Retrieve available resources
  - Parameters: Optional filters (e.g., type, uriPattern)

- Request
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "resources/list"
}
```

- Response
```json
{
  "resources": [
    {
      "uri": "file:///logs/app.log",
      "name": "Application Logs",
      "mimeType": "text/plain",
      "lastModified": "2025-04-07T09:30:00Z"
    }
  ]
}
```

### resources/read

- Purpose: Fetch resource content
- Parameters: uri (required resource identifier)
- Similar to a REST GET endpoint
- Servers must implement URI validation, content caching, and OAuth2 scopes for sensitive resources. 
- Resources should be treated as read-only GET operations without side effects

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
- There are additional requirements if using the SSE transport, not recovered here.
- resources/subscribe
- resources/unsubscrbe

## Prompts Endpoint (/prompts)
  - Delivers user-controlled prompt templates for inference optimization.
  - Allows retrieval or management of pre-defined prompts or templates.

### /prompts/list
  - Purpose: List available prompt templates
  - Parameters: Optional filters (namePattern, category)
  - Implementation: Return all prompts or filtered subset

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

### /prompts/get

  - Purpose: Retrieve a specific prompt template
  - Parameters: name (required), arguments (dynamic inputs)
  - Requirements: Validate arguments against template schema

- Request

```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "prompts/get",
  "name": "commit",
}
```

- Response
```json
{
  "messages": [
    {
      "role": "user",
      "content": {
        "type": "text",
        "text": "Generate a commit message for:\n\n{{changes}}"
      }
    }
  ]
}
```

- There are additional requirements if using the SSE transport, not recovered here.
- prompts/subscribe
- prompts/unsubscrbe


## Shutdown (/terminate)
  - requests clean shutdown of the server

### /terminate
- Pending Operations: Allow in-flight requests to complete (up to configurable timeout)
- Resource Cleanup: Release file handles, database connections, etc.
- Connection Closure: Close network sockets/stdio channels
- If shutdown takes too long, send an error notification
- Best Practices
  - Implement timeout configuration (default 30 seconds)
  - Use transaction patterns for atomic operations
  - Maintain connection state tracking:

- Request  
```json
{
  "jsonrpc": "2.0",
  "method": "exit",
  "params": {}
}
```

- Response
```json
{
  "jsonrpc": "2.0",
  "result": {}
}
```

## Error (/error)

In the Model Context Protocol (MCP), the /error messages are used to handle and report errors in a structured way, primarily following the JSON-RPC 2.0 standard. These error messages are critical for ensuring that clients and servers can communicate issues effectively and take corrective actions.

- Error Scenarios:
  - Protocol-Level Errors: Issues with message parsing, invalid requests, or unsupported methods.
  - Application-Level Errors: Issues specific to tools, resources, or prompts (e.g., invalid parameters or unavailable resources).
  - Transport Errors: Failures in connection, timeouts, or message delivery.
- Error Propagation:
  - Errors are sent as responses to requests that fail.
  - They include a standardized error object with a code, message, and optional data for additional context.
- Error Codes:
  - MCP uses standard JSON-RPC error codes and allows custom application-specific codes:

- Best Practices for Implementing /error
  - Standardized Codes: Use JSON-RPC standard codes for common errors and custom codes for domain-specific issues.
  - Detailed Context: Include data field to provide additional debugging information (e.g., expected schema, received values).
  - Graceful Degradation: Ensure clients can handle errors without crashing by providing meaningful messages.
  - Logging and Monitoring: Log all errors on the server side for debugging and analytics.


```json
{
  "ParseError": -32700,
  "InvalidRequest": -32600,
  "MethodNotFound": -32601,
  "InvalidParams": -32602,
  "InternalError": -32603
}
```

- Custom error codes should be above -32000.
- Error Response Structure:
  - code: Numeric identifier for the error.
  - message: Short description of the error.
  - data (optional): Additional details about the error.

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

