In directory mcp-client, create a 'go' command line application that implements a model context protocol client that is compatible with mcp/mcp-server. this application is named 'mcp-client' and it is used for testing the mcp-server. the applicatiuon is in mcp/mcp-client and all new files should be added to that directory.

currently, the application in mcp/mcp-client/main.go does:
- get an optional command line argument for the path of the mcp-server log.
- get an optional command line argument for the path of the mcp-server executable
- opens stdin for input from the mcp-server
- opens stdout for output to the mcp-server

## logging
The client will include log messages  where needed to facilitate debug. 
The client will log when sending a request.
the client will log when receiving a response.
Logs will output the json request and response only to logs
It will use the golang standard 'log' package.
It will log to stdout

## transport
The application will use json-rpc 2.0 over STDIO  for messages. 
The client will spawn the mcp-server as a subprocess and initialize the stdio transport.
The default server path will be "./mcp-server". 

## other requirements
The application will use only standard golang libraries. no third party libraries 
The code should be partitioned into logical components in the mcp-server directory where it makes sense.

## Messages
Message  types are in directory "pkg/mcp", including required types and marshal/unmarshal functions for each message type.

To begin with, the client will send an mcp 'initialization' message to the server and wait for a response. after it receives a response it will send the 'notification/initialized' message to the server. it will not expect a response to that message. 
then terminate the application.

======================================================================================
now, in mcp-client/client.go, instead of terminating after the initialization sequence, call the 'ping' mcp tool in the server and log its output. then terminate the client 

======================================================================================
now, in mcp-client/client.go, instead of terminating after the ping call, execute the resource template for "data://random_data" specifying a length of 10 random characters

======================================================================================
now, in mcp-client/client.go, instead of terminating after the resource template call, execute the prompt call for 'query'

======================================================================================
in mcp-client/client.go, move the ping mcp tool request and response code into a separate function. do the same for the ping call and the resource template call. then have the client Run function must call those functions with the appropriate parameters instead of executing the calls inline

======================================================================================
now create a new file "list.go" in directory mcp-client, and add functions in that file that execute the mcp requests for tools/list, resources/templates/list and prompts/list. each in a separate function. then execute this functions in the client Run function 
