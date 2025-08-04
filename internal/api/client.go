package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Client represents the Portal64 API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewClient creates a new Portal64 API client
func NewClient(baseURL string, timeout time.Duration, logger *logrus.Logger) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

// buildURL constructs API URLs with query parameters
func (c *Client) buildURL(endpoint string, params interface{}) string {
	u := c.baseURL + endpoint
	
	if params == nil {
		return u
	}

	values := url.Values{}
	
	switch p := params.(type) {
	case SearchParams:
		c.addSearchParams(&values, p)
	case DateRangeParams:
		c.addDateRangeParams(&values, p)
	case map[string]string:
		for k, v := range p {
			if v != "" {
				values.Set(k, v)
			}
		}
	}

	if len(values) > 0 {
		u += "?" + values.Encode()
	}

	return u
}

// addSearchParams adds search parameters to URL values
func (c *Client) addSearchParams(values *url.Values, params SearchParams) {
	if params.Query != "" {
		values.Set("query", params.Query)
	}
	if params.Limit > 0 {
		values.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		values.Set("offset", strconv.Itoa(params.Offset))
	}
	if params.SortBy != "" {
		values.Set("sort_by", params.SortBy)
	}
	if params.SortOrder != "" {
		values.Set("sort_order", params.SortOrder)
	}
	if params.FilterBy != "" {
		values.Set("filter_by", params.FilterBy)
	}
	if params.FilterValue != "" {
		values.Set("filter_value", params.FilterValue)
	}
	if params.Active != nil {
		values.Set("active", strconv.FormatBool(*params.Active))
	}
}
// addDateRangeParams adds date range parameters to URL values
func (c *Client) addDateRangeParams(values *url.Values, params DateRangeParams) {
	values.Set("start_date", params.StartDate.Format("2006-01-02"))
	values.Set("end_date", params.EndDate.Format("2006-01-02"))
	c.addSearchParams(values, params.SearchParams)
}

// doRequest performs HTTP request with error handling
func (c *Client) doRequest(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.WithFields(logrus.Fields{
		"method": method,
		"url":    url,
	}).Debug("Making API request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("API request failed")
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, c.handleErrorResponse(resp)
	}

	return resp, nil
}

// handleErrorResponse handles non-200 HTTP responses
func (c *Client) handleErrorResponse(resp *http.Response) error {
	var errorBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errorBody); err != nil {
		return fmt.Errorf("API error %d: failed to parse error response", resp.StatusCode)
	}

	if message, ok := errorBody["message"].(string); ok {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, message)
	}

	return fmt.Errorf("API error %d: %v", resp.StatusCode, errorBody)
}

// decodeResponse decodes JSON response into provided interface
func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		c.logger.WithError(err).Error("Failed to decode API response")
		return fmt.Errorf("response parsing failed: %w", err)
	}

	return nil
}

// Health checks API health status
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	url := c.buildURL("/api/v1/health", nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var health HealthResponse
	if err := c.decodeResponse(resp, &health); err != nil {
		return nil, err
	}

	return &health, nil
}

// CacheStats retrieves cache performance statistics
func (c *Client) CacheStats(ctx context.Context) (*CacheStatsResponse, error) {
	url := c.buildURL("/api/v1/admin/cache", nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var stats CacheStatsResponse
	if err := c.decodeResponse(resp, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}
// SearchPlayers searches for players with filtering and pagination
func (c *Client) SearchPlayers(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	url := c.buildURL("/api/v1/players", params)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := c.decodeResponse(resp, &searchResp); err != nil {
		return nil, err
	}

	// Convert data to []PlayerResponse
	if dataSlice, ok := searchResp.Data.([]interface{}); ok {
		players := make([]PlayerResponse, len(dataSlice))
		for i, item := range dataSlice {
			if playerData, ok := item.(map[string]interface{}); ok {
				// Convert map to PlayerResponse struct
				playerBytes, _ := json.Marshal(playerData)
				json.Unmarshal(playerBytes, &players[i])
			}
		}
		searchResp.Data = players
	}

	return &searchResp, nil
}

// GetPlayerProfile retrieves comprehensive player profile with rating history
func (c *Client) GetPlayerProfile(ctx context.Context, playerID string) (*PlayerResponse, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/players/%s", playerID), nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var player PlayerResponse
	if err := c.decodeResponse(resp, &player); err != nil {
		return nil, err
	}

	return &player, nil
}

// GetPlayerRatingHistory retrieves player's DWZ rating evolution over time
func (c *Client) GetPlayerRatingHistory(ctx context.Context, playerID string) ([]Evaluation, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/players/%s/history", playerID), nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var evaluations []Evaluation
	if err := c.decodeResponse(resp, &evaluations); err != nil {
		return nil, err
	}

	return evaluations, nil
}

// SearchClubs searches for clubs with filtering and pagination
func (c *Client) SearchClubs(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	url := c.buildURL("/api/v1/clubs", params)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := c.decodeResponse(resp, &searchResp); err != nil {
		return nil, err
	}

	// Convert data to []ClubResponse
	if dataSlice, ok := searchResp.Data.([]interface{}); ok {
		clubs := make([]ClubResponse, len(dataSlice))
		for i, item := range dataSlice {
			if clubData, ok := item.(map[string]interface{}); ok {
				clubBytes, _ := json.Marshal(clubData)
				json.Unmarshal(clubBytes, &clubs[i])
			}
		}
		searchResp.Data = clubs
	}

	return &searchResp, nil
}
// GetClubProfile retrieves comprehensive club profile with members and statistics
func (c *Client) GetClubProfile(ctx context.Context, clubID string) (*ClubProfileResponse, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/clubs/%s/profile", clubID), nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var profile ClubProfileResponse
	if err := c.decodeResponse(resp, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetClubPlayers retrieves club members with search and filtering
func (c *Client) GetClubPlayers(ctx context.Context, clubID string, params SearchParams) (*SearchResponse, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/clubs/%s/players", clubID), params)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := c.decodeResponse(resp, &searchResp); err != nil {
		return nil, err
	}

	// Convert data to []PlayerResponse
	if dataSlice, ok := searchResp.Data.([]interface{}); ok {
		players := make([]PlayerResponse, len(dataSlice))
		for i, item := range dataSlice {
			if playerData, ok := item.(map[string]interface{}); ok {
				playerBytes, _ := json.Marshal(playerData)
				json.Unmarshal(playerBytes, &players[i])
			}
		}
		searchResp.Data = players
	}

	return &searchResp, nil
}

// GetClubStatistics retrieves club performance statistics and member analytics
func (c *Client) GetClubStatistics(ctx context.Context, clubID string) (*ClubRatingStats, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/clubs/%s/statistics", clubID), nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var stats ClubRatingStats
	if err := c.decodeResponse(resp, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// SearchTournaments searches for tournaments with date and status filtering
func (c *Client) SearchTournaments(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	url := c.buildURL("/api/v1/tournaments", params)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := c.decodeResponse(resp, &searchResp); err != nil {
		return nil, err
	}

	// Convert data to []TournamentResponse
	if dataSlice, ok := searchResp.Data.([]interface{}); ok {
		tournaments := make([]TournamentResponse, len(dataSlice))
		for i, item := range dataSlice {
			if tournamentData, ok := item.(map[string]interface{}); ok {
				tournamentBytes, _ := json.Marshal(tournamentData)
				json.Unmarshal(tournamentBytes, &tournaments[i])
			}
		}
		searchResp.Data = tournaments
	}

	return &searchResp, nil
}
// SearchTournamentsByDate searches tournaments by date range
func (c *Client) SearchTournamentsByDate(ctx context.Context, params DateRangeParams) (*SearchResponse, error) {
	url := c.buildURL("/api/v1/tournaments/search", params)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := c.decodeResponse(resp, &searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}

// GetRecentTournaments retrieves recent tournaments
func (c *Client) GetRecentTournaments(ctx context.Context, days, limit int) ([]TournamentResponse, error) {
	params := map[string]string{
		"days":  strconv.Itoa(days),
		"limit": strconv.Itoa(limit),
	}
	url := c.buildURL("/api/v1/tournaments/recent", params)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var tournaments []TournamentResponse
	if err := c.decodeResponse(resp, &tournaments); err != nil {
		return nil, err
	}

	return tournaments, nil
}

// GetTournamentDetails retrieves detailed tournament information
func (c *Client) GetTournamentDetails(ctx context.Context, tournamentID string) (*EnhancedTournamentResponse, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/tournaments/%s", tournamentID), nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var tournament EnhancedTournamentResponse
	if err := c.decodeResponse(resp, &tournament); err != nil {
		return nil, err
	}

	return &tournament, nil
}

// GetRegions retrieves available regions for address lookups
func (c *Client) GetRegions(ctx context.Context) ([]RegionInfo, error) {
	url := c.buildURL("/api/v1/addresses/regions", nil)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var regions []RegionInfo
	if err := c.decodeResponse(resp, &regions); err != nil {
		return nil, err
	}

	return regions, nil
}

// GetRegionAddresses retrieves addresses for chess officials by region
func (c *Client) GetRegionAddresses(ctx context.Context, region, addressType string) ([]RegionAddressResponse, error) {
	params := map[string]string{}
	if addressType != "" {
		params["type"] = addressType
	}
	
	url := c.buildURL(fmt.Sprintf("/api/v1/addresses/%s", region), params)
	
	resp, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		return nil, err
	}

	var addresses []RegionAddressResponse
	if err := c.decodeResponse(resp, &addresses); err != nil {
		return nil, err
	}

	return addresses, nil
}
