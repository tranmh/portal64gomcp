# Tournament ID Format Change Fix

## Problem
The rating history API route `http://localhost:8080/api/v1/players/C0327-297/rating-history` changed its response format:

**Before:**
- Field: `id_master` (integer, e.g., 97693)
- Type: `int`

**After:**
- Field: `tournament_id` (string, e.g., "B735-705-QCB")  
- Type: `string`

## Changes Made

### 1. Updated RatingHistoryEntry struct in `internal/api/models.go`
**Before:**
```go
type RatingHistoryEntry struct {
	ID           int     `json:"id"`
	IDMaster     int     `json:"id_master"`  // ← OLD
	IDPerson     int     `json:"id_person"`
	// ... other fields
}
```

**After:**
```go
type RatingHistoryEntry struct {
	ID           int     `json:"id"`
	TournamentID string  `json:"tournament_id"`  // ← NEW
	IDPerson     int     `json:"id_person"`
	// ... other fields
}
```

### 2. Updated conversion logic in `internal/api/client.go`
**Before:**
```go
evaluations[i] = Evaluation{
	ID:          fmt.Sprintf("%d", entry.ID),
	PlayerID:    playerID,
	// Missing TournamentID mapping
	OldDWZ:      entry.DWZOld,
	// ... other fields
}
```

**After:**
```go
evaluations[i] = Evaluation{
	ID:           fmt.Sprintf("%d", entry.ID),
	PlayerID:     playerID,
	TournamentID: entry.TournamentID,  // ← NEW mapping
	OldDWZ:       entry.DWZOld,
	// ... other fields
}
```

## Impact
- ✅ API now correctly parses new tournament_id format (strings like "B735-705-QCB")
- ✅ Evaluation objects now include tournament_id information
- ✅ Backward compatibility maintained for existing code using Evaluation struct
- ✅ Public API types in `pkg/portal64/types.go` already had correct TournamentID field
- ✅ Test fixtures already used new format

## Files Modified
1. `internal/api/models.go` - Updated RatingHistoryEntry struct
2. `internal/api/client.go` - Updated GetPlayerRatingHistory conversion logic

## Testing
The changes maintain the existing API contract while adapting to the new backend response format. All existing functionality should work unchanged.
