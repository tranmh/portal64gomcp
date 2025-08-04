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
	
	return &SearchResponse{
		Data: result.Data,
		Pagination: PaginationMetadata{
			Total:  result.Pagination.Total,
			Limit:  result.Pagination.Limit,
			Offset: result.Pagination.Offset,
			Pages:  result.Pagination.Pages,
			Page:   result.Pagination.Page,
		},
	}, nil
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
	
	return &SearchResponse{
		Data: result.Data,
		Pagination: PaginationMetadata{
			Total:  result.Pagination.Total,
			Limit:  result.Pagination.Limit,
			Offset: result.Pagination.Offset,
			Pages:  result.Pagination.Pages,
			Page:   result.Pagination.Page,
		},
	}, nil
}

func (c *clientImpl) GetClubProfile(ctx context.Context, clubID string) (*ClubProfileResponse, error) {
	result, err := c.apiClient.GetClubProfile(ctx, clubID)
	if err != nil {
		return nil, err
	}
	
	return convertClubProfileResponse(result), nil
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
	
	return &SearchResponse{
		Data: result.Data,
		Pagination: PaginationMetadata{
			Total:  result.Pagination.Total,
			Limit:  result.Pagination.Limit,
			Offset: result.Pagination.Offset,
			Pages:  result.Pagination.Pages,
			Page:   result.Pagination.Page,
		},
	}, nil
}

func (c *clientImpl) GetClubStatistics(ctx context.Context, clubID string) (*ClubRatingStats, error) {
	result, err := c.apiClient.GetClubStatistics(ctx, clubID)
	if err != nil {
		return nil, err
	}
	
	return &ClubRatingStats{
		AverageRating:      result.AverageRating,
		MedianRating:       result.MedianRating,
		HighestRating:      result.HighestRating,
		LowestRating:       result.LowestRating,
		RatingDistribution: result.RatingDistribution,
	}, nil
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
	
	return &SearchResponse{
		Data: result.Data,
		Pagination: PaginationMetadata{
			Total:  result.Pagination.Total,
			Limit:  result.Pagination.Limit,
			Offset: result.Pagination.Offset,
			Pages:  result.Pagination.Pages,
			Page:   result.Pagination.Page,
		},
	}, nil
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
	
	return &SearchResponse{
		Data: result.Data,
		Pagination: PaginationMetadata{
			Total:  result.Pagination.Total,
			Limit:  result.Pagination.Limit,
			Offset: result.Pagination.Offset,
			Pages:  result.Pagination.Pages,
			Page:   result.Pagination.Page,
		},
	}, nil
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
	
	return convertEnhancedTournamentResponse(result), nil
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
	
	return convertHealthResponse(result), nil
}

func (c *clientImpl) CacheStats(ctx context.Context) (*CacheStatsResponse, error) {
	result, err := c.apiClient.CacheStats(ctx)
	if err != nil {
		return nil, err
	}
	
	return convertCacheStatsResponse(result), nil
}


// Helper functions to convert between internal API types and public types

func convertClubProfileResponse(result *api.ClubProfileResponse) *ClubProfileResponse {
	if result == nil {
		return nil
	}
	
	response := &ClubProfileResponse{
		Club: convertClubResponse(result.Club),
		Contact: convertClubContact(result.Contact),
		RatingStats: convertClubRatingStats(result.RatingStats),
		PlayerCount: result.PlayerCount,
		ActivePlayerCount: result.ActivePlayerCount,
		TournamentCount: result.TournamentCount,
	}
	
	// Convert players slice
	if result.Players != nil {
		response.Players = make([]PlayerResponse, len(result.Players))
		for i, player := range result.Players {
			response.Players[i] = convertPlayerResponse(&player)
		}
	}
	
	// Convert teams slice
	if result.Teams != nil {
		response.Teams = make([]ClubTeam, len(result.Teams))
		for i, team := range result.Teams {
			response.Teams[i] = convertClubTeam(&team)
		}
	}
	
	// Convert recent tournaments slice
	if result.RecentTournaments != nil {
		response.RecentTournaments = make([]TournamentResponse, len(result.RecentTournaments))
		for i, tournament := range result.RecentTournaments {
			response.RecentTournaments[i] = convertTournamentResponse(&tournament)
		}
	}
	
	return response
}

func convertClubResponse(result *api.ClubResponse) *ClubResponse {
	if result == nil {
		return nil
	}
	return &ClubResponse{
		ID: result.ID,
		Name: result.Name,
		ShortName: result.ShortName,
		Association: result.Association,
		Region: result.Region,
		City: result.City,
		State: result.State,
		Country: result.Country,
		FoundingYear: result.FoundingYear,
		MemberCount: result.MemberCount,
		ActiveCount: result.ActiveCount,
		Status: result.Status,
	}
}

func convertPlayerResponse(result *api.PlayerResponse) PlayerResponse {
	return PlayerResponse{
		ID: result.ID,
		Name: result.Name,
		Firstname: result.Firstname,
		ClubID: result.ClubID,
		Club: result.Club,
		CurrentDWZ: result.CurrentDWZ,
		DWZIndex: result.DWZIndex,
		BirthYear: result.BirthYear,
		Gender: result.Gender,
		Nation: result.Nation,
		Status: result.Status,
		FideID: result.FideID,
	}
}

func convertClubContact(result *api.ClubContact) *ClubContact {
	if result == nil {
		return nil
	}
	return &ClubContact{
		President: result.President,
		VicePresident: result.VicePresident,
		Secretary: result.Secretary,
		Treasurer: result.Treasurer,
		Coach: result.Coach,
		Email: result.Email,
		Phone: result.Phone,
		Website: result.Website,
		Address: result.Address,
	}
}

func convertClubTeam(result *api.ClubTeam) ClubTeam {
	return ClubTeam{
		ID: result.ID,
		Name: result.Name,
		League: result.League,
		Division: result.Division,
		Season: result.Season,
	}
}

func convertClubRatingStats(result *api.ClubRatingStats) *ClubRatingStats {
	if result == nil {
		return nil
	}
	return &ClubRatingStats{
		AverageRating: result.AverageRating,
		MedianRating: result.MedianRating,
		HighestRating: result.HighestRating,
		LowestRating: result.LowestRating,
		RatingDistribution: result.RatingDistribution,
	}
}

func convertTournamentResponse(result *api.TournamentResponse) TournamentResponse {
	return TournamentResponse{
		ID: result.ID,
		Name: result.Name,
		Organizer: result.Organizer,
		OrganizerClubID: result.OrganizerClubID,
		StartDate: result.StartDate,
		EndDate: result.EndDate,
		Location: result.Location,
		City: result.City,
		State: result.State,
		Country: result.Country,
		TournamentType: result.TournamentType,
		TimeControl: result.TimeControl,
		Rounds: result.Rounds,
		Participants: result.Participants,
		Status: result.Status,
		EvaluationStatus: result.EvaluationStatus,
	}
}

func convertEnhancedTournamentResponse(result *api.EnhancedTournamentResponse) *EnhancedTournamentResponse {
	if result == nil {
		return nil
	}
	
	response := &EnhancedTournamentResponse{
		Tournament: convertTournamentResponsePtr(result.Tournament),
		Statistics: convertTournamentStatistics(result.Statistics),
	}
	
	// Convert participants slice
	if result.Participants != nil {
		response.Participants = make([]PlayerResponse, len(result.Participants))
		for i, participant := range result.Participants {
			response.Participants[i] = convertPlayerResponse(&participant)
		}
	}
	
	// Convert games slice
	if result.Games != nil {
		response.Games = make([]GameResult, len(result.Games))
		for i, game := range result.Games {
			response.Games[i] = convertGameResult(&game)
		}
	}
	
	// Convert evaluations slice
	if result.Evaluations != nil {
		response.Evaluations = make([]Evaluation, len(result.Evaluations))
		for i, evaluation := range result.Evaluations {
			response.Evaluations[i] = convertEvaluation(&evaluation)
		}
	}
	
	return response
}

func convertTournamentResponsePtr(result *api.TournamentResponse) *TournamentResponse {
	if result == nil {
		return nil
	}
	converted := convertTournamentResponse(result)
	return &converted
}

func convertTournamentStatistics(result *api.TournamentStatistics) *TournamentStatistics {
	if result == nil {
		return nil
	}
	return &TournamentStatistics{
		AverageRating: result.AverageRating,
		RatingRange: RatingRange{
			Min: result.RatingRange.Min,
			Max: result.RatingRange.Max,
		},
		NationDistribution: result.NationDistribution,
		AgeDistribution: result.AgeDistribution,
		GenderDistribution: result.GenderDistribution,
	}
}

func convertGameResult(result *api.GameResult) GameResult {
	return GameResult{
		ID: result.ID,
		TournamentID: result.TournamentID,
		Round: result.Round,
		WhitePlayer: result.WhitePlayer,
		BlackPlayer: result.BlackPlayer,
		Result: result.Result,
		Date: result.Date,
		PGN: result.PGN,
	}
}

func convertEvaluation(result *api.Evaluation) Evaluation {
	return Evaluation{
		ID: result.ID,
		PlayerID: result.PlayerID,
		TournamentID: result.TournamentID,
		OldDWZ: result.OldDWZ,
		NewDWZ: result.NewDWZ,
		DWZChange: result.DWZChange,
		Performance: result.Performance,
		Games: result.Games,
		Points: result.Points,
		Date: result.Date,
		Type: result.Type,
	}
}

func convertHealthResponse(result *api.HealthResponse) *HealthResponse {
	if result == nil {
		return nil
	}
	
	response := &HealthResponse{
		Status: result.Status,
		ResponseTime: result.ResponseTime,
		APIVersion: result.APIVersion,
		Timestamp: result.Timestamp,
	}
	
	// Convert services map
	if result.Services != nil {
		response.Services = make(map[string]ServiceHealth)
		for key, service := range result.Services {
			response.Services[key] = ServiceHealth{
				Status: service.Status,
				ResponseTime: service.ResponseTime,
				LastCheck: service.LastCheck,
				ErrorMessage: service.ErrorMessage,
			}
		}
	}
	
	return response
}

func convertCacheStatsResponse(result *api.CacheStatsResponse) *CacheStatsResponse {
	if result == nil {
		return nil
	}
	
	return &CacheStatsResponse{
		HitRatio: result.HitRatio,
		Operations: CacheOperations{
			Hits: result.Operations.Hits,
			Misses: result.Operations.Misses,
			Sets: result.Operations.Sets,
			Deletes: result.Operations.Deletes,
			Flushes: result.Operations.Flushes,
		},
		Performance: CachePerformance{
			AverageGetTime: result.Performance.AverageGetTime,
			AverageSetTime: result.Performance.AverageSetTime,
			ConnectionTime: result.Performance.ConnectionTime,
		},
		Usage: CacheUsage{
			UsedMemory: result.Usage.UsedMemory,
			MaxMemory: result.Usage.MaxMemory,
			MemoryPercent: result.Usage.MemoryPercent,
			KeyCount: result.Usage.KeyCount,
			ExpiredKeys: result.Usage.ExpiredKeys,
		},
		Timestamp: result.Timestamp,
	}
}
