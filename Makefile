.PHONY:	build clean test

build:
	$(MAKE) -C mcp-server build
	$(MAKE) -C mcp-client build

clean:
	$(MAKE) -C mcp-server clean
	$(MAKE) -C mcp-client clean
	@rm -f bin/*

test: build
	./bin/mcp-client -server ./bin/mcp-server
