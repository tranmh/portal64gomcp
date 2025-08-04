package portal64

import (
	"context"
	"time"

	"github.com/svw-info/portal64gomcp/internal/api"
)

// Client provides a public interface to the Portal64 API
type Client interface {
	// Player operations
	SearchPlayers(ctx context.Context, params SearchParams) (*SearchResponse, error)
	GetPlayerProfile(ctx context.Context, playerID string) (*PlayerResponse, error)
	GetPlayerRatingHistory(ctx context.Context, playerID string) ([]Evaluation, error)

	// Club operations
	SearchClubs(ctx context.Context, params SearchParams) (*SearchResponse, error)
	GetClubProfile(ctx context.Context, clubID string) (*ClubProfileResponse, error)
	GetClubPlayers(ctx context.Context, clubID string, params SearchParams) (*SearchResponse, error)
	GetClubStatistics(ctx context.Context, clubID string) (*ClubRatingStats, error)

	// Tournament operations
	SearchTournaments(ctx context.Context, params SearchParams) (*SearchResponse, error)
	SearchTournamentsByDate(ctx context.Context, params DateRangeParams) (*SearchResponse, error)
	GetRecentTournaments(ctx context.Context, days, limit int) ([]TournamentResponse, error)
	GetTournamentDetails(ctx context.Context, tournamentID string) (*EnhancedTournamentResponse, error)

	// Regional operations
	GetRegions(ctx context.Context) ([]RegionInfo, error)
	GetRegionAddresses(ctx context.Context, region, addressType string) ([]RegionAddressResponse, error)

	// Administrative operations
	Health(ctx context.Context) (*HealthResponse, error)
	CacheStats(ctx context.Context) (*CacheStatsResponse, error)
}

// clientImpl implements the Client interface
type clientImpl struct {
	apiClient *api.Client
}

// NewClient creates a new Portal64 client
func NewClient(baseURL string, timeout time.Duration) Client {
	apiClient := api.NewClient(baseURL, timeout, nil)
	return &clientImpl{
		apiClient: apiClient,
	}
}

// Player operations
func (c *clientImpl) SearchPlayers(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	apiParams := api.SearchParams{
		Query:       params.Query,
		Limit:       params.Limit,
		Offset:      params.Offset,
		SortBy:      params.SortBy,
		SortOrder:   params.SortOrder,
		FilterBy:    params.FilterBy,
		FilterValue: params.FilterValue,
		Active:      params.Active,
	}
	
	result, err := c.apiClient.SearchPlayers(ctx, apiParams)
	if err != nil {
		return nil, err
	}
	
	return (*SearchResponse)(result), nil
}

func (c *clientImpl) GetPlayerProfile(ctx context.Context, playerID string) (*PlayerResponse, error) {
	result, err := c.apiClient.GetPlayerProfile(ctx, playerID)
	if err != nil {
		return nil, err
	}
	
	return (*PlayerResponse)(result), nil
}

func (c *clientImpl) GetPlayerRatingHistory(ctx context.Context, playerID string) ([]Evaluation, error) {
	result, err := c.apiClient.GetPlayerRatingHistory(ctx, playerID)
	if err != nil {
		return nil, err
	}
	
	evaluations := make([]Evaluation, len(result))
	for i, eval := range result {
		evaluations[i] = Evaluation(eval)
	}
	
	return evaluations, nil
}

// Club operations
func (c *clientImpl) SearchClubs(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	apiParams := api.SearchParams{
		Query:       params.Query,
		Limit:       params.Limit,
		Offset:      params.Offset,
		SortBy:      params.SortBy,
		SortOrder:   params.SortOrder,
		FilterBy:    params.FilterBy,
		FilterValue: params.FilterValue,
		Active:      params.Active,
	}
	
	result, err := c.apiClient.SearchClubs(ctx, apiParams)
	if err != nil {
		return nil, err
	}
	
	return (*SearchResponse)(result), nil
}

func (c *clientImpl) GetClubProfile(ctx context.Context, clubID string) (*ClubProfileResponse, error) {
	result, err := c.apiClient.GetClubProfile(ctx, clubID)
	if err != nil {
		return nil, err
	}
	
	return (*ClubProfileResponse)(result), nil
}

func (c *clientImpl) GetClubPlayers(ctx context.Context, clubID string, params SearchParams) (*SearchResponse, error) {
	apiParams := api.SearchParams{
		Query:       params.Query,
		Limit:       params.Limit,
		Offset:      params.Offset,
		SortBy:      params.SortBy,
		SortOrder:   params.SortOrder,
		FilterBy:    params.FilterBy,
		FilterValue: params.FilterValue,
		Active:      params.Active,
	}
	
	result, err := c.apiClient.GetClubPlayers(ctx, clubID, apiParams)
	if err != nil {
		return nil, err
	}
	
	return (*SearchResponse)(result), nil
}

func (c *clientImpl) GetClubStatistics(ctx context.Context, clubID string) (*ClubRatingStats, error) {
	result, err := c.apiClient.GetClubStatistics(ctx, clubID)
	if err != nil {
		return nil, err
	}
	
	return (*ClubRatingStats)(result), nil
}
// Tournament operations
func (c *clientImpl) SearchTournaments(ctx context.Context, params SearchParams) (*SearchResponse, error) {
	apiParams := api.SearchParams{
		Query:       params.Query,
		Limit:       params.Limit,
		Offset:      params.Offset,
		SortBy:      params.SortBy,
		SortOrder:   params.SortOrder,
		FilterBy:    params.FilterBy,
		FilterValue: params.FilterValue,
		Active:      params.Active,
	}
	
	result, err := c.apiClient.SearchTournaments(ctx, apiParams)
	if err != nil {
		return nil, err
	}
	
	return (*SearchResponse)(result), nil
}

func (c *clientImpl) SearchTournamentsByDate(ctx context.Context, params DateRangeParams) (*SearchResponse, error) {
	apiParams := api.DateRangeParams{
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
		SearchParams: api.SearchParams{
			Query:       params.Query,
			Limit:       params.Limit,
			Offset:      params.Offset,
			SortBy:      params.SortBy,
			SortOrder:   params.SortOrder,
			FilterBy:    params.FilterBy,
			FilterValue: params.FilterValue,
			Active:      params.Active,
		},
	}
	
	result, err := c.apiClient.SearchTournamentsByDate(ctx, apiParams)
	if err != nil {
		return nil, err
	}
	
	return (*SearchResponse)(result), nil
}

func (c *clientImpl) GetRecentTournaments(ctx context.Context, days, limit int) ([]TournamentResponse, error) {
	result, err := c.apiClient.GetRecentTournaments(ctx, days, limit)
	if err != nil {
		return nil, err
	}
	
	tournaments := make([]TournamentResponse, len(result))
	for i, tournament := range result {
		tournaments[i] = TournamentResponse(tournament)
	}
	
	return tournaments, nil
}

func (c *clientImpl) GetTournamentDetails(ctx context.Context, tournamentID string) (*EnhancedTournamentResponse, error) {
	result, err := c.apiClient.GetTournamentDetails(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	
	return (*EnhancedTournamentResponse)(result), nil
}

// Regional operations
func (c *clientImpl) GetRegions(ctx context.Context) ([]RegionInfo, error) {
	result, err := c.apiClient.GetRegions(ctx)
	if err != nil {
		return nil, err
	}
	
	regions := make([]RegionInfo, len(result))
	for i, region := range result {
		regions[i] = RegionInfo(region)
	}
	
	return regions, nil
}

func (c *clientImpl) GetRegionAddresses(ctx context.Context, region, addressType string) ([]RegionAddressResponse, error) {
	result, err := c.apiClient.GetRegionAddresses(ctx, region, addressType)
	if err != nil {
		return nil, err
	}
	
	addresses := make([]RegionAddressResponse, len(result))
	for i, address := range result {
		addresses[i] = RegionAddressResponse(address)
	}
	
	return addresses, nil
}

// Administrative operations
func (c *clientImpl) Health(ctx context.Context) (*HealthResponse, error) {
	result, err := c.apiClient.Health(ctx)
	if err != nil {
		return nil, err
	}
	
	return (*HealthResponse)(result), nil
}

func (c *clientImpl) CacheStats(ctx context.Context) (*CacheStatsResponse, error) {
	result, err := c.apiClient.CacheStats(ctx)
	if err != nil {
		return nil, err
	}
	
	return (*CacheStatsResponse)(result), nil
}
