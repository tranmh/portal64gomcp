package api

import "time"

// SearchResponse represents a paginated search response
type SearchResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Pages  int `json:"pages"`
	Page   int `json:"page"`
}

// PlayerResponse represents a player in the system
type PlayerResponse struct {
	ID          string `json:"id"`           // Format: C0101-123
	Name        string `json:"name"`
	Firstname   string `json:"firstname"`
	ClubID      string `json:"club_id"`      // Format: C0101
	Club        string `json:"club"`
	CurrentDWZ  int    `json:"current_dwz"`
	DWZIndex    int    `json:"dwz_index"`
	BirthYear   int    `json:"birth_year"`
	Gender      string `json:"gender"`
	Nation      string `json:"nation"`
	Status      string `json:"status"`
	FideID      int    `json:"fide_id"`
}

// ClubResponse represents a chess club
type ClubResponse struct {
	ID            string `json:"id"`              // Format: C0101
	Name          string `json:"name"`
	ShortName     string `json:"short_name"`
	Association   string `json:"association"`
	Region        string `json:"region"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	FoundingYear  int    `json:"founding_year"`
	MemberCount   int    `json:"member_count"`
	ActiveCount   int    `json:"active_count"`
	Status        string `json:"status"`
}

// ClubProfileResponse represents comprehensive club information
type ClubProfileResponse struct {
	Club                *ClubResponse        `json:"club"`
	Players             []PlayerResponse     `json:"players"`
	Contact             *ClubContact         `json:"contact"`
	Teams               []ClubTeam           `json:"teams"`
	RatingStats         *ClubRatingStats     `json:"rating_stats"`
	RecentTournaments   []TournamentResponse `json:"recent_tournaments"`
	PlayerCount         int                  `json:"player_count"`
	ActivePlayerCount   int                  `json:"active_player_count"`
	TournamentCount     int                  `json:"tournament_count"`
}

// ClubContact represents club contact information
type ClubContact struct {
	President    string `json:"president"`
	VicePresident string `json:"vice_president"`
	Secretary    string `json:"secretary"`
	Treasurer    string `json:"treasurer"`
	Coach        string `json:"coach"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Address      string `json:"address"`
}
// ClubTeam represents a club team
type ClubTeam struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	League   string `json:"league"`
	Division string `json:"division"`
	Season   string `json:"season"`
}

// ClubRatingStats represents club rating statistics
type ClubRatingStats struct {
	AverageRating     float64 `json:"average_rating"`
	MedianRating      float64 `json:"median_rating"`
	HighestRating     int     `json:"highest_rating"`
	LowestRating      int     `json:"lowest_rating"`
	RatingDistribution map[string]int `json:"rating_distribution"`
}

// TournamentResponse represents a chess tournament
type TournamentResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Organizer        string    `json:"organizer"`
	OrganizerClubID  string    `json:"organizer_club_id"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Location         string    `json:"location"`
	City             string    `json:"city"`
	State            string    `json:"state"`
	Country          string    `json:"country"`
	TournamentType   string    `json:"tournament_type"`
	TimeControl      string    `json:"time_control"`
	Rounds           int       `json:"rounds"`
	Participants     int       `json:"participants"`
	Status           string    `json:"status"`
	EvaluationStatus string    `json:"evaluation_status"`
}

// EnhancedTournamentResponse represents detailed tournament information
type EnhancedTournamentResponse struct {
	Tournament   *TournamentResponse    `json:"tournament"`
	Participants []PlayerResponse       `json:"participants"`
	Games        []GameResult           `json:"games"`
	Evaluations  []Evaluation           `json:"evaluations"`
	Statistics   *TournamentStatistics  `json:"statistics"`
}

// GameResult represents a single game result
type GameResult struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournament_id"`
	Round        int       `json:"round"`
	WhitePlayer  string    `json:"white_player"`
	BlackPlayer  string    `json:"black_player"`
	Result       string    `json:"result"`     // "1-0", "0-1", "1/2-1/2"
	Date         time.Time `json:"date"`
	PGN          string    `json:"pgn,omitempty"`
}

// Evaluation represents DWZ rating evaluation
type Evaluation struct {
	ID           string    `json:"id"`
	PlayerID     string    `json:"player_id"`
	TournamentID string    `json:"tournament_id"`
	OldDWZ       int       `json:"old_dwz"`
	NewDWZ       int       `json:"new_dwz"`
	DWZChange    int       `json:"dwz_change"`
	Performance  int       `json:"performance"`
	Games        int       `json:"games"`
	Points       float64   `json:"points"`
	Date         time.Time `json:"date"`
	Type         string    `json:"type"`       // "tournament", "rapid", "blitz"
}

// TournamentStatistics represents tournament statistics
type TournamentStatistics struct {
	AverageRating    float64            `json:"average_rating"`
	RatingRange      RatingRange        `json:"rating_range"`
	NationDistribution map[string]int   `json:"nation_distribution"`
	AgeDistribution    map[string]int   `json:"age_distribution"`
	GenderDistribution map[string]int   `json:"gender_distribution"`
}

// RatingRange represents rating range statistics
type RatingRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}
// RegionInfo represents information about a region
type RegionInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	AddressTypes []string `json:"address_types"`
}

// RegionAddressResponse represents regional addresses
type RegionAddressResponse struct {
	ID          string `json:"id"`
	Region      string `json:"region"`
	Type        string `json:"type"`        // "president", "secretary", "treasurer", etc.
	Name        string `json:"name"`
	Position    string `json:"position"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	City        string `json:"city"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"country"`
}

// HealthResponse represents API health status
type HealthResponse struct {
	Status       string                 `json:"status"`        // "healthy", "degraded", "unhealthy"
	ResponseTime int64                  `json:"response_time"` // in milliseconds
	APIVersion   string                 `json:"api_version"`
	Timestamp    time.Time              `json:"timestamp"`
	Services     map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents individual service health
type ServiceHealth struct {
	Status       string `json:"status"`
	ResponseTime int64  `json:"response_time"`
	LastCheck    time.Time `json:"last_check"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CacheStatsResponse represents cache performance metrics
type CacheStatsResponse struct {
	HitRatio    float64                `json:"hit_ratio"`
	Operations  CacheOperations        `json:"operations"`
	Performance CachePerformance       `json:"performance"`
	Usage       CacheUsage             `json:"usage"`
	Timestamp   time.Time              `json:"timestamp"`
}

// CacheOperations represents cache operation statistics
type CacheOperations struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
	Sets   int64 `json:"sets"`
	Deletes int64 `json:"deletes"`
	Flushes int64 `json:"flushes"`
}

// CachePerformance represents cache performance metrics
type CachePerformance struct {
	AverageGetTime time.Duration `json:"average_get_time"`
	AverageSetTime time.Duration `json:"average_set_time"`
	ConnectionTime time.Duration `json:"connection_time"`
}

// CacheUsage represents cache usage statistics
type CacheUsage struct {
	UsedMemory    int64   `json:"used_memory"`     // in bytes
	MaxMemory     int64   `json:"max_memory"`      // in bytes
	MemoryPercent float64 `json:"memory_percent"`
	KeyCount      int64   `json:"key_count"`
	ExpiredKeys   int64   `json:"expired_keys"`
}

// SearchParams represents common search parameters
type SearchParams struct {
	Query       string `json:"query,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
	SortBy      string `json:"sort_by,omitempty"`
	SortOrder   string `json:"sort_order,omitempty"`
	FilterBy    string `json:"filter_by,omitempty"`
	FilterValue string `json:"filter_value,omitempty"`
	Active      *bool  `json:"active,omitempty"`
}

// DateRangeParams represents date range search parameters
type DateRangeParams struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	SearchParams
}
