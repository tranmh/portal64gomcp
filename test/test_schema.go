package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/svw-info/portal64gomcp/internal/mcp"
)

func main() {
	// Create a mock server to test the tool definitions
	server := &mcp.Server{}
	
	// Test the get_player_rating_history function definition
	toolDef := server.GetToolDefinition("get_player_rating_history")
	
	fmt.Println("=== get_player_rating_history Tool Definition ===")
	fmt.Printf("Name: %s\n", toolDef.Name)
	fmt.Printf("Description: %s\n", toolDef.Description)
	
	// Pretty print the schema
	schemaJSON, err := json.MarshalIndent(toolDef.InputSchema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Input Schema:")
	fmt.Println(string(schemaJSON))
	
	// Check if player_id is in required fields
	if len(toolDef.InputSchema.Required) > 0 {
		fmt.Printf("\nRequired parameters: %v\n", toolDef.InputSchema.Required)
	} else {
		fmt.Println("\nNo required parameters found!")
	}
	
	// Test a few other functions that should have required parameters
	fmt.Println("\n=== Other Functions Test ===")
	testFunctions := []string{
		"get_club_profile",
		"get_tournament_details", 
		"get_club_players",
		"get_club_statistics",
		"get_region_addresses",
	}
	
	for _, funcName := range testFunctions {
		def := server.GetToolDefinition(funcName)
		fmt.Printf("%s: required=%v\n", funcName, def.InputSchema.Required)
	}
}
