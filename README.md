# client prompt

in directory mcp-client, create a model context protocol client. the client should do the following:
- the client will include a main function and any other functions as needed
- the code will be written using the go language
- the client should use the go standard library for json-rpc
- the client will start the mcp server as a subprocess
- the client will connect to the mcp server using stdio
- the client will send a model context protocol initialization message to the server
- the client will receive the initialization message from the server
- the client will print the contents of the initialization message
- the client will configure itself to call the methods that the mcp server specifies