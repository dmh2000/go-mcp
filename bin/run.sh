#!/bin/bash
pushd .. && make build && popd
./mcp-client --server-path ./mcp-server --server-log mcp-server.log

