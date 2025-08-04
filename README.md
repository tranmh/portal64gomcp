# Portal64 MCP Server

A Golang-based Model Context Protocol (MCP) server that provides structured access to the German chess DWZ (Deutsche Wertungszahl) rating system via the Portal64 REST API.

## Overview

This MCP server acts as a bridge between AI systems (like Claude) and the German chess rating infrastructure, enabling intelligent queries about players, clubs, tournaments, and administrative information through a standardized MCP interface.

## Features

### Search Tools
- **search_players**: Search for players with filtering and pagination
- **search_clubs**: Search for clubs with geographic and membership filtering  
- **search_tournaments**: Search for tournaments with date and status filtering
- **get_recent_tournaments**: Retrieve recent tournaments within specified days
- **search_tournaments_by_date**: Search tournaments within date ranges

### Detail Tools
- **get_player_profile**: Get comprehensive player profiles with rating history
- **get_club_profile**: Get comprehensive club profiles with members and statistics
- **get_tournament_details**: Get detailed tournament information with participants
- **get_club_players**: Get club members with search and filtering

### Analysis Tools
- **get_player_rating_history**: Get player's DWZ rating evolution over time
- **get_club_statistics**: Get club performance statistics and member analytics

### Administrative Tools
- **check_api_health**: Check Portal64 API connectivity and health
- **get_cache_stats**: Get API cache performance metrics
- **get_regions**: Get available regions for address lookups
- **get_region_addresses**: Get chess official addresses by region

### Resources
Direct access to structured data via URI-based resources:
- `players://{id}` - Individual player details
- `clubs://{id}` - Individual club details  
- `clubs://{id}/profile` - Comprehensive club profiles
- `tournaments://{id}` - Individual tournament details
- `addresses://regions` - Available regions list
- `addresses://{region}` - Regional addresses
- `admin://health` - API health status
- `admin://cache` - Cache statistics

## Prerequisites

- Go 1.21 or later
- Access to Portal64 API (default: http://localhost:8080)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd portal64gomcp
```

2. Install dependencies:
```bash
go mod download
```

3. Build the server:
```bash
make build
```

## Configuration

### Environment Variables
```bash
PORTAL64_API_URL=http://localhost:8080    # Portal64 API base URL
MCP_SERVER_PORT=3000                      # MCP server port (unused for stdio)
LOG_LEVEL=info                            # Logging level
API_TIMEOUT=30s                           # API request timeout
```

### Configuration File
Create a `config.yaml` file:
```yaml
api:
  base_url: "http://localhost:8080"
  timeout: "30s"
  
mcp:
  port: 3000
  
logging:
  level: "info"
  format: "json"
```

## Usage

### Running the Server
```bash
# Run with default configuration
./bin/portal64-mcp

# Run with custom config file
./bin/portal64-mcp -config config.yaml

# Run with debug logging
./bin/portal64-mcp -log-level debug
```

### MCP Client Integration
The server communicates via stdio following the MCP protocol. Configure your MCP client to launch the server executable.

## Development

### Building
```bash
make build          # Build binary
make clean          # Clean build artifacts
make test           # Run tests (when implemented)
make run            # Build and run
```

### Project Structure
```
portal64gomcp/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── config/config.go         # Configuration management
│   ├── api/                     # Portal64 API client
│   │   ├── client.go           # HTTP client implementation
│   │   └── models.go           # API response models
│   └── mcp/                    # MCP server implementation
│       ├── server.go           # Main server logic
│       ├── protocol.go         # MCP protocol structures
│       ├── tools.go            # Tool handlers
│       └── resources.go        # Resource handlers
├── docs/                       # Documentation
└── README.md                   # This file
```

## API Integration

The server integrates with the Portal64 REST API:
- **Base URL**: Configurable (default: http://localhost:8080)
- **Timeout**: Configurable (default: 30 seconds)
- **Authentication**: None required (matches Portal64 API)
- **Data Format**: JSON responses

## Error Handling

The server provides comprehensive error handling:
- **API Unavailable**: Returns MCP error with clear message
- **Invalid Parameters**: Returns validation error details
- **Not Found**: Returns empty results with metadata
- **Network Errors**: Returns connection error details

## Logging

Structured logging with configurable levels and formats:
- **Levels**: debug, info, warn, error
- **Formats**: text (default), json
- **Output**: stderr (stdout reserved for MCP protocol)

## Performance

- **Stateless Design**: No local caching, delegates to Portal64 API
- **Connection Pooling**: Efficient HTTP client with keep-alive
- **Concurrent Handling**: Goroutine-based request processing
- **Timeout Management**: Configurable request timeouts

## Security

- **Local Only**: Server binds to localhost by default
- **No Authentication**: Follows Portal64 API security model
- **Privacy Compliant**: Maintains Portal64's GDPR compliance
- **Data Passthrough**: No additional PII exposure

## Troubleshooting

### Common Issues

1. **API Connection Failed**
   - Verify Portal64 API is running on configured URL
   - Check network connectivity and firewall settings

2. **Tool Not Found**
   - Verify tool name matches registered tools
   - Check MCP client tool listing

3. **Invalid Parameters**
   - Verify parameter types match tool schema
   - Check required parameters are provided

4. **Resource Not Found**
   - Verify resource URI format
   - Check resource scheme is supported

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[License information to be added]

## Support

For issues and questions:
- Check the documentation in `docs/`
- Review the high-level design document
- Submit issues via the repository issue tracker
