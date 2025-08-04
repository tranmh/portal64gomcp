# Portal64 MCP Server - High Level Design

## Overview

The Portal64 MCP Server is a Golang-based Model Context Protocol (MCP) remote server that provides structured access to the German chess DWZ (Deutsche Wertungszahl) rating system via the existing Portal64 REST API. This server enables AI systems to query player ratings, club information, tournament data, and administrative information through a standardized MCP interface.

## Architecture

### System Architecture
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   MCP Client    │◄──►│  MCP Go Server   │◄──►│  Portal64 API   │
│   (Claude)      │    │  (This Project)  │    │ (localhost:8080) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Component Overview
- **MCP Server**: Golang application implementing MCP protocol
- **API Client**: HTTP client for Portal64 REST API integration
- **Tool Handlers**: Business logic for each MCP tool
- **Resource Handlers**: Resource management for detailed data access
- **Models**: Go structs matching Portal64 API response schemas

## MCP Interface Design

### Hybrid Approach: Tools + Resources

**Tools**: For active operations and searches
**Resources**: For detailed data access and browsing

### Tools Specification

#### Search Tools
```go
// Player search with filtering and pagination
search_players(query, limit, offset, sort_by, sort_order, active) 
→ []PlayerResponse + pagination metadata

// Club search with geographic and membership filtering  
search_clubs(query, limit, offset, sort_by, filter_by, filter_value)
→ []ClubResponse + pagination metadata

// Tournament search with date and status filtering
search_tournaments(query, limit, offset, sort_by, filter_by, filter_value)
→ []TournamentResponse + pagination metadata

// Recent tournament lookup
get_recent_tournaments(days, limit)
→ []TournamentResponse

// Tournament search by date range
search_tournaments_by_date(start_date, end_date, query, limit, offset)
→ []TournamentResponse + pagination metadata
```

#### Detail Tools
```go
// Get comprehensive player profile with rating history
get_player_profile(player_id) 
→ PlayerResponse + rating_history[]

// Get comprehensive club profile with members and statistics
get_club_profile(club_id)
→ ClubProfileResponse (includes players, stats, contact, teams)

// Get detailed tournament information with participants and games
get_tournament_details(tournament_id)
→ EnhancedTournamentResponse (includes participants, games, evaluations)

// Get club members with search and filtering
get_club_players(club_id, query, limit, offset, sort_by, active)
→ []PlayerResponse + pagination metadata
```

#### Analysis Tools
```go
// Get player's DWZ rating evolution over time
get_player_rating_history(player_id)
→ []Evaluation (chronological rating changes)

// Get club performance statistics and member analytics
get_club_statistics(club_id)
→ ClubRatingStats + member_distribution + activity_metrics
```

#### Administrative Tools
```go
// Check Portal64 API connectivity and health
check_api_health()
→ health_status + response_time + api_version

// Get API cache performance metrics
get_cache_stats()
→ CacheStatsResponse (hit_ratio, operations, performance, usage)

// Get available regions for address lookups
get_regions()
→ []RegionInfo

// Get addresses for chess officials by region
get_region_addresses(region, type)
→ []RegionAddressResponse
```

### Resources Specification

Resources provide direct access to structured data for browsing and detailed analysis:

```
/players/{id}                    # Individual player details
/clubs/{id}                      # Individual club details  
/clubs/{id}/profile             # Comprehensive club profile
/tournaments/{id}               # Individual tournament details
/addresses/regions              # Available regions list
/addresses/{region}             # Regional addresses
/addresses/{region}/{type}      # Specific address types
/admin/health                   # API health status
/admin/cache                    # Cache statistics
```

## Implementation Structure

### Project Layout
```
portal64gomcp/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── api/
│   │   ├── client.go              # Portal64 API client
│   │   └── models.go              # API response models
│   ├── mcp/
│   │   ├── server.go              # MCP server implementation
│   │   ├── tools.go               # Tool handlers
│   │   ├── resources.go           # Resource handlers
│   │   └── protocol.go            # MCP protocol handling
│   └── handlers/
│       ├── players.go             # Player-related operations
│       ├── clubs.go               # Club-related operations
│       ├── tournaments.go         # Tournament-related operations
│       ├── addresses.go           # Address-related operations
│       └── admin.go               # Administrative operations
├── pkg/
│   └── portal64/
│       ├── client.go              # Public API client interface
│       └── types.go               # Public type definitions
├── docs/
│   ├── high-level-design.md       # This document
│   ├── api-reference.md           # MCP tool/resource reference
│   └── deployment.md              # Deployment instructions
├── go.mod
├── go.sum
├── README.md
└── Makefile
```

### Core Components

#### 1. MCP Server (`internal/mcp/server.go`)
- Implements MCP protocol specification
- Handles client connections and message routing
- Manages tool and resource registration
- Provides error handling and logging

#### 2. API Client (`internal/api/client.go`)
- HTTP client for Portal64 REST API
- Request/response handling with proper error management
- JSON unmarshaling to Go structs
- Connection pooling and timeout management

#### 3. Tool Handlers (`internal/mcp/tools.go`)
- Implementation of each MCP tool
- Parameter validation and sanitization
- API client integration
- Response formatting for MCP protocol

#### 4. Resource Handlers (`internal/mcp/resources.go`)
- Resource URI routing and handling
- Dynamic content generation
- Content-Type management (JSON only)
- Resource metadata management

### Data Models

Go structs will mirror the Portal64 API response schemas:

```go
type PlayerResponse struct {
    ID          string  `json:"id"`           // Format: C0101-123
    Name        string  `json:"name"`
    Firstname   string  `json:"firstname"`
    ClubID      string  `json:"club_id"`      // Format: C0101
    Club        string  `json:"club"`
    CurrentDWZ  int     `json:"current_dwz"`
    DWZIndex    int     `json:"dwz_index"`
    BirthYear   int     `json:"birth_year"`
    Gender      string  `json:"gender"`
    Nation      string  `json:"nation"`
    Status      string  `json:"status"`
    FideID      int     `json:"fide_id"`
}

type ClubProfileResponse struct {
    Club                *ClubResponse        `json:"club"`
    Players            []PlayerResponse      `json:"players"`
    Contact            *ClubContact         `json:"contact"`
    Teams              []ClubTeam           `json:"teams"`
    RatingStats        *ClubRatingStats     `json:"rating_stats"`
    RecentTournaments  []TournamentResponse `json:"recent_tournaments"`
    PlayerCount        int                  `json:"player_count"`
    ActivePlayerCount  int                  `json:"active_player_count"`
    TournamentCount    int                  `json:"tournament_count"`
}

// Additional models following Portal64 API schema...
```

## API Integration

### HTTP Client Configuration
- Base URL: `http://localhost:8080`
- Request timeout: 30 seconds
- Connection pooling: Enabled
- Keep-alive: Enabled
- Content-Type: `application/json`

### Error Handling Strategy
- **API Unavailable**: Return MCP error with "API temporarily unavailable"
- **Invalid Parameters**: Return MCP error with parameter validation details
- **Not Found**: Return empty results with appropriate metadata
- **Network Errors**: Return MCP error with connection details
- **Malformed Responses**: Return MCP error with parsing details

### Request Patterns
```go
// Standard API request pattern
func (c *Client) searchPlayers(params SearchParams) (*SearchResponse, error) {
    url := c.buildURL("/api/v1/players", params)
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("API request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, c.handleErrorResponse(resp)
    }
    
    var result SearchResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("response parsing failed: %w", err)
    }
    
    return &result, nil
}
```

## Configuration

### Environment Variables
```bash
PORTAL64_API_URL=http://localhost:8080    # Portal64 API base URL
MCP_SERVER_PORT=3000                      # MCP server port
LOG_LEVEL=info                            # Logging level
API_TIMEOUT=30s                           # API request timeout
```

### Configuration File Support
```yaml
# config.yaml
api:
  base_url: "http://localhost:8080"
  timeout: "30s"
  
mcp:
  port: 3000
  
logging:
  level: "info"
  format: "json"
```

## Development Workflow

### Build and Run
```bash
# Build the server
go build -o bin/portal64-mcp cmd/server/main.go

# Run with default configuration
./bin/portal64-mcp

# Run with custom config
./bin/portal64-mcp -config config.yaml
```

### Testing Strategy
- Unit tests for all tool and resource handlers
- Integration tests with mock Portal64 API
- MCP protocol compliance testing
- Performance testing with concurrent clients

### Dependencies
```go
// Core dependencies
require (
    github.com/gorilla/mux v1.8.1           // HTTP routing
    github.com/spf13/viper v1.18.2          // Configuration management
    github.com/sirupsen/logrus v1.9.3       // Logging
    github.com/stretchr/testify v1.8.4      // Testing framework
)
```

## Security Considerations

### Authentication
- No authentication required (matches Portal64 API)
- MCP server runs on localhost only

### Data Privacy
- Follows Portal64 API privacy implementation (GDPR compliant)
- Birth year only, no full birth dates
- No additional PII exposure through MCP layer

### Network Security
- Local-only binding by default
- No external network exposure
- API communication over HTTP (localhost)

## Performance Considerations

### Caching Strategy
- **No local caching**: Delegate completely to Portal64 API's Redis cache
- Direct API passthrough for all requests
- Stateless server design

### Concurrency
- Goroutine-based request handling
- Connection pooling for API requests
- Non-blocking I/O operations

### Resource Management
- Connection timeouts: 30 seconds
- Request size limits: Follow API constraints
- Memory usage: Minimal buffering, streaming responses where possible

## Monitoring and Observability

### Logging
- Structured JSON logging
- Request/response logging with correlation IDs
- Error tracking and categorization
- Performance metrics logging

### Health Checks
- MCP server health endpoint
- Portal64 API connectivity checks
- Resource utilization monitoring

### Metrics
- Request counts by tool/resource
- Response times and error rates
- API client connection metrics
- Memory and CPU usage tracking

## Future Enhancements

### Potential Extensions
1. **Batch Operations**: Multi-player/club queries in single request
2. **Real-time Updates**: WebSocket support for live tournament data
3. **Advanced Analytics**: Statistical analysis tools for chess performance
4. **Export Capabilities**: Structured data export in multiple formats
5. **Caching Layer**: Optional local caching with TTL management

### Scalability Considerations
- Horizontal scaling through load balancing
- Database connection pooling
- Request rate limiting implementation
- Circuit breaker patterns

## Conclusion

This design provides a robust, maintainable MCP server that efficiently bridges AI systems with the German chess rating infrastructure. The hybrid tool/resource approach offers both convenience for common operations and flexibility for detailed data access, while maintaining simplicity through direct API delegation and stateless architecture.

The implementation prioritizes:
- **Simplicity**: Direct API passthrough without complex caching
- **Reliability**: Comprehensive error handling and logging
- **Performance**: Efficient HTTP client usage and concurrent request handling
- **Maintainability**: Clean architecture with separation of concerns
- **Compliance**: Full adherence to existing privacy and data protection measures