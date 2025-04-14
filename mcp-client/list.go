package main

import (
	"encoding/json"
	"fmt"
	"sqirvy/mcp/pkg/mcp" // Use the correct module path
)

// --- Helper Functions for MCP List Calls ---

// listTools sends a tools/list request and processes the response.
func (c *Client) listTools() error {
	listID := c.nextID()
	// No parameters needed for a basic list request
	listRequestBytes, err := mcp.MarshalListToolsRequest(listID, nil)
	if err != nil {
		c.logger.Printf("Failed to marshal list tools request: %v", err)
		return fmt.Errorf("failed to marshal list tools request: %w", err)
	}

	c.logger.Println("Sending list tools request...")
	if err := c.transport.WriteMessage(listRequestBytes); err != nil {
		c.logger.Printf("Failed to send list tools request: %v", err)
		return fmt.Errorf("failed to send list tools request: %w", err)
	}

	c.logger.Println("Waiting for list tools response...")
	listResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read list tools response: %v", err)
		return fmt.Errorf("failed to read list tools response: %w", err)
	}
	c.logger.Printf("Received list tools response JSON: %s", string(listResponseBytes))

	listResult, listRespID, listRPCErr, listParseErr := mcp.UnmarshalListToolsResponse(listResponseBytes)
	if listParseErr != nil {
		c.logger.Printf("Failed to parse list tools response: %v", listParseErr)
		return fmt.Errorf("failed to parse list tools response: %w", listParseErr)
	}
	if fmt.Sprintf("%v", listRespID) != fmt.Sprintf("%v", listID) {
		c.logger.Printf("List tools response ID mismatch. Got: %v (%T), Want: %v (%T)", listRespID, listRespID, listID, listID)
		return fmt.Errorf("list tools response ID mismatch. Got: %v, Want: %v", listRespID, listID)
	}
	if listRPCErr != nil {
		c.logger.Printf("Received RPC error in list tools response: Code=%d, Message=%s, Data=%v", listRPCErr.Code, listRPCErr.Message, listRPCErr.Data)
		return fmt.Errorf("received RPC error in list tools response: %w", listRPCErr)
	}
	if listResult == nil {
		c.logger.Println("List tools response contained no result.")
		return fmt.Errorf("list tools response contained no result")
	}

	c.logger.Printf("Available Tools (%d):", len(listResult.Tools))
	for _, tool := range listResult.Tools {
		schemaBytes, _ := json.Marshal(tool.InputSchema) // Marshal schema for logging
		c.logger.Printf("  - Name: %s, Description: %s, Schema: %s", tool.Name, tool.Description, string(schemaBytes))
	}
	if listResult.NextCursor != "" {
		c.logger.Printf("  (Next Cursor: %s)", listResult.NextCursor)
	}

	c.logger.Println("List tools call complete.")
	return nil
}

// listResourceTemplates sends a resources/templates/list request and processes the response.
func (c *Client) listResourceTemplates() error {
	listID := c.nextID()
	// No parameters needed for a basic list request
	listRequestBytes, err := mcp.MarshalListResourceTemplatesRequest(listID, nil)
	if err != nil {
		c.logger.Printf("Failed to marshal list resource templates request: %v", err)
		return fmt.Errorf("failed to marshal list resource templates request: %w", err)
	}

	c.logger.Println("Sending list resource templates request...")
	if err := c.transport.WriteMessage(listRequestBytes); err != nil {
		c.logger.Printf("Failed to send list resource templates request: %v", err)
		return fmt.Errorf("failed to send list resource templates request: %w", err)
	}

	c.logger.Println("Waiting for list resource templates response...")
	listResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read list resource templates response: %v", err)
		return fmt.Errorf("failed to read list resource templates response: %w", err)
	}
	c.logger.Printf("Received list resource templates response JSON: %s", string(listResponseBytes))

	listResult, listRespID, listRPCErr, listParseErr := mcp.UnmarshalListResourceTemplatesResponse(listResponseBytes)
	if listParseErr != nil {
		c.logger.Printf("Failed to parse list resource templates response: %v", listParseErr)
		return fmt.Errorf("failed to parse list resource templates response: %w", listParseErr)
	}
	if fmt.Sprintf("%v", listRespID) != fmt.Sprintf("%v", listID) {
		c.logger.Printf("List resource templates response ID mismatch. Got: %v (%T), Want: %v (%T)", listRespID, listRespID, listID, listID)
		return fmt.Errorf("list resource templates response ID mismatch. Got: %v, Want: %v", listRespID, listID)
	}
	if listRPCErr != nil {
		c.logger.Printf("Received RPC error in list resource templates response: Code=%d, Message=%s, Data=%v", listRPCErr.Code, listRPCErr.Message, listRPCErr.Data)
		return fmt.Errorf("received RPC error in list resource templates response: %w", listRPCErr)
	}
	if listResult == nil {
		c.logger.Println("List resource templates response contained no result.")
		return fmt.Errorf("list resource templates response contained no result")
	}

	c.logger.Printf("Available Resource Templates (%d):", len(listResult.ResourceTemplates))
	for _, template := range listResult.ResourceTemplates {
		c.logger.Printf("  - Name: %s, URI Template: %s, Description: %s, MimeType: %s",
			template.Name, template.URITemplate, template.Description, template.MimeType)
	}
	if listResult.NextCursor != "" {
		c.logger.Printf("  (Next Cursor: %s)", listResult.NextCursor)
	}

	c.logger.Println("List resource templates call complete.")
	return nil
}

// listPrompts sends a prompts/list request and processes the response.
func (c *Client) listPrompts() error {
	listID := c.nextID()
	// No parameters needed for a basic list request
	listRequestBytes, err := mcp.MarshalListPromptsRequest(listID, nil)
	if err != nil {
		c.logger.Printf("Failed to marshal list prompts request: %v", err)
		return fmt.Errorf("failed to marshal list prompts request: %w", err)
	}

	c.logger.Println("Sending list prompts request...")
	if err := c.transport.WriteMessage(listRequestBytes); err != nil {
		c.logger.Printf("Failed to send list prompts request: %v", err)
		return fmt.Errorf("failed to send list prompts request: %w", err)
	}

	c.logger.Println("Waiting for list prompts response...")
	listResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read list prompts response: %v", err)
		return fmt.Errorf("failed to read list prompts response: %w", err)
	}
	c.logger.Printf("Received list prompts response JSON: %s", string(listResponseBytes))

	listResult, listRespID, listRPCErr, listParseErr := mcp.UnmarshalListPromptsResponse(listResponseBytes)
	if listParseErr != nil {
		c.logger.Printf("Failed to parse list prompts response: %v", listParseErr)
		return fmt.Errorf("failed to parse list prompts response: %w", listParseErr)
	}
	if fmt.Sprintf("%v", listRespID) != fmt.Sprintf("%v", listID) {
		c.logger.Printf("List prompts response ID mismatch. Got: %v (%T), Want: %v (%T)", listRespID, listRespID, listID, listID)
		return fmt.Errorf("list prompts response ID mismatch. Got: %v, Want: %v", listRespID, listID)
	}
	if listRPCErr != nil {
		c.logger.Printf("Received RPC error in list prompts response: Code=%d, Message=%s, Data=%v", listRPCErr.Code, listRPCErr.Message, listRPCErr.Data)
		return fmt.Errorf("received RPC error in list prompts response: %w", listRPCErr)
	}
	if listResult == nil {
		c.logger.Println("List prompts response contained no result.")
		return fmt.Errorf("list prompts response contained no result")
	}

	c.logger.Printf("Available Prompts (%d):", len(listResult.Prompts))
	for _, prompt := range listResult.Prompts {
		argsStr := ""
		if len(prompt.Arguments) > 0 {
			args := make([]string, len(prompt.Arguments))
			for i, arg := range prompt.Arguments {
				reqStr := ""
				if arg.Required {
					reqStr = " (required)"
				}
				args[i] = fmt.Sprintf("%s%s", arg.Name, reqStr)
			}
			argsStr = fmt.Sprintf(" Args: [%s]", args)
		}
		c.logger.Printf("  - Name: %s, Description: %s%s", prompt.Name, prompt.Description, argsStr)
	}
	if listResult.NextCursor != "" {
		c.logger.Printf("  (Next Cursor: %s)", listResult.NextCursor)
	}

	c.logger.Println("List prompts call complete.")
	return nil
}
