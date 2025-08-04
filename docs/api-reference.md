# Portal64 MCP Server - API Reference

This document provides a comprehensive reference for all MCP tools and resources available in the Portal64 MCP Server.

## Tools

### Search Tools

#### `search_players`
Search for players with filtering and pagination support.

**Parameters:**
- `query` (string, optional): Search query for player name
- `limit` (integer, optional): Maximum number of results (default: 50, max: 200)
- `offset` (integer, optional): Number of results to skip (default: 0)
- `sort_by` (string, optional): Field to sort by (`name`, `current_dwz`, `club`)
- `sort_order` (string, optional): Sort order (`asc`, `desc`)
- `active` (boolean, optional): Filter for active players only

**Example:**
```json
{
  "query": "Mueller",
  "limit": 10,
  "sort_by": "current_dwz",
  "sort_order": "desc",
  "active": true
}
```

#### `search_clubs`
Search for clubs with geographic and membership filtering.

**Parameters:**
- `query` (string, optional): Search query for club name
- `limit` (integer, optional): Maximum number of results (default: 50, max: 200)
- `offset` (integer, optional): Number of results to skip (default: 0)
- `sort_by` (string, optional): Field to sort by (`name`, `member_count`, `city`)
- `sort_order` (string, optional): Sort order (`asc`, `desc`)
- `filter_by` (string, optional): Field to filter by (`region`, `state`, `city`)
- `filter_value` (string, optional): Value to filter by when filter_by is specified

**Example:**
```json
{
  "query": "Schachclub",
  "filter_by": "state",
  "filter_value": "Bayern",
  "sort_by": "member_count",
  "sort_order": "desc"
}
```

#### `search_tournaments`
Search for tournaments with date and status filtering.

**Parameters:**
- `query` (string, optional): Search query for tournament name
- `limit` (integer, optional): Maximum number of results (default: 50, max: 200)
- `offset` (integer, optional): Number of results to skip (default: 0)
- `sort_by` (string, optional): Field to sort by
- `sort_order` (string, optional): Sort order (`asc`, `desc`)
- `filter_by` (string, optional): Field to filter by
- `filter_value` (string, optional): Value to filter by when filter_by is specified

#### `get_recent_tournaments`
Retrieve recent tournaments within specified days.

**Parameters:**
- `days` (integer, optional): Number of days to look back (default: 30)
- `limit` (integer, optional): Maximum number of results (default: 50)

**Example:**
```json
{
  "days": 7,
  "limit": 20
}
```

#### `search_tournaments_by_date`
Search tournaments within specific date ranges.

**Parameters:**
- `start_date` (string, required): Start date in YYYY-MM-DD format
- `end_date` (string, required): End date in YYYY-MM-DD format
- `query` (string, optional): Search query for tournament name
- `limit` (integer, optional): Maximum number of results (default: 50)
- `offset` (integer, optional): Number of results to skip (default: 0)

**Example:**
```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "query": "Stadtmeisterschaft"
}
```

### Detail Tools

#### `get_player_profile`
Get comprehensive player profile with rating history.

**Parameters:**
- `player_id` (string, required): Player ID in format C0101-123

**Example:**
```json
{
  "player_id": "C0101-12345"
}
```

#### `get_club_profile`
Get comprehensive club profile with members and statistics.

**Parameters:**
- `club_id` (string, required): Club ID in format C0101

**Example:**
```json
{
  "club_id": "C0101"
}
```

#### `get_tournament_details`
Get detailed tournament information with participants and games.

**Parameters:**
- `tournament_id` (string, required): Tournament ID

**Example:**
```json
{
  "tournament_id": "T2024-001"
}
```

#### `get_club_players`
Get club members with search and filtering.

**Parameters:**
- `club_id` (string, required): Club ID in format C0101
- `query` (string, optional): Search query for player name
- `limit` (integer, optional): Maximum number of results (default: 50)
- `offset` (integer, optional): Number of results to skip (default: 0)
- `sort_by` (string, optional): Field to sort by
- `active` (boolean, optional): Filter for active players only

### Analysis Tools

#### `get_player_rating_history`
Get player's DWZ rating evolution over time.

**Parameters:**
- `player_id` (string, required): Player ID in format C0101-123

#### `get_club_statistics`
Get club performance statistics and member analytics.

**Parameters:**
- `club_id` (string, required): Club ID in format C0101

### Administrative Tools

#### `check_api_health`
Check Portal64 API connectivity and health status.

**Parameters:** None

#### `get_cache_stats`
Get API cache performance metrics.

**Parameters:** None

#### `get_regions`
Get available regions for address lookups.

**Parameters:** None

#### `get_region_addresses`
Get chess official addresses by region.

**Parameters:**
- `region` (string, required): Region code
- `type` (string, optional): Address type (president, secretary, treasurer, etc.)

**Example:**
```json
{
  "region": "Bayern",
  "type": "president"
}
```

## Resources

Resources provide direct access to structured data via URI-based requests.

### Player Resources

#### `players://{id}`
Individual player details and information.

**URI Format:** `players://C0101-12345`

### Club Resources

#### `clubs://{id}`
Individual club details.

**URI Format:** `clubs://C0101`

#### `clubs://{id}/profile`
Comprehensive club profile with members and statistics.

**URI Format:** `clubs://C0101/profile`

### Tournament Resources

#### `tournaments://{id}`
Individual tournament details with participants and games.

**URI Format:** `tournaments://T2024-001`

### Address Resources

#### `addresses://regions`
List of available regions for address lookups.

#### `addresses://{region}`
Regional addresses for chess officials.

**URI Format:** `addresses://Bayern`

#### `addresses://{region}/{type}`
Specific address types by region.

**URI Format:** `addresses://Bayern/president`

### Administrative Resources

#### `admin://health`
Portal64 API health status and connectivity information.

#### `admin://cache`
API cache performance metrics and statistics.

## Response Formats

### Standard Response Structure

All tool responses follow this structure:

```json
{
  "content": [
    {
      "type": "text",
      "text": "JSON formatted response data"
    }
  ],
  "isError": false
}
```

### Error Response Structure

Error responses include additional error information:

```json
{
  "content": [
    {
      "type": "text",
      "text": "Error: Description of the error"
    }
  ],
  "isError": true
}
```

### Pagination Metadata

Search responses include pagination information:

```json
{
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 50,
    "offset": 0,
    "pages": 3,
    "page": 1
  }
}
```

## Data Models

### Player Response
```json
{
  "id": "C0101-12345",
  "name": "Mustermann",
  "firstname": "Max",
  "club_id": "C0101",
  "club": "Schachclub München",
  "current_dwz": 1850,
  "dwz_index": 15,
  "birth_year": 1985,
  "gender": "M",
  "nation": "GER",
  "status": "active",
  "fide_id": 4567890
}
```

### Club Response
```json
{
  "id": "C0101",
  "name": "Schachclub München e.V.",
  "short_name": "SC München",
  "association": "Bayerischer Schachbund",
  "region": "Bayern",
  "city": "München",
  "state": "Bayern",
  "country": "Deutschland",
  "founding_year": 1920,
  "member_count": 150,
  "active_count": 120,
  "status": "active"
}
```

### Tournament Response
```json
{
  "id": "T2024-001",
  "name": "Münchener Stadtmeisterschaft 2024",
  "organizer": "Schachclub München e.V.",
  "organizer_club_id": "C0101",
  "start_date": "2024-03-15T18:00:00Z",
  "end_date": "2024-05-17T22:00:00Z",
  "location": "Vereinsheim",
  "city": "München",
  "state": "Bayern",
  "country": "Deutschland",
  "tournament_type": "Swiss System",
  "time_control": "90 min + 30 sec/move",
  "rounds": 9,
  "participants": 45,
  "status": "completed",
  "evaluation_status": "evaluated"
}
```

## Error Codes

The server uses standard MCP error codes:

- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found (tool/resource not found)
- `-32602`: Invalid params (parameter validation failed)
- `-32603`: Internal error (API communication failed)

## Rate Limiting

The server does not implement rate limiting as it delegates all requests to the Portal64 API, which has its own rate limiting policies.

## Authentication

No authentication is required for the MCP server or the underlying Portal64 API. All data is publicly accessible in accordance with German chess federation policies.
