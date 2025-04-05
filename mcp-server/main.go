package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"time"
)

// MCPService represents the Model Context Protocol service
type MCPService struct{}

// Initialization related types
type InitArgs struct{}

type InitResponse struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

// RandomStringArgs contains arguments for the RandomString method
type RandomStringArgs struct {
	Length int `json:"length"`
}

// RandomStringResponse contains the response for the RandomString method
type RandomStringResponse struct {
	Result string `json:"result"`
}

// Initialize returns information about the MCP server and its capabilities
func (s *MCPService) Initialize(args *InitArgs, reply *InitResponse) error {
	fmt.Fprintf(os.Stderr, "MCP Server initialized\n")
	*reply = InitResponse{
		Name:    "Go MCP Server",
		Version: "1.0.0",
		Capabilities: []string{
			"random_string",
		},
	}
	return nil
}

// RandomString generates and returns a random string of specified length
func (s *MCPService) RandomString(args *RandomStringArgs, reply *RandomStringResponse) error {
	fmt.Fprintf(os.Stderr, "Generating random string of length %d\n", args.Length)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := args.Length
	if length <= 0 {
		length = 10 // Default length
	}

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	reply.Result = string(b)
	return nil
}

func main() {
	// Register the MCPService
	service := new(MCPService)
	err := rpc.Register(service)
	if err != nil {
		log.Fatalf("Error registering MCPService: %v", err)
	}

	// Setup JSON-RPC server over stdio
	log.Println("MCP Server starting, using stdio for communication")

	// Send initialization message to stdout
	initMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "MCPService.Initialize",
		"params":  []interface{}{InitArgs{}},
		"id":      0,
	}
	initJSON, err := json.Marshal(initMsg)
	if err != nil {
		log.Fatalf("Error marshaling initialization message: %v", err)
	}
	fmt.Println(string(initJSON))

	// Create a codec that communicates over stdin/stdout
	codec := jsonrpc.NewServerCodec(struct {
		*json.Decoder
		*json.Encoder
		*os.File
	}{
		json.NewDecoder(os.Stdin),
		json.NewEncoder(os.Stdout),
		os.Stdin,
	})

	// ServeCodec will block until the client disconnects
	rpc.ServeCodec(codec)
}
