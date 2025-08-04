package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// registerResources registers all available MCP resources
func (s *Server) registerResources() {
	s.resources["players"] = s.handlePlayerResource
	s.resources["clubs"] = s.handleClubResource
	s.resources["tournaments"] = s.handleTournamentResource
	s.resources["addresses"] = s.handleAddressResource
	s.resources["admin"] = s.handleAdminResource
}

// handlePlayerResource handles player resource requests
func (s *Server) handlePlayerResource(ctx context.Context, path string) (*ReadResourceResponse, error) {
	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")
	
	if path == "" {
		return nil, fmt.Errorf("player ID is required")
	}

	// Extract player ID from path
	playerID := path
	
	// Get player profile
	player, err := s.apiClient.GetPlayerProfile(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player profile: %w", err)
	}

	// Serialize response
	data, err := json.MarshalIndent(player, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize player data: %w", err)
	}

	return &ReadResourceResponse{
		Contents: []ResourceContent{{
			URI:      fmt.Sprintf("players://%s", playerID),
			MimeType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// handleClubResource handles club resource requests
func (s *Server) handleClubResource(ctx context.Context, path string) (*ReadResourceResponse, error) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	
	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("club ID is required")
	}

	clubID := parts[0]

	// Check if this is a profile request
	if len(parts) > 1 && parts[1] == "profile" {
		profile, err := s.apiClient.GetClubProfile(ctx, clubID)
		if err != nil {
			return nil, fmt.Errorf("failed to get club profile: %w", err)
		}

		data, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize club profile data: %w", err)
		}

		return &ReadResourceResponse{
			Contents: []ResourceContent{{
				URI:      fmt.Sprintf("clubs://%s/profile", clubID),
				MimeType: "application/json",
				Text:     string(data),
			}},
		}, nil
	}

	// Get basic club information (this would need to be implemented in the API client)
	// For now, we'll return an error suggesting to use the profile endpoint
	return nil, fmt.Errorf("basic club information not available, use clubs://%s/profile instead", clubID)
}

// handleTournamentResource handles tournament resource requests
func (s *Server) handleTournamentResource(ctx context.Context, path string) (*ReadResourceResponse, error) {
	path = strings.TrimPrefix(path, "/")
	
	if path == "" {
		return nil, fmt.Errorf("tournament ID is required")
	}

	tournamentID := path
	
	// Get tournament details
	tournament, err := s.apiClient.GetTournamentDetails(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament details: %w", err)
	}

	data, err := json.MarshalIndent(tournament, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tournament data: %w", err)
	}

	return &ReadResourceResponse{
		Contents: []ResourceContent{{
			URI:      fmt.Sprintf("tournaments://%s", tournamentID),
			MimeType: "application/json",
			Text:     string(data),
		}},
	}, nil
}
// handleAddressResource handles address resource requests
func (s *Server) handleAddressResource(ctx context.Context, path string) (*ReadResourceResponse, error) {
	path = strings.TrimPrefix(path, "/")
	
	if path == "regions" {
		// Get list of regions
		regions, err := s.apiClient.GetRegions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get regions: %w", err)
		}

		data, err := json.MarshalIndent(regions, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize regions data: %w", err)
		}

		return &ReadResourceResponse{
			Contents: []ResourceContent{{
				URI:      "addresses://regions",
				MimeType: "application/json",
				Text:     string(data),
			}},
		}, nil
	}

	// Parse region and optional type
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("region is required")
	}

	region := parts[0]
	addressType := ""
	if len(parts) > 1 {
		addressType = parts[1]
	}

	// Get regional addresses
	addresses, err := s.apiClient.GetRegionAddresses(ctx, region, addressType)
	if err != nil {
		return nil, fmt.Errorf("failed to get region addresses: %w", err)
	}

	data, err := json.MarshalIndent(addresses, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize addresses data: %w", err)
	}

	uri := fmt.Sprintf("addresses://%s", region)
	if addressType != "" {
		uri = fmt.Sprintf("addresses://%s/%s", region, addressType)
	}

	return &ReadResourceResponse{
		Contents: []ResourceContent{{
			URI:      uri,
			MimeType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// handleAdminResource handles administrative resource requests
func (s *Server) handleAdminResource(ctx context.Context, path string) (*ReadResourceResponse, error) {
	path = strings.TrimPrefix(path, "/")
	
	switch path {
	case "health":
		health, err := s.apiClient.Health(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get health status: %w", err)
		}

		data, err := json.MarshalIndent(health, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize health data: %w", err)
		}

		return &ReadResourceResponse{
			Contents: []ResourceContent{{
				URI:      "admin://health",
				MimeType: "application/json",
				Text:     string(data),
			}},
		}, nil

	case "cache":
		stats, err := s.apiClient.CacheStats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get cache stats: %w", err)
		}

		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize cache stats: %w", err)
		}

		return &ReadResourceResponse{
			Contents: []ResourceContent{{
				URI:      "admin://cache",
				MimeType: "application/json",
				Text:     string(data),
			}},
		}, nil

	default:
		return nil, fmt.Errorf("unknown admin resource: %s", path)
	}
}
