In directory mcp-server, create a 'go' command line application that implements a model context protocol server. 

The application will accept a command line argument "--log <logfile>" 

## logging
The server will include log messages  where needed to facilitate debug. 
The server will log the receipt of client messages including the json, including a descriptive comment.
The server will log when sending a response message, including the json and a descriptive comment.
It will use the golang standard 'log' package.
It will log to a file, not stderr or stdout. 
The default filename will be 'mcp-server.log' unless the --log command line specfies a different filename

## transport
The application will use json-rpc 2.0 over STDIO  for messages.

## other requirements
The application will use only standard golang libraries. no third party libraries 
The code should be partitioned into logical components in the mcp-server directory where it makes sense.

## Messages
Message  types are in directory "pkg/mcp", including required types and marshal/unmarshal functions for each message type.

to begin with, the server will not have any tools, prompts or resources implemented. 
Initially the server will respond to the client 'initialize' request with a response that indicates no capabilties.
The server will return the ErrorCodeMethodNotFound for any other requests that are not supported. Implementations of tools, prompts and resources will be added incrementally later. The server code will attempt to make it reasonably easy to add capabilities incrementally.

 ### Server flow

- server waits for an 'initialize' request from the client
- server responses with an 'initialize' response including a list of capabilties, if any.
- server waits for a 'notification/initialized' message from the client. 
- The server does not respond to 'notification' messages, it just logs them.

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



