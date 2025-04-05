# MCP Client

This is a Go implementation of a Model Context Protocol (MCP) client. The client connects to an MCP server via stdio and can call methods exposed by the server.

## Features

- Starts the MCP server as a subprocess
- Communicates with the server using JSON-RPC over stdio
- Handles initialization and capability discovery
- Provides a clean API for calling server methods

## Usage

To run the client:

```bash
go run main.go
```

The client will:
1. Start the MCP server as a subprocess
2. Process the initialization message from the server
3. Display the server's capabilities
4. Test the available capabilities (currently RandomString)

## Extending

To add support for additional server methods:
1. Define the appropriate request/response types
2. Add a method to the MCPClient struct that calls the server method
3. Check for the capability before calling the method
