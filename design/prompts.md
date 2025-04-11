# prompts

create file pkg/mcp/prompts.go. Include a go type definition for and model context protocol messages prompts/list and prompts/get, for both the requests and responses.  Refer to "design/schema.json" for the json type definitions. in schema.json, the types are:

- prompts/list
  - ListPromptsRequest
  - ListPromptsResult
- prompts/read  
  - GetPromptRequest
  - GetPromptResponse

===================================================================

in pkg/mcp/prompts.go, add a function that marshals and unmarshals a prompts/list request and result. here are examples. note that the "id" field can be a number or a string

Request
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "prompts/list",
  "params": {}
}


Response
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "prompts": [
      {
        "name": "generate_commit_message",
        "description": "Generate a concise Git commit message based on code changes.",
        "arguments": [
          {
            "name": "changes",
            "description": "A description of the code changes.",
            "required": true
          }
        ]
      },
      {
        "name": "summarize_text",
        "description": "Summarize a given block of text into a concise summary.",
        "arguments": [
          {
            "name": "text",
            "description": "The text to summarize.",
            "required": true
          },
          {
            "name": "length",
            "description": "The desired length of the summary (e.g., short, medium, long).",
            "required": false
          }
        ]
      }
    ]
  }
}

===================================================================

in pkg/mcp/prompts.go, add a function that marshals and unmarshals a prompts/get request and result. here are examples. note that the "id" field can be a number or a string

request
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "prompts/get",
  "params": {
    "name": "summarize_text",
    "arguments": {
      "text": "Artificial intelligence is a branch of computer science that aims to create systems capable of performing tasks that typically require human intelligence.",
      "length": "short"
    }
  }
}

response
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "messages": [
      {
        "role": "system",
        "content": {
          "type": "text",
          "text": "You are an AI assistant that generates concise summaries of text."
        }
      },
      {
        "role": "user",
        "content": {
          "type": "text",
          "text": "Summarize the following text into a short summary:\n\nArtificial intelligence is a branch of computer science that aims to create systems capable of performing tasks that typically require human intelligence."
        }
      }
    ]
  }
}

===================================================================

create a file, pkg/mcp/prompts_test.go, that tests marshalling and unmarshaling ompts/list and prompts/get functions. 

===================================================================


the schema for the model context protocol is in file design/schema.json. can you analyze that schema for me and create a summary of the request and response message types. write the result in a new file, "design/messages.md'. use markdown for the output text          


using design/schema.json, add an apendix with two sections: "client to server" to design/messages.md. the first section will include for each client to server request and associated server response, add a simple example of the request and response json. The second section will include for each server to client request and the associated client response


===================================================================

I want to add a new MCP promt that pings network address 192.168.5.4. the tools will be named ping. that will required adding the tool the the handleListTools function, and then it will need to add it to the handleCallTool function. then create a new file, ping.go, that will contain the handler for calling the tool, and then a function names 'ping' that will actually execute the ping command. it will send one ping request and then return the results to the tool handler. then the tool handler will return those results to the mcp client.

I want to add a new MCP prompt named sqirvy_query. that will require adding the prompt to the handleListPrompts function, and then it will need to add it to the handleGetPrompt function for when a client calls it.  create a new file, mcp-server/sqirvy_query.go, this file will contain the handler for calling the prompt, and then a function named "sqirvy_prompt" that returns a prompt as a string. the prompt handler will return the prompt from sqirvy_prompt. the structure is similar to the code in ping.go, except it supports the prompts/list and prompts/get functions instead of tool calls.

can you analyze pkg/mcp and files in mcp-server, and report what types in pkg/mcp are not being used in the mcp-server code.   