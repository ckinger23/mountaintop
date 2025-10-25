# Leaderboard Query Builder Refactoring

## Overview

Refactored the `GetLeaderboard` handler from raw SQL to a type-safe, testable query builder pattern.

## Changes Made

### 1. New Package: `internal/leaderboard`

Created a dedicated package for leaderboard logic with:
- **Query Builder Pattern**: Fluent API for building leaderboard queries
- **Type Safety**: Compile-time checking of field names and types
- **Testability**: Easy to unit test with in-memory databases

### 2. Files Created

#### `internal/leaderboard/leaderboard.go`
```go
// Query builder with fluent API
query := NewQuery(db).ForSeason(seasonID).Execute()

// Or use convenience function
entries, err := GetLeaderboard(db, &seasonID)
```

**Key Features:**
- GORM query builder (no raw SQL strings)
- Optional season filtering
- Proper error handling
- Clear return types

#### `internal/leaderboard/leaderboard_test.go`
Comprehensive test suite with 6 test cases:
1. ✅ All seasons leaderboard
2. ✅ Filtered by season (2024)
3. ✅ Filtered by season (2025)
4. ✅ Query builder chaining
5. ✅ Empty database
6. ✅ Users with no picks

**Test Coverage:**
- In-memory SQLite for fast, isolated tests
- Seeds realistic test data (users, seasons, weeks, games, picks)
- Tests aggregations, win percentages, ordering
- Tests edge cases (no picks, empty DB)

### 3. Handler Refactored: `internal/handlers/games.go`

**Before (Raw SQL):**
```go
query := `
    SELECT u.id as user_id, ...
    FROM users u
    LEFT JOIN picks p ON u.id = p.user_id
    ...
`
if seasonIDStr != "" {
    seasonID, _ := strconv.Atoi(seasonIDStr)
    query += ` WHERE w.season_id = ?`
    a.DB.Raw(query+` GROUP BY u.id ...`, seasonID).Scan(&leaderboard)
} else {
    a.DB.Raw(query + ` GROUP BY u.id ...`).Scan(&leaderboard)
}
```

**After (Type-Safe Builder):**
```go
var seasonID *uint
if seasonIDStr != "" {
    id, err := strconv.ParseUint(seasonIDStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid season_id parameter", http.StatusBadRequest)
        return
    }
    uid := uint(id)
    seasonID = &uid
}

entries, err := leaderboard.GetLeaderboard(a.DB, seasonID)
if err != nil {
    http.Error(w, "Error fetching leaderboard", http.StatusInternalServerError)
    return
}
```

## Benefits Gained

### Type Safety ✅
- Compiler catches field name typos
- IDE autocomplete works
- No runtime errors from SQL syntax mistakes

### Testability ✅
- Unit tests run in milliseconds
- No need for test database setup
- Easy to test edge cases

### Maintainability ✅
- Query logic separated from HTTP handler
- Reusable across different contexts
- Single source of truth for leaderboard calculation

### Performance ⚡
- Still uses efficient database aggregation
- No N+1 queries
- Minimal memory footprint

### Error Handling ✅
- Proper input validation (ParseUint vs Atoi)
- Clear error messages
- No silent failures

## SQL vs Go Code Tradeoffs

| Aspect | Raw SQL | GORM Builder |
|--------|---------|--------------|
| **Performance** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ Same |
| **Type Safety** | ⭐ | ⭐⭐⭐⭐⭐ |
| **Testability** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Readability** | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **DB Portability** | ⭐⭐ | ⭐⭐⭐⭐ |
| **Debuggability** | ⭐⭐ | ⭐⭐⭐⭐ |

## Running Tests

```bash
cd backend
go test ./internal/leaderboard/... -v
```

All 6 tests pass with realistic scenarios covering:
- Multiple seasons
- Multiple users
- Correct/incorrect picks
- Win percentage calculations
- Edge cases

## Future Enhancements

Possible additions to the query builder:
- Filter by week
- Filter by user
- Pagination support
- Custom sort orders (by win %, username, etc.)
- Date range filtering

Example:
```go
leaderboard.NewQuery(db).
    ForSeason(seasonID).
    ForWeek(weekID).
    OrderBy("win_pct").
    Limit(10).
    Execute()
```
