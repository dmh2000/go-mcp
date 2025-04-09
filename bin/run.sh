#!/bin/bash
pushd .. && make build && popd
npx @modelcontextprotocol/inspector ./mcp-server -log ./mcp-server.log

