.PHONY: build clean

build:
	staticcheck ./...
	go build -o ../bin/mcp-client .

 clean:
	@rm -f  mcp-client.log
