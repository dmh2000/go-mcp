In directory mcp-client, create a 'go' command line application that implements a model context protocol client. this application is named 'mcp-client' and it is used for testing the mcp-server.

the application will log output to stdout

## logging
The client will include log messages  where needed to facilitate debug. 
The client will log when sending a request.
the client will log when receiving a response.
Logs will includ the json and a descriptive comment.
It will use the golang standard 'log' package.
It will log to stdout


## transport
The application will use json-rpc 2.0 over STDIO  for messages. 
The client will start the mcp-server as a subprocess and initialize the stdio transport.
The default server path will be "./mcp-server". 

## other requirements
The application will use only standard golang libraries. no third party libraries 
The code should be partitioned into logical components in the mcp-server directory where it makes sense.

## Messages
Message  types are in directory "pkg/mcp", including required types and marshal/unmarshal functions for each message type.

To begin with, the client will send an mcp 'initialization' message to the server and wait for a response. after it receives a response it will send the 'notification/initialized' message to the server. it will not expect a response to that message. 

then terminate the application.


MCP-CLIENT: 2025/04/08 12:51:32 client.go:88: Received initialize response (ID: 1). Result: &{Meta:map[] Capabilities:{Experimental:map[] Logging:map[] Prompts:<nil> Resources:<nil> Tools:<nil>} Instructions:Welcome to the Go MCP Example Server! Currently, no tools, prompts, or resources are enabled. ProtocolVersion:2024-11-05 ServerInfo:{Name:GoMCPExampleServer Version:0.1.0}}