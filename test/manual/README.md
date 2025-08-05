# Manual Test Scripts

This directory contains manual testing scripts and test data for the Portal64 MCP API.

## Files

### PowerShell Scripts (*.ps1)
These scripts make HTTP requests to test specific API endpoints:

- `test_club_players.ps1` - Test get_club_players endpoint
- `test_club_profile.ps1` - Test get_club_profile endpoint  
- `test_player_history.ps1` - Test get_player_history endpoint
- `test_player_profile.ps1` - Test get_player_profile endpoint
- `test_recent_tournaments.ps1` - Test get_recent_tournaments endpoint
- `test_schema_fix.ps1` - Test schema-related functionality
- `test_tool.ps1` - Test general tool functionality

### JSON Files (*.json)
These files contain test payload data:

- `test_health.json` - Test data for health check endpoint
- `test_rating_history.json` - Test data for rating history endpoint
- `test_regions.json` - Test data for regions endpoint

## Usage

1. Ensure the MCP server is running on localhost:8888
2. Run any PowerShell script to test the corresponding endpoint:
   ```powershell
   .\test_tool.ps1
   ```

## Note

These are development/debugging tools for manual testing. For automated tests, see the parent test directory.
