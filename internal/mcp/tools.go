package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/svw-info/portal64gomcp/internal/api"
)

// registerTools registers all available MCP tools
func (s *Server) registerTools() {
	// Search tools
	s.tools["search_players"] = s.handleSearchPlayers
	s.tools["search_clubs"] = s.handleSearchClubs
	s.tools["search_tournaments"] = s.handleSearchTournaments
	s.tools["get_recent_tournaments"] = s.handleGetRecentTournaments
	s.tools["search_tournaments_by_date"] = s.handleSearchTournamentsByDate

	// Detail tools
	s.tools["get_player_profile"] = s.handleGetPlayerProfile
	s.tools["get_club_profile"] = s.handleGetClubProfile
	s.tools["get_tournament_details"] = s.handleGetTournamentDetails
	s.tools["get_club_players"] = s.handleGetClubPlayers

	// Analysis tools
	s.tools["get_player_rating_history"] = s.handleGetPlayerRatingHistory
	s.tools["get_club_statistics"] = s.handleGetClubStatistics

	// Administrative tools
	s.tools["check_api_health"] = s.handleCheckAPIHealth
	s.tools["get_cache_stats"] = s.handleGetCacheStats
	s.tools["get_regions"] = s.handleGetRegions
	s.tools["get_region_addresses"] = s.handleGetRegionAddresses
}

// getToolDefinition returns the schema definition for a tool
func (s *Server) getToolDefinition(name string) Tool {
	definitions := map[string]Tool{
		"search_players": {
			Name:        "search_players",
			Description: "Search for players with filtering and pagination support",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for player name",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results (default: 50)",
						"minimum":     1,
						"maximum":     200,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to skip (default: 0)",
						"minimum":     0,
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Field to sort by",
						"enum":        []string{"name", "current_dwz", "club"},
					},
					"sort_order": map[string]interface{}{
						"type":        "string",
						"description": "Sort order",
						"enum":        []string{"asc", "desc"},
					},
					"active": map[string]interface{}{
						"type":        "boolean",
						"description": "Filter for active players only",
					},
				},
			},
		},
		"search_clubs": {
			Name:        "search_clubs",
			Description: "Search for clubs with geographic and membership filtering",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for club name",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results (default: 50)",
						"minimum":     1,
						"maximum":     200,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to skip (default: 0)",
						"minimum":     0,
					},
					"sort_by": map[string]interface{}{
						"type":        "string",
						"description": "Field to sort by",
						"enum":        []string{"name", "member_count", "city"},
					},
					"sort_order": map[string]interface{}{
						"type":        "string",
						"description": "Sort order",
						"enum":        []string{"asc", "desc"},
					},
					"filter_by": map[string]interface{}{
						"type":        "string",
						"description": "Field to filter by",
						"enum":        []string{"region", "state", "city"},
					},
					"filter_value": map[string]interface{}{
						"type":        "string",
						"description": "Value to filter by when filter_by is specified",
					},
				},
			},
		},
	}

	if def, exists := definitions[name]; exists {
		return def
	}

	// Return a generic definition for tools not explicitly defined
	return Tool{
		Name:        name,
		Description: fmt.Sprintf("Execute %s operation", name),
		InputSchema: ToolSchema{Type: "object"},
	}
}
// handleSearchPlayers handles player search requests
func (s *Server) handleSearchPlayers(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	// Parse arguments
	params := api.SearchParams{}
	
	if query, ok := args["query"].(string); ok {
		params.Query = query
	}
	if limit, ok := args["limit"].(float64); ok {
		params.Limit = int(limit)
	} else {
		params.Limit = 50 // default
	}
	if offset, ok := args["offset"].(float64); ok {
		params.Offset = int(offset)
	}
	if sortBy, ok := args["sort_by"].(string); ok {
		params.SortBy = sortBy
	}
	if sortOrder, ok := args["sort_order"].(string); ok {
		params.SortOrder = sortOrder
	}
	if active, ok := args["active"].(bool); ok {
		params.Active = &active
	}

	// Call API
	result, err := s.apiClient.SearchPlayers(ctx, params)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error searching players: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Format response
	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleSearchClubs handles club search requests
func (s *Server) handleSearchClubs(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	params := api.SearchParams{}
	
	if query, ok := args["query"].(string); ok {
		params.Query = query
	}
	if limit, ok := args["limit"].(float64); ok {
		params.Limit = int(limit)
	} else {
		params.Limit = 50
	}
	if offset, ok := args["offset"].(float64); ok {
		params.Offset = int(offset)
	}
	if sortBy, ok := args["sort_by"].(string); ok {
		params.SortBy = sortBy
	}
	if sortOrder, ok := args["sort_order"].(string); ok {
		params.SortOrder = sortOrder
	}
	if filterBy, ok := args["filter_by"].(string); ok {
		params.FilterBy = filterBy
	}
	if filterValue, ok := args["filter_value"].(string); ok {
		params.FilterValue = filterValue
	}

	result, err := s.apiClient.SearchClubs(ctx, params)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error searching clubs: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetPlayerProfile handles player profile requests
func (s *Server) handleGetPlayerProfile(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	playerID, ok := args["player_id"].(string)
	if !ok || playerID == "" {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: player_id is required",
			}},
			IsError: true,
		}, nil
	}

	result, err := s.apiClient.GetPlayerProfile(ctx, playerID)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting player profile: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}
// handleSearchTournaments handles tournament search requests
func (s *Server) handleSearchTournaments(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	params := api.SearchParams{}
	
	if query, ok := args["query"].(string); ok {
		params.Query = query
	}
	if limit, ok := args["limit"].(float64); ok {
		params.Limit = int(limit)
	} else {
		params.Limit = 50
	}
	if offset, ok := args["offset"].(float64); ok {
		params.Offset = int(offset)
	}
	if sortBy, ok := args["sort_by"].(string); ok {
		params.SortBy = sortBy
	}
	if sortOrder, ok := args["sort_order"].(string); ok {
		params.SortOrder = sortOrder
	}
	if filterBy, ok := args["filter_by"].(string); ok {
		params.FilterBy = filterBy
	}
	if filterValue, ok := args["filter_value"].(string); ok {
		params.FilterValue = filterValue
	}

	result, err := s.apiClient.SearchTournaments(ctx, params)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error searching tournaments: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetRecentTournaments handles recent tournament requests
func (s *Server) handleGetRecentTournaments(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	days := 30 // default
	if d, ok := args["days"].(float64); ok {
		days = int(d)
	}
	
	limit := 50 // default
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	result, err := s.apiClient.GetRecentTournaments(ctx, days, limit)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting recent tournaments: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleSearchTournamentsByDate handles tournament search by date range
func (s *Server) handleSearchTournamentsByDate(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	startDateStr, ok1 := args["start_date"].(string)
	endDateStr, ok2 := args["end_date"].(string)
	
	if !ok1 || !ok2 {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: start_date and end_date are required (format: YYYY-MM-DD)",
			}},
			IsError: true,
		}, nil
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: invalid start_date format (use YYYY-MM-DD)",
			}},
			IsError: true,
		}, nil
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: invalid end_date format (use YYYY-MM-DD)",
			}},
			IsError: true,
		}, nil
	}

	params := api.DateRangeParams{
		StartDate: startDate,
		EndDate:   endDate,
		SearchParams: api.SearchParams{
			Limit: 50,
		},
	}

	if query, ok := args["query"].(string); ok {
		params.SearchParams.Query = query
	}
	if limit, ok := args["limit"].(float64); ok {
		params.SearchParams.Limit = int(limit)
	}
	if offset, ok := args["offset"].(float64); ok {
		params.SearchParams.Offset = int(offset)
	}

	result, err := s.apiClient.SearchTournamentsByDate(ctx, params)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error searching tournaments by date: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}
// handleGetClubProfile handles club profile requests
func (s *Server) handleGetClubProfile(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	clubID, ok := args["club_id"].(string)
	if !ok || clubID == "" {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: club_id is required",
			}},
			IsError: true,
		}, nil
	}

	result, err := s.apiClient.GetClubProfile(ctx, clubID)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting club profile: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetTournamentDetails handles tournament details requests
func (s *Server) handleGetTournamentDetails(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	tournamentID, ok := args["tournament_id"].(string)
	if !ok || tournamentID == "" {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: tournament_id is required",
			}},
			IsError: true,
		}, nil
	}

	result, err := s.apiClient.GetTournamentDetails(ctx, tournamentID)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting tournament details: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetClubPlayers handles club players requests
func (s *Server) handleGetClubPlayers(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	clubID, ok := args["club_id"].(string)
	if !ok || clubID == "" {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: club_id is required",
			}},
			IsError: true,
		}, nil
	}

	params := api.SearchParams{Limit: 50}
	if query, ok := args["query"].(string); ok {
		params.Query = query
	}
	if limit, ok := args["limit"].(float64); ok {
		params.Limit = int(limit)
	}
	if offset, ok := args["offset"].(float64); ok {
		params.Offset = int(offset)
	}
	if sortBy, ok := args["sort_by"].(string); ok {
		params.SortBy = sortBy
	}
	if active, ok := args["active"].(bool); ok {
		params.Active = &active
	}

	result, err := s.apiClient.GetClubPlayers(ctx, clubID, params)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting club players: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetPlayerRatingHistory handles player rating history requests
func (s *Server) handleGetPlayerRatingHistory(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	playerID, ok := args["player_id"].(string)
	if !ok || playerID == "" {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: player_id is required",
			}},
			IsError: true,
		}, nil
	}

	result, err := s.apiClient.GetPlayerRatingHistory(ctx, playerID)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting player rating history: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}
// handleGetClubStatistics handles club statistics requests
func (s *Server) handleGetClubStatistics(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	clubID, ok := args["club_id"].(string)
	if !ok || clubID == "" {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: club_id is required",
			}},
			IsError: true,
		}, nil
	}

	result, err := s.apiClient.GetClubStatistics(ctx, clubID)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting club statistics: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleCheckAPIHealth handles API health check requests
func (s *Server) handleCheckAPIHealth(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	result, err := s.apiClient.Health(ctx)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error checking API health: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetCacheStats handles cache statistics requests
func (s *Server) handleGetCacheStats(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	result, err := s.apiClient.CacheStats(ctx)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting cache stats: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetRegions handles region listing requests
func (s *Server) handleGetRegions(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	result, err := s.apiClient.GetRegions(ctx)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting regions: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}

// handleGetRegionAddresses handles region address requests
func (s *Server) handleGetRegionAddresses(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	region, ok := args["region"].(string)
	if !ok || region == "" {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: "Error: region is required",
			}},
			IsError: true,
		}, nil
	}

	addressType := ""
	if t, ok := args["type"].(string); ok {
		addressType = t
	}

	result, err := s.apiClient.GetRegionAddresses(ctx, region, addressType)
	if err != nil {
		return &CallToolResponse{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error getting region addresses: %v", err),
			}},
			IsError: true,
		}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return &CallToolResponse{
		Content: []ToolContent{{
			Type: "text",
			Text: string(data),
		}},
	}, nil
}
