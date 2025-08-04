package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/svw-info/portal64gomcp/internal/api"
	"github.com/svw-info/portal64gomcp/internal/config"
)

// Server represents the MCP server
type Server struct {
	config     *config.Config
	logger     *logrus.Logger
	apiClient  *api.Client
	tools      map[string]ToolHandler
	resources  map[string]ResourceHandler
	listener   net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
}

// ToolHandler represents a function that handles tool calls
type ToolHandler func(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error)

// ResourceHandler represents a function that handles resource requests
type ResourceHandler func(ctx context.Context, uri string) (*ReadResourceResponse, error)

// NewServer creates a new MCP server
func NewServer(cfg *config.Config, logger *logrus.Logger, apiClient *api.Client) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	
	server := &Server{
		config:    cfg,
		logger:    logger,
		apiClient: apiClient,
		tools:     make(map[string]ToolHandler),
		resources: make(map[string]ResourceHandler),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Register tools and resources
	server.registerTools()
	server.registerResources()

	return server
}

// Start starts the MCP server
func (s *Server) Start() error {
	s.logger.Info("Starting MCP server on stdio")

	// MCP servers typically use stdio for communication
	return s.handleStdioConnection()
}

// Stop stops the MCP server
func (s *Server) Stop() {
	s.logger.Info("Stopping MCP server")
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}
}

// handleStdioConnection handles stdio-based communication
func (s *Server) handleStdioConnection() error {
	scanner := bufio.NewScanner(os.Stdin)
	writer := os.Stdout

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		s.logger.WithField("message", line).Debug("Received message")

		response, err := s.handleMessage([]byte(line))
		if err != nil {
			s.logger.WithError(err).Error("Error handling message")
			continue
		}

		if response != nil {
			responseData, err := SerializeMessage(response)
			if err != nil {
				s.logger.WithError(err).Error("Error serializing response")
				continue
			}

			s.logger.WithField("response", string(responseData)).Debug("Sending response")
			
			if _, err := writer.Write(responseData); err != nil {
				s.logger.WithError(err).Error("Error writing response")
				continue
			}
			
			if _, err := writer.Write([]byte("\n")); err != nil {
				s.logger.WithError(err).Error("Error writing newline")
				continue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	return nil
}
// handleMessage processes incoming MCP messages
func (s *Server) handleMessage(data []byte) (*Message, error) {
	msg, err := ParseMessage(data)
	if err != nil {
		return NewErrorResponse(nil, ParseError, "Parse error", err.Error()), nil
	}

	// Handle notifications (no response expected)
	if msg.ID == nil {
		return s.handleNotification(msg)
	}

	// Handle requests
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "tools/list":
		return s.handleListTools(msg)
	case "tools/call":
		return s.handleCallTool(msg)
	case "resources/list":
		return s.handleListResources(msg)
	case "resources/read":
		return s.handleReadResource(msg)
	default:
		return NewErrorResponse(msg.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", msg.Method), nil), nil
	}
}

// handleNotification processes MCP notifications
func (s *Server) handleNotification(msg *Message) (*Message, error) {
	switch msg.Method {
	case "notifications/initialized":
		s.logger.Info("Client initialized")
		return nil, nil
	default:
		s.logger.WithField("method", msg.Method).Warn("Unknown notification method")
		return nil, nil
	}
}

// handleInitialize processes initialization requests
func (s *Server) handleInitialize(msg *Message) (*Message, error) {
	var req InitializeRequest
	if err := s.parseParams(msg.Params, &req); err != nil {
		return NewErrorResponse(msg.ID, InvalidParams, "Invalid parameters", err.Error()), nil
	}

	s.logger.WithFields(logrus.Fields{
		"client":           req.ClientInfo.Name,
		"client_version":   req.ClientInfo.Version,
		"protocol_version": req.ProtocolVersion,
	}).Info("Client initializing")

	response := InitializeResponse{
		ProtocolVersion: MCPVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: true,
			},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: true,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "portal64gomcp",
			Version: "1.0.0",
		},
	}

	return NewSuccessResponse(msg.ID, response), nil
}

// handleListTools processes tool listing requests
func (s *Server) handleListTools(msg *Message) (*Message, error) {
	tools := make([]Tool, 0, len(s.tools))
	
	// Add all registered tools
	for name := range s.tools {
		tool := s.getToolDefinition(name)
		tools = append(tools, tool)
	}

	response := ListToolsResponse{
		Tools: tools,
	}

	return NewSuccessResponse(msg.ID, response), nil
}

// handleCallTool processes tool execution requests
func (s *Server) handleCallTool(msg *Message) (*Message, error) {
	var req CallToolRequest
	if err := s.parseParams(msg.Params, &req); err != nil {
		return NewErrorResponse(msg.ID, InvalidParams, "Invalid parameters", err.Error()), nil
	}

	handler, exists := s.tools[req.Name]
	if !exists {
		return NewErrorResponse(msg.ID, MethodNotFound, fmt.Sprintf("Tool not found: %s", req.Name), nil), nil
	}

	s.logger.WithFields(logrus.Fields{
		"tool": req.Name,
		"args": req.Arguments,
	}).Info("Executing tool")

	result, err := handler(s.ctx, req.Arguments)
	if err != nil {
		s.logger.WithError(err).Error("Tool execution failed")
		return NewErrorResponse(msg.ID, InternalError, "Tool execution failed", err.Error()), nil
	}

	return NewSuccessResponse(msg.ID, result), nil
}
// handleListResources processes resource listing requests
func (s *Server) handleListResources(msg *Message) (*Message, error) {
	resources := []Resource{
		{
			URI:         "players://{id}",
			Name:        "Player Details",
			Description: "Individual player information and rating details",
			MimeType:    "application/json",
		},
		{
			URI:         "clubs://{id}",
			Name:        "Club Details",
			Description: "Individual club information",
			MimeType:    "application/json",
		},
		{
			URI:         "clubs://{id}/profile",
			Name:        "Club Profile",
			Description: "Comprehensive club profile with members and statistics",
			MimeType:    "application/json",
		},
		{
			URI:         "tournaments://{id}",
			Name:        "Tournament Details",
			Description: "Individual tournament information",
			MimeType:    "application/json",
		},
		{
			URI:         "addresses://regions",
			Name:        "Available Regions",
			Description: "List of available regions for address lookups",
			MimeType:    "application/json",
		},
		{
			URI:         "addresses://{region}",
			Name:        "Regional Addresses",
			Description: "Chess official addresses by region",
			MimeType:    "application/json",
		},
		{
			URI:         "admin://health",
			Name:        "API Health Status",
			Description: "Portal64 API health and connectivity status",
			MimeType:    "application/json",
		},
		{
			URI:         "admin://cache",
			Name:        "Cache Statistics",
			Description: "API cache performance metrics",
			MimeType:    "application/json",
		},
	}

	response := ListResourcesResponse{
		Resources: resources,
	}

	return NewSuccessResponse(msg.ID, response), nil
}

// handleReadResource processes resource reading requests
func (s *Server) handleReadResource(msg *Message) (*Message, error) {
	var req ReadResourceRequest
	if err := s.parseParams(msg.Params, &req); err != nil {
		return NewErrorResponse(msg.ID, InvalidParams, "Invalid parameters", err.Error()), nil
	}

	s.logger.WithField("uri", req.URI).Info("Reading resource")

	// Parse URI and find appropriate handler
	parts := strings.SplitN(req.URI, "://", 2)
	if len(parts) != 2 {
		return NewErrorResponse(msg.ID, InvalidParams, "Invalid resource URI format", nil), nil
	}

	scheme := parts[0]
	path := parts[1]

	handler, exists := s.resources[scheme]
	if !exists {
		return NewErrorResponse(msg.ID, MethodNotFound, fmt.Sprintf("Resource scheme not found: %s", scheme), nil), nil
	}

	result, err := handler(s.ctx, path)
	if err != nil {
		s.logger.WithError(err).Error("Resource reading failed")
		return NewErrorResponse(msg.ID, InternalError, "Resource reading failed", err.Error()), nil
	}

	return NewSuccessResponse(msg.ID, result), nil
}

// parseParams parses message parameters into the provided struct
func (s *Server) parseParams(params interface{}, target interface{}) error {
	if params == nil {
		return fmt.Errorf("missing parameters")
	}

	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	return nil
}
