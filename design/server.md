These is a summary of the request/response implementation requirements for an MCP server using the STDIO transport. There are additional requirements for transport SSE subscriptions, not covered here.

in directory cmd/mcp-server.go, create an mcp server with the following:

# message flow from client to server
client -> initialize request
server -> initialize response
client -> notification/initialized message (server does not respond)

following initialization, the server will have an infinite loop where 
it waits for the client to send a request, and the server will send a response. 
the request/response message types supported include:
- tools/list
- tools/call
- prompts/list
- prompts/get
- resources/list
- resources/read
- error (response if other requests fail on the server side)

### initialize
  - Handles the initial handshake between the client and server.
  - Accepts the protocol version and capabilities from the client.
  - Responds with the server's supported protocol version and capabilities.
  - results in a notification/initialized message with no response

### tools/list
      - Implementation: Server returns tool metadata including name, description, and JSON Schema for inputs
      - Purpose: List all available tools and their schemas
      - Parameters: None required

### tools/call
  - Purpose: execute the specific tool

### resources/list
  - Purpose: Retrieve available resources
  - Parameters: Optional filters (e.g., type, uriPattern)

### resources/read
- Purpose: Fetch resource content
- Parameters: uri (required resource identifier)
- Similar to a REST GET endpoint
- Servers must implement URI validation, content caching, and OAuth2 scopes for sensitive resources. 
- Resources should be treated as read-only GET operations without side effects
- do not support resources/subscribe and resources/unsubscribe

### /prompts/list
  - Purpose: List available prompt templates
  - Parameters: Optional filters (namePattern, category)
  - Implementation: Return all prompts or filtered subset
  - do not add support for prompts/subscribe and prompts/unsubscribe

### prompts/get
  - purpose : The prompts/get method in the Model Context Protocol (MCP) is designed to retrieve and resolve a specific prompt template. Its purpose is to allow clients to dynamically fetch pre-defined prompts from the server, optionally providing arguments to customize the prompt for specific use cases.

### error

The server will return an appropriate 'error' message from the list in pkg/mcp/error.go, in the event any request cannot be completed.
