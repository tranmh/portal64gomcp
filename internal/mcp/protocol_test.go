package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPMessage_Serialization(t *testing.T) {
	testCases := []struct {
		name    string
		message Message
	}{
		{
			name: "Request message",
			message: Message{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "test_method",
				Params:  map[string]string{"key": "value"},
			},
		},
		{
			name: "Response message with result",
			message: Message{
				JSONRPC: "2.0",
				ID:      1,
				Result:  map[string]string{"result": "success"},
			},
		},
		{
			name: "Error response message",
			message: Message{
				JSONRPC: "2.0",
				ID:      1,
				Error: &MCPError{
					Code:    InvalidRequest,
					Message: "Invalid request format",
					Data:    map[string]string{"detail": "missing required field"},
				},
			},
		},
		{
			name: "Notification message",
			message: Message{
				JSONRPC: "2.0",
				Method:  "notification",
				Params:  "test_param",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tc.message)
			require.NoError(t, err)

			// Test JSON deserialization
			var deserializedMessage Message
			err = json.Unmarshal(jsonData, &deserializedMessage)
			require.NoError(t, err)

			// Verify basic fields
			assert.Equal(t, tc.message.JSONRPC, deserializedMessage.JSONRPC)
			
			// ID can be string or number, and JSON converts numbers to float64
			if tc.message.ID != nil {
				switch expectedID := tc.message.ID.(type) {
				case int:
					assert.Equal(t, float64(expectedID), deserializedMessage.ID)
				default:
					assert.Equal(t, tc.message.ID, deserializedMessage.ID)
				}
			}
			
			assert.Equal(t, tc.message.Method, deserializedMessage.Method)

			// Verify error if present
			if tc.message.Error != nil {
				require.NotNil(t, deserializedMessage.Error)
				assert.Equal(t, tc.message.Error.Code, deserializedMessage.Error.Code)
				assert.Equal(t, tc.message.Error.Message, deserializedMessage.Error.Message)
			}
		})
	}
}

func TestMCPError_StandardCodes(t *testing.T) {
	testCases := []struct {
		name         string
		code         int
		expectedName string
	}{
		{"Parse Error", ParseError, "ParseError"},
		{"Invalid Request", InvalidRequest, "InvalidRequest"},
		{"Method Not Found", MethodNotFound, "MethodNotFound"},
		{"Invalid Params", InvalidParams, "InvalidParams"},
		{"Internal Error", InternalError, "InternalError"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &MCPError{
				Code:    tc.code,
				Message: tc.name,
			}

			jsonData, jsonErr := json.Marshal(err)
			require.NoError(t, jsonErr)

			var deserializedError MCPError
			jsonErr = json.Unmarshal(jsonData, &deserializedError)
			require.NoError(t, jsonErr)

			assert.Equal(t, tc.code, deserializedError.Code)
			assert.Equal(t, tc.name, deserializedError.Message)
		})
	}
}

func TestInitializeRequest_Validation(t *testing.T) {
	validRequest := InitializeRequest{
		ProtocolVersion: MCPVersion,
		Capabilities: ClientCapabilities{
			Roots: &RootsCapability{
				ListChanged: true,
			},
		},
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	jsonData, err := json.Marshal(validRequest)
	require.NoError(t, err)

	var deserializedRequest InitializeRequest
	err = json.Unmarshal(jsonData, &deserializedRequest)
	require.NoError(t, err)

	assert.Equal(t, MCPVersion, deserializedRequest.ProtocolVersion)
	assert.Equal(t, "test-client", deserializedRequest.ClientInfo.Name)
	assert.Equal(t, "1.0.0", deserializedRequest.ClientInfo.Version)
}

func TestInitializeResponse_Construction(t *testing.T) {
	response := InitializeResponse{
		ProtocolVersion: MCPVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: true,
			},
			Resources: &ResourcesCapability{
				Subscribe:   true,
				ListChanged: true,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "portal64-mcp",
			Version: "1.0.0",
		},
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var deserializedResponse InitializeResponse
	err = json.Unmarshal(jsonData, &deserializedResponse)
	require.NoError(t, err)

	assert.Equal(t, MCPVersion, deserializedResponse.ProtocolVersion)
	assert.Equal(t, "portal64-mcp", deserializedResponse.ServerInfo.Name)
	assert.NotNil(t, deserializedResponse.Capabilities.Tools)
	assert.NotNil(t, deserializedResponse.Capabilities.Resources)
	assert.True(t, deserializedResponse.Capabilities.Tools.ListChanged)
	assert.True(t, deserializedResponse.Capabilities.Resources.Subscribe)
}

func TestTool_Serialization(t *testing.T) {
	tool := Tool{
		Name:        "search_players",
		Description: "Search for chess players",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results",
					"default":     20,
				},
			},
			Required: []string{"query"},
		},
	}

	jsonData, err := json.Marshal(tool)
	require.NoError(t, err)

	var deserializedTool Tool
	err = json.Unmarshal(jsonData, &deserializedTool)
	require.NoError(t, err)

	assert.Equal(t, "search_players", deserializedTool.Name)
	assert.Equal(t, "Search for chess players", deserializedTool.Description)
	assert.Equal(t, "object", deserializedTool.InputSchema.Type)
	assert.Contains(t, deserializedTool.InputSchema.Properties, "query")
	assert.Contains(t, deserializedTool.InputSchema.Properties, "limit")
}

func TestResource_Serialization(t *testing.T) {
	resource := Resource{
		URI:         "players://12345",
		Name:        "Player Details",
		Description: "Detailed information about a chess player",
		MimeType:    "application/json",
	}

	jsonData, err := json.Marshal(resource)
	require.NoError(t, err)

	var deserializedResource Resource
	err = json.Unmarshal(jsonData, &deserializedResource)
	require.NoError(t, err)

	assert.Equal(t, "players://12345", deserializedResource.URI)
	assert.Equal(t, "Player Details", deserializedResource.Name)
	assert.Equal(t, "Detailed information about a chess player", deserializedResource.Description)
	assert.Equal(t, "application/json", deserializedResource.MimeType)
}

func TestCallToolRequest_Validation(t *testing.T) {
	request := CallToolRequest{
		Name: "search_players",
		Arguments: map[string]interface{}{
			"query": "Carlsen",
			"limit": 10,
		},
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var deserializedRequest CallToolRequest
	err = json.Unmarshal(jsonData, &deserializedRequest)
	require.NoError(t, err)

	assert.Equal(t, "search_players", deserializedRequest.Name)
	assert.NotNil(t, deserializedRequest.Arguments)

	// Arguments is already a map[string]interface{}
	args := deserializedRequest.Arguments
	assert.Equal(t, "Carlsen", args["query"])
	assert.Equal(t, float64(10), args["limit"]) // JSON numbers become float64
}

func TestCallToolResponse_Success(t *testing.T) {
	response := CallToolResponse{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Found 5 players matching 'Carlsen'",
			},
			{
				Type: "application/json",
				Text: `{"players": [{"name": "Magnus Carlsen", "rating": 2830}]}`,
			},
		},
		IsError: false,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var deserializedResponse CallToolResponse
	err = json.Unmarshal(jsonData, &deserializedResponse)
	require.NoError(t, err)

	assert.False(t, deserializedResponse.IsError)
	assert.Len(t, deserializedResponse.Content, 2)
	assert.Equal(t, "text", deserializedResponse.Content[0].Type)
	assert.Equal(t, "application/json", deserializedResponse.Content[1].Type)
	assert.Contains(t, deserializedResponse.Content[0].Text, "Found 5 players")
}

func TestCallToolResponse_Error(t *testing.T) {
	response := CallToolResponse{
		Content: []ToolContent{
			{
				Type: "text",
				Text: "Error: Player not found",
			},
		},
		IsError: true,
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var deserializedResponse CallToolResponse
	err = json.Unmarshal(jsonData, &deserializedResponse)
	require.NoError(t, err)

	assert.True(t, deserializedResponse.IsError)
	assert.Len(t, deserializedResponse.Content, 1)
	assert.Equal(t, "text", deserializedResponse.Content[0].Type)
	assert.Contains(t, deserializedResponse.Content[0].Text, "Error:")
}

func TestMessage_EdgeCases(t *testing.T) {
	testCases := []struct {
		name    string
		jsonStr string
		valid   bool
	}{
		{
			name:    "Valid minimal request",
			jsonStr: `{"jsonrpc":"2.0","method":"test"}`,
			valid:   true,
		},
		{
			name:    "Valid response with null result",
			jsonStr: `{"jsonrpc":"2.0","id":1,"result":null}`,
			valid:   true,
		},
		{
			name:    "Invalid JSONRPC version",
			jsonStr: `{"jsonrpc":"1.0","method":"test"}`,
			valid:   true, // Still parseable, validation would happen at protocol level
		},
		{
			name:    "Empty object",
			jsonStr: `{}`,
			valid:   true, // Parseable but invalid at protocol level
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var message Message
			err := json.Unmarshal([]byte(tc.jsonStr), &message)

			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestListToolsResponse(t *testing.T) {
	response := ListToolsResponse{
		Tools: []Tool{
			{
				Name:        "search_players",
				Description: "Search for players",
			},
			{
				Name:        "get_player",
				Description: "Get player details",
			},
		},
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var deserializedResponse ListToolsResponse
	err = json.Unmarshal(jsonData, &deserializedResponse)
	require.NoError(t, err)

	assert.Len(t, deserializedResponse.Tools, 2)
	assert.Equal(t, "search_players", deserializedResponse.Tools[0].Name)
	assert.Equal(t, "get_player", deserializedResponse.Tools[1].Name)
}

func TestListResourcesResponse(t *testing.T) {
	response := ListResourcesResponse{
		Resources: []Resource{
			{
				URI:         "players://12345",
				Name:        "Player Details",
				Description: "Player information",
			},
			{
				URI:         "clubs://001",
				Name:        "Club Details", 
				Description: "Club information",
			},
		},
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var deserializedResponse ListResourcesResponse
	err = json.Unmarshal(jsonData, &deserializedResponse)
	require.NoError(t, err)

	assert.Len(t, deserializedResponse.Resources, 2)
	assert.Equal(t, "players://12345", deserializedResponse.Resources[0].URI)
	assert.Equal(t, "clubs://001", deserializedResponse.Resources[1].URI)
}
