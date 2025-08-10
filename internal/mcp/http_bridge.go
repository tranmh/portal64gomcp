package mcp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/svw-info/portal64gomcp/internal/logger"
	"github.com/svw-info/portal64gomcp/internal/ssl"
)

// HTTPBridge provides HTTP access to MCP functionality
type HTTPBridge struct {
	server *Server
	logger logger.Logger
}

// NewHTTPBridge creates a new HTTP bridge for MCP server
func NewHTTPBridge(server *Server, logger logger.Logger) *HTTPBridge {
	return &HTTPBridge{
		server: server,
		logger: logger,
	}
}

// SetupRoutes configures HTTP routes for MCP functionality
func (h *HTTPBridge) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Add security middleware
	r.Use(h.securityMiddleware)
	r.Use(h.corsMiddleware)
	r.Use(h.loggingMiddleware)
	r.Use(h.clientCertMiddleware) // mTLS support

	// Health endpoints
	r.HandleFunc("/health", h.handleHealth).Methods("GET")
	r.HandleFunc("/api/v1/health", h.handleHealth).Methods("GET")
	
	// Admin endpoints
	r.HandleFunc("/api/v1/admin/cache", h.handleCacheStats).Methods("GET")

	// SSL info endpoint
	if h.server.config.MCP.SSL.Enabled {
		r.HandleFunc("/api/v1/ssl/info", h.handleSSLInfo).Methods("GET")
	}

	// MCP protocol endpoints
	r.HandleFunc("/tools/list", h.handleListTools).Methods("POST", "GET")
	r.HandleFunc("/tools/call", h.handleCallTool).Methods("POST")
	r.HandleFunc("/resources/list", h.handleListResources).Methods("POST", "GET")
	r.HandleFunc("/resources/read", h.handleReadResource).Methods("POST")

	// Player endpoints (both versioned and non-versioned)
	r.HandleFunc("/api/v1/players", h.handleSearchPlayers).Methods("GET")
	r.HandleFunc("/api/players/", h.handleSearchPlayers).Methods("GET")
	r.HandleFunc("/api/v1/players/{id}", h.handleGetPlayerProfile).Methods("GET")
	r.HandleFunc("/api/players/{id}", h.handleGetPlayerProfile).Methods("GET")
	r.HandleFunc("/api/v1/players/{id}/history", h.handleGetPlayerRatingHistory).Methods("GET")

	// Club endpoints (both versioned and non-versioned)
	r.HandleFunc("/api/v1/clubs", h.handleSearchClubs).Methods("GET")
	r.HandleFunc("/api/clubs/", h.handleSearchClubs).Methods("GET")
	r.HandleFunc("/api/v1/clubs/{id}", h.handleGetClubProfile).Methods("GET")
	r.HandleFunc("/api/clubs/{id}", h.handleGetClubProfile).Methods("GET")
	r.HandleFunc("/api/v1/clubs/{id}/profile", h.handleGetClubProfile).Methods("GET")
	r.HandleFunc("/api/v1/clubs/{id}/players", h.handleGetClubPlayers).Methods("GET")
	r.HandleFunc("/api/v1/clubs/{id}/statistics", h.handleGetClubStatistics).Methods("GET")

	// Tournament endpoints (both versioned and non-versioned)
	r.HandleFunc("/api/v1/tournaments", h.handleSearchTournaments).Methods("GET")
	r.HandleFunc("/api/tournaments/", h.handleSearchTournaments).Methods("GET")
	r.HandleFunc("/api/v1/tournaments/search", h.handleSearchTournamentsByDate).Methods("GET")
	r.HandleFunc("/api/v1/tournaments/recent", h.handleGetRecentTournaments).Methods("GET")
	r.HandleFunc("/api/v1/tournaments/{id}", h.handleGetTournamentDetails).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}", h.handleGetTournamentDetails).Methods("GET")

	// Region endpoints
	r.HandleFunc("/api/v1/addresses/regions", h.handleGetRegions).Methods("GET")
	r.HandleFunc("/api/v1/addresses/{region}", h.handleGetRegionAddresses).Methods("GET")

	return r
}

// corsMiddleware adds CORS headers
func (h *HTTPBridge) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (h *HTTPBridge) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		h.logger.WithFields(map[string]interface{}{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(start),
		}).Info("HTTP request processed")
	})
}

// Helper function to write JSON responses
func (h *HTTPBridge) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// Helper function to write error responses
func (h *HTTPBridge) writeErrorResponse(w http.ResponseWriter, statusCode int, message, code string) {
	h.writeJSONResponse(w, statusCode, map[string]interface{}{
		"message": message,
		"code":    code,
	})
}

// Health endpoint handler
func (h *HTTPBridge) handleHealth(w http.ResponseWriter, r *http.Request) {
	result, err := h.callMCPTool(r.Context(), "check_api_health", map[string]interface{}{})
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Health check failed", "HEALTH_CHECK_FAILED")
		return
	}

	// Extract health data from MCP response
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// If we have actual health data from the MCP tool, use it
	if result != nil && len(result.Content) > 0 {
		textContent := result.Content[0].Text
		if textContent != "" {
			var healthData map[string]interface{}
			if err := json.Unmarshal([]byte(textContent), &healthData); err == nil {
				health = healthData
			}
		}
	}

	h.writeJSONResponse(w, http.StatusOK, health)
}

// Cache stats endpoint handler
func (h *HTTPBridge) handleCacheStats(w http.ResponseWriter, r *http.Request) {
	result, err := h.callMCPTool(r.Context(), "get_cache_stats", map[string]interface{}{})
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get cache stats", "CACHE_STATS_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// MCP Protocol handlers

// handleListTools handles tool listing requests
func (h *HTTPBridge) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools := make([]Tool, 0, len(h.server.tools))
	
	// Add all registered tools
	for name := range h.server.tools {
		tool := h.server.GetToolDefinition(name)
		tools = append(tools, tool)
	}

	response := ListToolsResponse{
		Tools: tools,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// handleCallTool handles tool execution requests
func (h *HTTPBridge) handleCallTool(w http.ResponseWriter, r *http.Request) {
	var req CallToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	result, err := h.callMCPTool(r.Context(), req.Name, req.Arguments)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Tool execution failed: %v", err), "TOOL_EXECUTION_FAILED")
		return
	}

	// For MCP tool calls, return the raw MCP response format
	h.writeJSONResponse(w, http.StatusOK, result)
}

// handleListResources handles resource listing requests
func (h *HTTPBridge) handleListResources(w http.ResponseWriter, r *http.Request) {
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

	h.writeJSONResponse(w, http.StatusOK, response)
}

// handleReadResource handles resource reading requests
func (h *HTTPBridge) handleReadResource(w http.ResponseWriter, r *http.Request) {
	var req ReadResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Parse URI and find appropriate handler
	parts := strings.SplitN(req.URI, "://", 2)
	if len(parts) != 2 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid resource URI format", "INVALID_URI")
		return
	}

	scheme := parts[0]
	path := parts[1]

	handler, exists := h.server.resources[scheme]
	if !exists {
		h.writeErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Resource scheme not found: %s", scheme), "RESOURCE_NOT_FOUND")
		return
	}

	result, err := handler(r.Context(), path)
	if err != nil {
		h.logger.WithError(err).Error("Resource reading failed")
		h.writeErrorResponse(w, http.StatusInternalServerError, "Resource reading failed", "RESOURCE_READ_FAILED")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// Player handlers

// handleSearchPlayers handles player search requests
func (h *HTTPBridge) handleSearchPlayers(w http.ResponseWriter, r *http.Request) {
	params := h.parseSearchParams(r)
	
	result, err := h.callMCPTool(r.Context(), "search_players", map[string]interface{}{
		"query":        params["query"],
		"limit":        params["limit"],
		"offset":       params["offset"],
		"sort_by":      params["sort_by"],
		"sort_order":   params["sort_order"],
		"filter_by":    params["filter_by"],
		"filter_value": params["filter_value"],
		"active":       params["active"],
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Player search failed", "PLAYER_SEARCH_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetPlayerProfile handles player profile requests
func (h *HTTPBridge) handleGetPlayerProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID := vars["id"]

	result, err := h.callMCPTool(r.Context(), "get_player_profile", map[string]interface{}{
		"player_id": playerID,
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Player profile retrieval failed", "PLAYER_PROFILE_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetPlayerRatingHistory handles player rating history requests
func (h *HTTPBridge) handleGetPlayerRatingHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID := vars["id"]

	result, err := h.callMCPTool(r.Context(), "get_player_rating_history", map[string]interface{}{
		"player_id": playerID,
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Player rating history retrieval failed", "PLAYER_HISTORY_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// Club handlers

// handleSearchClubs handles club search requests
func (h *HTTPBridge) handleSearchClubs(w http.ResponseWriter, r *http.Request) {
	params := h.parseSearchParams(r)
	
	result, err := h.callMCPTool(r.Context(), "search_clubs", map[string]interface{}{
		"query":        params["query"],
		"limit":        params["limit"],
		"offset":       params["offset"],
		"sort_by":      params["sort_by"],
		"sort_order":   params["sort_order"],
		"filter_by":    params["filter_by"],
		"filter_value": params["filter_value"],
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Club search failed", "CLUB_SEARCH_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetClubProfile handles club profile requests
func (h *HTTPBridge) handleGetClubProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clubID := vars["id"]

	result, err := h.callMCPTool(r.Context(), "get_club_profile", map[string]interface{}{
		"club_id": clubID,
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Club profile retrieval failed", "CLUB_PROFILE_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetClubPlayers handles club players requests
func (h *HTTPBridge) handleGetClubPlayers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clubID := vars["id"]
	params := h.parseSearchParams(r)

	result, err := h.callMCPTool(r.Context(), "get_club_players", map[string]interface{}{
		"club_id":      clubID,
		"query":        params["query"],
		"limit":        params["limit"],
		"offset":       params["offset"],
		"sort_by":      params["sort_by"],
		"sort_order":   params["sort_order"],
		"filter_by":    params["filter_by"],
		"filter_value": params["filter_value"],
		"active":       params["active"],
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Club players retrieval failed", "CLUB_PLAYERS_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetClubStatistics handles club statistics requests
func (h *HTTPBridge) handleGetClubStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clubID := vars["id"]

	result, err := h.callMCPTool(r.Context(), "get_club_statistics", map[string]interface{}{
		"club_id": clubID,
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Club statistics retrieval failed", "CLUB_STATISTICS_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// Tournament handlers

// handleSearchTournaments handles tournament search requests
func (h *HTTPBridge) handleSearchTournaments(w http.ResponseWriter, r *http.Request) {
	params := h.parseSearchParams(r)
	
	result, err := h.callMCPTool(r.Context(), "search_tournaments", map[string]interface{}{
		"query":        params["query"],
		"limit":        params["limit"],
		"offset":       params["offset"],
		"sort_by":      params["sort_by"],
		"sort_order":   params["sort_order"],
		"filter_by":    params["filter_by"],
		"filter_value": params["filter_value"],
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Tournament search failed", "TOURNAMENT_SEARCH_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleSearchTournamentsByDate handles tournament search by date range requests
func (h *HTTPBridge) handleSearchTournamentsByDate(w http.ResponseWriter, r *http.Request) {
	params := h.parseSearchParams(r)
	
	// Parse dates
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	
	result, err := h.callMCPTool(r.Context(), "search_tournaments_by_date", map[string]interface{}{
		"start_date":   startDate,
		"end_date":     endDate,
		"query":        params["query"],
		"limit":        params["limit"],
		"offset":       params["offset"],
		"sort_by":      params["sort_by"],
		"sort_order":   params["sort_order"],
		"filter_by":    params["filter_by"],
		"filter_value": params["filter_value"],
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Tournament date search failed", "TOURNAMENT_DATE_SEARCH_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetRecentTournaments handles recent tournaments requests
func (h *HTTPBridge) handleGetRecentTournaments(w http.ResponseWriter, r *http.Request) {
	days := 30 // default
	limit := 25 // default
	
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	result, err := h.callMCPTool(r.Context(), "get_recent_tournaments", map[string]interface{}{
		"days":  days,
		"limit": limit,
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Recent tournaments retrieval failed", "RECENT_TOURNAMENTS_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetTournamentDetails handles tournament details requests
func (h *HTTPBridge) handleGetTournamentDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	result, err := h.callMCPTool(r.Context(), "get_tournament_details", map[string]interface{}{
		"tournament_id": tournamentID,
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Tournament details retrieval failed", "TOURNAMENT_DETAILS_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// Region handlers

// handleGetRegions handles regions requests
func (h *HTTPBridge) handleGetRegions(w http.ResponseWriter, r *http.Request) {
	result, err := h.callMCPTool(r.Context(), "get_regions", map[string]interface{}{})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Regions retrieval failed", "REGIONS_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// handleGetRegionAddresses handles region addresses requests
func (h *HTTPBridge) handleGetRegionAddresses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	region := vars["region"]
	
	addressType := r.URL.Query().Get("type")

	result, err := h.callMCPTool(r.Context(), "get_region_addresses", map[string]interface{}{
		"region":       region,
		"address_type": addressType,
	})
	
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Region addresses retrieval failed", "REGION_ADDRESSES_FAILED")
		return
	}

	h.writeMCPToolResponse(w, result)
}

// Helper functions

// parseSearchParams parses search parameters from HTTP request
func (h *HTTPBridge) parseSearchParams(r *http.Request) map[string]interface{} {
	params := make(map[string]interface{})
	
	query := r.URL.Query()
	
	if q := query.Get("query"); q != "" {
		params["query"] = q
	}
	
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params["limit"] = limit
		}
	}
	
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params["offset"] = offset
		}
	}
	
	if sortBy := query.Get("sort_by"); sortBy != "" {
		params["sort_by"] = sortBy
	}
	
	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		params["sort_order"] = sortOrder
	}
	
	if filterBy := query.Get("filter_by"); filterBy != "" {
		params["filter_by"] = filterBy
	}
	
	if filterValue := query.Get("filter_value"); filterValue != "" {
		params["filter_value"] = filterValue
	}
	
	if activeStr := query.Get("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			params["active"] = active
		}
	}
	
	return params
}

// callMCPTool calls an MCP tool and returns the result
func (h *HTTPBridge) callMCPTool(ctx context.Context, toolName string, args map[string]interface{}) (*CallToolResponse, error) {
	handler, exists := h.server.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	h.logger.WithFields(map[string]interface{}{
		"tool": toolName,
		"args": args,
	}).Debug("Executing tool via HTTP bridge")

	result, err := handler(ctx, args)
	if err != nil {
		h.logger.WithError(err).Error("Tool execution failed via HTTP bridge")
		return nil, err
	}

	return result, nil
}

// writeMCPToolResponse writes an MCP tool response as HTTP JSON
func (h *HTTPBridge) writeMCPToolResponse(w http.ResponseWriter, result *CallToolResponse) {
	if result == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "No response from tool", "NO_RESPONSE")
		return
	}

	if result.IsError {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Tool execution error", "TOOL_ERROR")
		return
	}

	// If we have content, try to parse it as JSON
	if len(result.Content) > 0 {
		// Handle text content
		textContent := result.Content[0].Text
		if textContent != "" {
			// Try to parse as JSON first
			var jsonData interface{}
			if err := json.Unmarshal([]byte(textContent), &jsonData); err == nil {
				h.writeJSONResponse(w, http.StatusOK, jsonData)
				return
			}
			// If not JSON, return as text response
			h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
				"data": textContent,
			})
			return
		}
		
		// Handle direct data content
		if result.Content[0].Data != nil {
			h.writeJSONResponse(w, http.StatusOK, result.Content[0].Data)
			return
		}
	}

	// Fallback: return the raw MCP response
	h.writeJSONResponse(w, http.StatusOK, result)
}


// securityMiddleware adds security headers
func (h *HTTPBridge) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		secureHeaders := ssl.SecureHeaders(h.server.config.MCP.SSL.HSTSMaxAge)
		for key, value := range secureHeaders {
			w.Header().Set(key, value)
		}

		// Add additional SSL-specific headers if HTTPS
		if r.TLS != nil {
			w.Header().Set("X-SSL-Enabled", "true")
			w.Header().Set("X-TLS-Version", h.server.formatTLSVersion(r.TLS.Version))
			
			if len(r.TLS.PeerCertificates) > 0 {
				w.Header().Set("X-Client-Cert-Present", "true")
				// Add client cert subject if available
				if r.TLS.PeerCertificates[0].Subject.CommonName != "" {
					w.Header().Set("X-Client-Cert-CN", r.TLS.PeerCertificates[0].Subject.CommonName)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// clientCertMiddleware handles client certificate authentication
func (h *HTTPBridge) clientCertMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply client cert logic if SSL is enabled and configured
		if !h.server.config.MCP.SSL.Enabled || !h.server.config.MCP.SSL.RequireClientCert {
			next.ServeHTTP(w, r)
			return
		}

		// Check if client certificate is present
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			h.logger.WithFields(map[string]interface{}{
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
			}).Warn("Client certificate required but not provided")
			
			http.Error(w, "Client certificate required", http.StatusUnauthorized)
			return
		}

		// Log client certificate information
		cert := r.TLS.PeerCertificates[0]
		h.logger.WithFields(map[string]interface{}{
			"client_cert_cn":     cert.Subject.CommonName,
			"client_cert_org":    strings.Join(cert.Subject.Organization, ","),
			"client_cert_serial": cert.SerialNumber.String(),
			"path":               r.URL.Path,
		}).Info("Client certificate authenticated")

		next.ServeHTTP(w, r)
	})
}

// Enhanced corsMiddleware handles CORS with SSL considerations
func (h *HTTPBridge) enhancedCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS for development, but be more restrictive in production
		if !h.server.config.MCP.SSL.Enabled {
			// Development mode - more permissive CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Production mode - restrict origins
			origin := r.Header.Get("Origin")
			allowedOrigins := []string{
				"https://localhost:" + strconv.Itoa(h.server.config.MCP.HTTPPort),
				"https://127.0.0.1:" + strconv.Itoa(h.server.config.MCP.HTTPPort),
			}
			
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Client-Cert")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Enhanced loggingMiddleware logs HTTP requests with SSL information
func (h *HTTPBridge) enhancedLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Log request details
		duration := time.Since(start)
		
		logFields := map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status":      wrapped.statusCode,
			"duration_ms": duration.Milliseconds(),
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}

		// Add SSL information if available
		if r.TLS != nil {
			logFields["tls_version"] = h.server.formatTLSVersion(r.TLS.Version)
			logFields["cipher_suite"] = tls.CipherSuiteName(r.TLS.CipherSuite)
			
			if len(r.TLS.PeerCertificates) > 0 {
				cert := r.TLS.PeerCertificates[0]
				logFields["client_cert_cn"] = cert.Subject.CommonName
				logFields["client_cert_org"] = strings.Join(cert.Subject.Organization, ",")
			}
		}

		// Choose log level based on status code
		if wrapped.statusCode >= 500 {
			h.logger.WithFields(logFields).Error("HTTP request completed with server error")
		} else if wrapped.statusCode >= 400 {
			h.logger.WithFields(logFields).Warn("HTTP request completed with client error")
		} else {
			h.logger.WithFields(logFields).Info("HTTP request completed")
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleSSLInfo provides SSL configuration information
func (h *HTTPBridge) handleSSLInfo(w http.ResponseWriter, r *http.Request) {
	if !h.server.config.MCP.SSL.Enabled {
		http.Error(w, "SSL is not enabled", http.StatusNotFound)
		return
	}

	info := map[string]interface{}{
		"ssl_enabled":           true,
		"min_tls_version":       h.server.config.MCP.SSL.MinVersion,
		"max_tls_version":       h.server.config.MCP.SSL.MaxVersion,
		"require_client_cert":   h.server.config.MCP.SSL.RequireClientCert,
		"auto_generate_certs":   h.server.config.MCP.SSL.AutoGenerateCerts,
		"hsts_max_age":          h.server.config.MCP.SSL.HSTSMaxAge,
		"cert_file":             h.server.config.MCP.SSL.CertFile,
	}

	// Add connection-specific SSL information
	if r.TLS != nil {
		info["connection"] = map[string]interface{}{
			"tls_version":           h.server.formatTLSVersion(r.TLS.Version),
			"cipher_suite":          tls.CipherSuiteName(r.TLS.CipherSuite),
			"server_name":           r.TLS.ServerName,
			"client_cert_present":   len(r.TLS.PeerCertificates) > 0,
		}

		if len(r.TLS.PeerCertificates) > 0 {
			cert := r.TLS.PeerCertificates[0]
			info["connection"].(map[string]interface{})["client_cert"] = map[string]interface{}{
				"subject":      cert.Subject.String(),
				"issuer":       cert.Issuer.String(),
				"serial":       cert.SerialNumber.String(),
				"not_before":   cert.NotBefore.Format(time.RFC3339),
				"not_after":    cert.NotAfter.Format(time.RFC3339),
				"dns_names":    cert.DNSNames,
				"ip_addresses": cert.IPAddresses,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// Enhanced handleHealth with SSL information
func (h *HTTPBridge) handleEnhancedHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0", // This could be injected from build info
		"ssl": map[string]interface{}{
			"enabled": h.server.config.MCP.SSL.Enabled,
		},
	}

	// Add SSL-specific health information
	if h.server.config.MCP.SSL.Enabled && r.TLS != nil {
		health["ssl"].(map[string]interface{})["tls_version"] = h.server.formatTLSVersion(r.TLS.Version)
		health["ssl"].(map[string]interface{})["cipher_suite"] = tls.CipherSuiteName(r.TLS.CipherSuite)
		health["ssl"].(map[string]interface{})["client_cert_present"] = len(r.TLS.PeerCertificates) > 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
