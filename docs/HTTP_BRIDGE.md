# HTTP Bridge Implementation

This document describes the HTTP bridge implementation that allows the Portal64 MCP server to be accessed via REST API endpoints in addition to the standard MCP stdio protocol.

## Overview

The HTTP bridge provides REST API endpoints that wrap the existing MCP tools and resources, making them accessible via HTTP requests. This is essential for e2e testing where test frameworks expect HTTP endpoints rather than stdio communication.

## Configuration

The server now supports three modes via the `mcp.mode` configuration:

- `stdio`: Traditional MCP server using stdin/stdout communication (default)
- `http`: HTTP-only server exposing REST API endpoints
- `both`: Runs both stdio and HTTP servers simultaneously

Additional configuration options:
- `mcp.http_port`: Port for the HTTP server (default: 8888)
- `mcp.port`: Port for the traditional MCP server (only used in 'both' mode)

### Example Configuration

```yaml
api:
  base_url: "http://localhost:8080"
  timeout: "30s"

mcp:
  port: 3000
  mode: "http"        # or "stdio", "both"
  http_port: 8888

logging:
  level: "info"
  format: "json"
```

## Supported Endpoints

### Health and Admin
- `GET /health` - API health check
- `GET /api/v1/health` - API health check (versioned)
- `GET /api/v1/admin/cache` - Cache statistics

### MCP Protocol
- `GET /tools/list` - List available MCP tools
- `POST /tools/call` - Execute MCP tool
- `GET /resources/list` - List available MCP resources
- `POST /resources/read` - Read MCP resource

### Players
- `GET /api/v1/players` - Search players
- `GET /api/players/` - Search players (non-versioned)
- `GET /api/v1/players/{id}` - Get player profile
- `GET /api/players/{id}` - Get player profile (non-versioned)
- `GET /api/v1/players/{id}/history` - Get player rating history

### Clubs
- `GET /api/v1/clubs` - Search clubs
- `GET /api/clubs/` - Search clubs (non-versioned)
- `GET /api/v1/clubs/{id}` - Get club profile
- `GET /api/clubs/{id}` - Get club profile (non-versioned)
- `GET /api/v1/clubs/{id}/profile` - Get comprehensive club profile
- `GET /api/v1/clubs/{id}/players` - Get club players
- `GET /api/v1/clubs/{id}/statistics` - Get club statistics

### Tournaments
- `GET /api/v1/tournaments` - Search tournaments
- `GET /api/tournaments/` - Search tournaments (non-versioned)
- `GET /api/v1/tournaments/search` - Search tournaments by date range
- `GET /api/v1/tournaments/recent` - Get recent tournaments
- `GET /api/v1/tournaments/{id}` - Get tournament details
- `GET /api/tournaments/{id}` - Get tournament details (non-versioned)

### Regions
- `GET /api/v1/addresses/regions` - Get available regions
- `GET /api/v1/addresses/{region}` - Get region addresses

## Query Parameters

Most search endpoints support these query parameters:
- `query` - Search query string
- `limit` - Maximum number of results (default varies by endpoint)
- `offset` - Number of results to skip for pagination
- `sort_by` - Field to sort by
- `sort_order` - Sort order ("asc" or "desc")
- `filter_by` - Field to filter by
- `filter_value` - Value to filter by
- `active` - Filter for active records only (boolean)

Date range endpoints also support:
- `start_date` - Start date (YYYY-MM-DD format)
- `end_date` - End date (YYYY-MM-DD format)

## Response Format

All endpoints return JSON responses with appropriate HTTP status codes:
- `200 OK` - Successful request
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error responses follow this format:
```json
{
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

## CORS Support

All endpoints include CORS headers to allow cross-origin requests:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

## Architecture

The HTTP bridge (`HTTPBridge`) wraps the existing MCP server functionality:

1. **Request Handling**: HTTP requests are parsed and converted to MCP tool calls
2. **Parameter Mapping**: Query parameters are mapped to MCP tool arguments
3. **Response Translation**: MCP tool responses are converted to HTTP JSON responses
4. **Error Handling**: MCP errors are translated to appropriate HTTP status codes

The bridge maintains compatibility with the existing MCP protocol while providing convenient HTTP access for testing and integration purposes.

## Usage for E2E Testing

For e2e testing, start the server in HTTP mode:

```bash
./portal64mcp.exe -config test/e2e-config.yaml
```

The server will be available at `http://localhost:8888` and can be tested using standard HTTP clients or testing frameworks.

## Implementation Files

- `internal/mcp/http_bridge.go` - HTTP bridge implementation
- `internal/mcp/server.go` - Modified to support HTTP mode
- `internal/config/config.go` - Updated configuration structure
- `cmd/server/main.go` - Updated startup logging
