
in directory mpc-server, add a handler for the model context protocol 'ping' request and response. refer to design/schema.json for the requirements. put the handler in a new file "mcp-server/ping.go". 
here are examples of the formats
request:

{
  "jsonrpc": "2.0",
  "id": "123",
  "method": "ping"
}


response

{
  "jsonrpc": "2.0",
  "id": "123",
  "result": {}
}
