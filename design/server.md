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



===========================================================

add a new file mcp-server/resources.go. in this file add a function. RandomData that is given a parameter for length and it returns a string of random ASCII characters of that length and an error. 
if length is less than or equal to 0 return an error. if the length is greater than 1024 characters, return an error.
use the crypto/rand function for the random data. remove bias from the output using rejection sampling. 

create an mcp 'resource' with 
- uri data://random_data?length=n 
- name 'random_data', 
- description  'returns a string of random characters of length 'n'"
mime type: 'text/plain'. 
the random_data resource will use the the RandomData function from resources.go to generate a string of random characters of length 'n' as specifed in the uri. add this to the resources/list and resources/read responses.

echo "{\"jsonrpc\":\"2.0\",\"method\":\"initialize\",\"params\":{\"capabilities\":{},\"clientInfo\":{\"name\":\"GoMCPExampleClient\",\"version\":\"0.1.0\"},\"protocolVersion\":\"2024-11-05\"},\"id\":1}" | ./mcp-server

Traceback (most recent call last): File "/home/dmh2000/projects/mcp/test/./stdio_reader.py", line 76, in <module>
main() File "/home/dmh2000/projects/mcp/test/./stdio_reader.py", line 62, in main read_stdin_task = anyio.create_task_group() ^^^^^^^^^^^^^^^^^^^^^^^^^ File "/home/dmh2000/.local/miniconda3/envs/aider/lib/python3.12/site-packages/anyio/_core/_tasks.py", line 158, in create_task_group return get_async_backend().create_task_group() ^^^^^^^^^^^^^^^^^^^ File "/home/dmh2000/.local/miniconda3/envs/aider/lib/python3.12/site-packages/anyio/_core/_eventloop.py", line 156, in get_async_backend asynclib_name = sniffio.current_async_library() ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ File "/home/dmh2000/.local/miniconda3/envs/aider/lib/python3.12/site-packages/sniffio/_impl.py", line 93, in current_async_library raise AsyncLibraryNotFoundError( sniffio._impl.AsyncLibraryNotFoundError: unknown async library, or not in async context
Traceback (most recent call last): File "/home/dmh2000/projects/mcp/test/./stdio_reader.py", line 76, in <module> main() File "/home/dmh2000/projects/mcp/test/./stdio_reader.py", line 62, in main
read_stdin_task = anyio.create_task_group() ^^^^^^^^^^^^^^^^^^^^^^^^^ File "/home/dmh2000/.local/miniconda3/envs/aider/lib/python3.12/site-packages/anyio/_core/_tasks.py", line 158, in create_task_group return get_async_backend().create_task_group() ^^^^^^^^^^^^^^^^^^^ File "/home/dmh2000/.local/miniconda3/envs/aider/lib/python3.12/site-packages/anyio/_core/_eventloop.py", line 156, in get_async_backend asynclib_name = sniffio.current_async_library() ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ File "/home/dmh2000/.local/miniconda3/envs/aider/lib/python3.12/site-packages/sniffio/_impl.py", line 93, in current_async_library raise AsyncLibraryNotFoundError( sniffio._impl.AsyncLibraryNotFoundError: unknown async library, or not in


================================================================================================
#### the mcp-client and mcp-server communicate with each other, however when I try to connect to the server with another known good client, it does not connect. is there a problem in transport.go/readMessage. it uses the net/textproto reader. is that compatible with json-rpc?  

You are correct to question the transport layer, as that's a common source of interoperability issues.

However, the use of `net/textproto` in `readMessage` is actually **correct** for the standard transport protocol used by MCP (which is derived from the Language Server Protocol). This protocol requires messages to be framed with:

1.  HTTP-style headers (like `Content-Length: NNN\r\n`).
2.  A blank line (`\r\n`).
3.  The JSON payload itself.

......

#### But the jsonrpc protocol does not use HTTP headers

After that I backed out what the ai did and scafford in what I thought was correct, and then let the AI fix it up and integrate it into the code. then the transport started working.