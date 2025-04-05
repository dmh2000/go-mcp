.PHONY:	build

build:
	go build .


build:
	$(MAKE) -C mcp-server build
	$(MAKE) -C mcp-client build

