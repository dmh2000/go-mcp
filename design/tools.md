# tools

create file pkg/mcp/tools.go. Include a go type definition for and model context protocol messages tools/list and tools/get, for both the requests and responses.  Refer to "design/schema.json" for the json type definitions. in schema.json, the types are:

- tools/list
  - ListToolsRequest
  - ListToolsResult
- tools/read  
  - CallToolsRequest
  - CallToolsResponse

===================================================================

in pkg/mcp/tools.go, add a function that marshals and unmarshals a tools/list request and result. here are examples. note that the "id" field can be a number or a string

Request
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "tools/list",
  "params": {}
}


Response
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "tools": [
      {
        "name": "calculate_sum",
        "description": "Adds two numbers together.",
        "inputSchema": {
          "type": "object",
          "properties": {
            "a": { "type": "number" },
            "b": { "type": "number" }
          },
          "required": ["a", "b"]
        }
      },
      {
        "name": "fetch_weather",
        "description": "Fetches the current weather for a given city.",
        "inputSchema": {
          "type": "object",
          "properties": {
            "city": { "type": "string" }
          },
          "required": ["city"]
        }
      },
      {
        "name": "translate_text",
        "description": "Translates text from one language to another.",
        "inputSchema": {
          "type": "object",
          "properties": {
            "text": { "type": "string" },
            "sourceLanguage": { "type": "string" },
            "targetLanguage": { "type": "string" }
          },
          "required": ["text", "sourceLanguage", "targetLanguage"]
        }
      }
    ]
  }
}

===================================================================

in pkg/mcp/tools.go, add a function that marshals and unmarshals a tools/call request and result. here are examples. note that the "id" field can be a number or a string

request

{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "tools/call",
  "params": {
    "name": "calculate_sum",
    "arguments": {
      "a": 10,
      "b": 15
    }
  }
}


response
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "25"
      }
    ]
  }
}



===================================================================

create a file, pkg/mcp/tools_test.go, that tests marshalling and unmarshaling tools/list and tools/call functions. 


{"jsonrpc":"2.0","result":{"capabilities":{"resources":{},"tools":{}},"instructions":"Welcome to the Go MCP Example Server! The 'random_data' resource and 'ping' tool are available.","protocolVersion":"2024-11-05","serverInfo":{"name":"GoMCPExampleServer","version":"0.1.0"}},"id":0}
