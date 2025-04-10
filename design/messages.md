# Model Context Protocol (MCP) Message Types

This document summarizes the request and response message types defined in the Model Context Protocol schema.

## Core Message Structure

All MCP messages follow the JSON-RPC 2.0 specification with these basic structures:

- **Request**: Contains `jsonrpc`, `method`, `params`, and `id` fields
- **Response**: Contains `jsonrpc`, `result`, and `id` fields
- **Error Response**: Contains `jsonrpc`, `error`, and `id` fields
- **Notification**: Contains `jsonrpc` and `method` fields (no `id` as no response is expected)

## Client Requests

| Method | Description | Key Parameters |
|--------|-------------|----------------|
| `initialize` | First request sent by client to server | `capabilities`, `clientInfo`, `protocolVersion` |
| `ping` | Check if server is alive | None |
| `resources/list` | Request list of available resources | `cursor` (optional) |
| `resources/templates/list` | Request list of resource templates | `cursor` (optional) |
| `resources/read` | Read a specific resource | `uri` |
| `resources/subscribe` | Subscribe to resource updates | `uri` |
| `resources/unsubscribe` | Unsubscribe from resource updates | `uri` |
| `prompts/list` | Request list of available prompts | `cursor` (optional) |
| `prompts/get` | Get a specific prompt | `name`, `arguments` (optional) |
| `tools/list` | Request list of available tools | `cursor` (optional) |
| `tools/call` | Call a specific tool | `name`, `arguments` (optional) |
| `logging/setLevel` | Set logging level | `level` |
| `completion/complete` | Request completion options | `argument`, `ref` |

## Server Requests

| Method | Description | Key Parameters |
|--------|-------------|----------------|
| `ping` | Check if client is alive | None |
| `sampling/createMessage` | Request LLM sampling via client | `messages`, `maxTokens`, various optional parameters |
| `roots/list` | Request list of root URIs from client | None |

## Client Notifications

| Method | Description | Key Parameters |
|--------|-------------|----------------|
| `notifications/initialized` | Sent after initialization completes | None |
| `notifications/cancelled` | Cancel a previous request | `requestId`, `reason` (optional) |
| `notifications/progress` | Progress update for long-running operation | `progress`, `progressToken`, `total` (optional) |
| `notifications/roots/list_changed` | Inform server that roots list changed | None |

## Server Notifications

| Method | Description | Key Parameters |
|--------|-------------|----------------|
| `notifications/cancelled` | Cancel a previous request | `requestId`, `reason` (optional) |
| `notifications/progress` | Progress update for long-running operation | `progress`, `progressToken`, `total` (optional) |
| `notifications/resources/list_changed` | Resource list has changed | None |
| `notifications/resources/updated` | A resource has been updated | `uri` |
| `notifications/prompts/list_changed` | Prompt list has changed | None |
| `notifications/tools/list_changed` | Tool list has changed | None |
| `notifications/message` | Log message | `data`, `level`, `logger` (optional) |

## Response Types

| Request Method | Response Result Type | Key Fields |
|----------------|----------------------|------------|
| `initialize` | `InitializeResult` | `capabilities`, `protocolVersion`, `serverInfo`, `instructions` (optional) |
| `resources/list` | `ListResourcesResult` | `resources`, `nextCursor` (optional) |
| `resources/templates/list` | `ListResourceTemplatesResult` | `resourceTemplates`, `nextCursor` (optional) |
| `resources/read` | `ReadResourceResult` | `contents` |
| `prompts/list` | `ListPromptsResult` | `prompts`, `nextCursor` (optional) |
| `prompts/get` | `GetPromptResult` | `messages`, `description` (optional) |
| `tools/list` | `ListToolsResult` | `tools`, `nextCursor` (optional) |
| `tools/call` | `CallToolResult` | `content`, `isError` (optional) |
| `completion/complete` | `CompleteResult` | `completion` (contains `values`, `hasMore`, `total`) |
| `sampling/createMessage` | `CreateMessageResult` | `content`, `model`, `role`, `stopReason` (optional) |
| `roots/list` | `ListRootsResult` | `roots` |

## Content Types

The protocol supports several content types that can be exchanged:

- **TextContent**: Plain text with optional annotations
- **ImageContent**: Base64-encoded image data with MIME type
- **EmbeddedResource**: Resource content embedded in a message
- **TextResourceContents**: Text-based resource content
- **BlobResourceContents**: Binary resource content (base64-encoded)

## Error Handling

Error responses follow the JSON-RPC 2.0 specification with these standard error codes:

- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

Additional application-specific error codes may be defined by implementations.
