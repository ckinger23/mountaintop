# GORM Concepts: Find vs Scan vs Preload

## 1. `.Find()` vs `.Scan()`

### `.Find()` - Maps to GORM Models

**What it does:**
- Maps database rows to **GORM model structs** (with all GORM features)
- Respects GORM field tags, hooks, associations
- Populates all model fields including relationships
- Works with GORM's tracking (soft deletes, timestamps, etc.)

**Example from your code ([games.go:27](backend/internal/handlers/games.go#L27)):**

```go
var games []models.Game
query := a.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week")

if err := query.Find(&games).Error; err != nil {
    // Handle error
}
```

**SQL Generated:**
```sql
SELECT * FROM games WHERE deleted_at IS NULL;
-- Then additional queries for each Preload
```

**Key Features:**
- ✅ Respects `gorm:"primaryKey"` tags
- ✅ Respects `DeletedAt` (soft deletes)
- ✅ Can use Preload for relationships
- ✅ Triggers GORM hooks (AfterFind, etc.)
- ✅ Handles `json:"-"` tags
- ❌ Can't use custom SELECT with aggregate functions

---

### `.Scan()` - Maps to Any Struct

**What it does:**
- Maps database rows to **any struct** (not necessarily a GORM model)
- Ignores GORM tags, hooks, associations
- Direct mapping from query result to struct fields
- Used for custom queries, aggregations, JOINs

**Example from your code ([leaderboard.go:55](backend/internal/leaderboard/leaderboard.go#L55)):**

```go
var results []models.LeaderboardEntry

query := db.Table("users u").
    Select(`
        u.id as user_id,
        u.username,
        COALESCE(SUM(p.points_earned), 0) as total_points,
        COUNT(p.id) as total_picks
    `).
    Joins("LEFT JOIN picks p ON u.id = p.user_id").
    Group("u.id")

if err := query.Scan(&results).Error; err != nil {
    return nil, err
}
```

**SQL Generated:**
```sql
SELECT
    u.id as user_id,
    u.username,
    COALESCE(SUM(p.points_earned), 0) as total_points,
    COUNT(p.id) as total_picks
FROM users u
LEFT JOIN picks p ON u.id = p.user_id
GROUP BY u.id
```

**Key Features:**
- ✅ Works with aggregate functions (SUM, COUNT, AVG)
- ✅ Works with custom JOINs
- ✅ Can map to non-model structs (LeaderboardEntry)
- ✅ Bypasses GORM hooks
- ❌ Can't use Preload
- ❌ Ignores soft delete filters (unless you add them manually)

---

## Comparison Table

| Feature | `.Find()` | `.Scan()` |
|---------|-----------|-----------|
| **Target** | GORM models | Any struct |
| **Use Case** | Standard CRUD | Custom queries, aggregations |
| **Relationships** | ✅ Via Preload | ❌ Must JOIN manually |
| **Soft Deletes** | ✅ Automatic | ❌ Manual WHERE needed |
| **Hooks** | ✅ Triggers | ❌ Bypassed |
| **SELECT** | `SELECT *` (all fields) | Custom SELECT with aggregates |
| **Type Safety** | ✅ Full | ✅ Full |

---

## 2. `.Preload()` - Eager Loading Relationships

**What it does:**
- Loads related data from other tables
- Solves the **N+1 query problem**
- Only works with `.Find()`, `.First()`, etc. (not `.Scan()`)

### Without Preload (N+1 Problem) ❌

```go
var games []models.Game
db.Find(&games) // 1 query

for _, game := range games {
    // Each access triggers a separate query!
    fmt.Println(game.HomeTeam.Name) // Query #2
    fmt.Println(game.AwayTeam.Name) // Query #3
    fmt.Println(game.Week.Name)     // Query #4
}
// Total: 1 + (3 * N) queries for N games!
```

### With Preload (Efficient) ✅

**Example from your code ([games.go:21](backend/internal/handlers/games.go#L21)):**

```go
var games []models.Game
db.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").Find(&games)

for _, game := range games {
    // No additional queries - data is already loaded!
    fmt.Println(game.HomeTeam.Name)
    fmt.Println(game.AwayTeam.Name)
    fmt.Println(game.Week.Name)
}
// Total: 4 queries (1 for games, 1 each for teams and weeks)
```

**SQL Generated:**
```sql
-- Query 1: Main query
SELECT * FROM games WHERE deleted_at IS NULL;

-- Query 2: Preload HomeTeam
SELECT * FROM teams WHERE id IN (1, 3, 5, 7, ...);

-- Query 3: Preload AwayTeam
SELECT * FROM teams WHERE id IN (2, 4, 6, 8, ...);

-- Query 4: Preload Week
SELECT * FROM weeks WHERE id IN (1, 1, 1, 2, ...);
```

### How Preload Knows What to Load

Looking at your [models.go](backend/internal/models/models.go#L90-L93):

```go
type Game struct {
    HomeTeamID  uint       `gorm:"not null" json:"home_team_id"`  // Foreign key
    AwayTeamID  uint       `gorm:"not null" json:"away_team_id"`  // Foreign key

    // Relationships - GORM uses these for Preload
    HomeTeam Team   `gorm:"foreignKey:HomeTeamID" json:"home_team,omitempty"`
    AwayTeam Team   `gorm:"foreignKey:AwayTeamID" json:"away_team,omitempty"`
}
```

When you call `Preload("HomeTeam")`:
1. GORM finds the `HomeTeam` field on the Game struct
2. Reads the `foreignKey:HomeTeamID` tag
3. Collects all unique `HomeTeamID` values from loaded games
4. Runs: `SELECT * FROM teams WHERE id IN (...)`
5. Populates each `game.HomeTeam` with the matching Team

---

## 3. Real-World Examples from Your Codebase

### Example 1: Games with Relationships (Find + Preload)

**Location:** [games.go:15-35](backend/internal/handlers/games.go#L15-L35)

```go
func GetGames(a *app.App) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var games []models.Game

        // Preload prevents N+1 queries when serializing to JSON
        query := a.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week")

        if weekID != "" {
            query = query.Where("week_id = ?", weekID)
        }

        // Find() is perfect here - we want full Game models with relationships
        if err := query.Find(&games).Error; err != nil {
            http.Error(w, "Error fetching games", http.StatusInternalServerError)
            return
        }

        // JSON encoder can access HomeTeam, AwayTeam, Week without extra queries
        json.NewEncoder(w).Encode(games)
    }
}
```

**Why Find?** We want complete Game models with all GORM features (soft deletes, relationships).

**Why Preload?** The JSON encoder will access `HomeTeam`, `AwayTeam`, `Week` fields. Without Preload, that would trigger dozens of queries.

---

### Example 2: Leaderboard Aggregations (Scan without Preload)

**Location:** [leaderboard.go:26-58](backend/internal/leaderboard/leaderboard.go#L26-L58)

```go
func (q *Query) Execute() ([]models.LeaderboardEntry, error) {
    var results []models.LeaderboardEntry

    query := q.db.Table("users u").
        Select(`
            u.id as user_id,
            u.username,
            COALESCE(SUM(p.points_earned), 0) as total_points,
            COUNT(p.id) as total_picks
        `).
        Joins("LEFT JOIN picks p ON u.id = p.user_id").
        Group("u.id")

    // Scan() is perfect here - LeaderboardEntry is not a GORM model
    if err := query.Scan(&results).Error; err != nil {
        return nil, err
    }

    return results, nil
}
```

**Why Scan?**
- We're using aggregate functions (SUM, COUNT)
- LeaderboardEntry is a DTO (Data Transfer Object), not a full model
- We don't need GORM features (no relationships, hooks, soft deletes)

**Why no Preload?**
- Can't use Preload with custom SELECT statements
- LeaderboardEntry doesn't have relationships to load

---

## 4. Common Pitfalls

### ❌ Pitfall 1: Using Find with Aggregates

```go
// This WON'T work!
var results []LeaderboardEntry
db.Select("username, COUNT(*) as total").
   Group("username").
   Find(&results) // Error: Can't map COUNT(*) to User model fields
```

**Solution:** Use `.Scan()` for aggregates.

---

### ❌ Pitfall 2: Forgetting Preload (N+1 Queries)

```go
// This works but triggers HUNDREDS of queries!
var picks []models.Pick
db.Find(&picks)

for _, pick := range picks {
    fmt.Println(pick.User.Username)      // Query per pick!
    fmt.Println(pick.PickedTeam.Name)    // Query per pick!
}
```

**Solution:** Add Preloads:
```go
db.Preload("User").Preload("PickedTeam").Find(&picks)
```

---

### ❌ Pitfall 3: Using Preload with Scan

```go
// This silently IGNORES Preload!
var results []LeaderboardEntry
db.Table("users").
   Preload("Picks"). // IGNORED!
   Scan(&results)
```

**Solution:** Use Joins if you need related data with Scan.

---

## 5. Decision Tree

```
Need to query data?
│
├─ Do you need aggregate functions (SUM, COUNT, AVG)?
│  └─ YES → Use .Scan()
│
├─ Do you need custom JOINs with complex logic?
│  └─ YES → Use .Scan() with Table() and Joins()
│
├─ Are you querying a GORM model with relationships?
│  ├─ Will you access relationship fields (game.HomeTeam)?
│  │  └─ YES → Use .Find() with .Preload()
│  └─ Just the main model fields?
│     └─ Use .Find() (no Preload needed)
│
└─ Default: Use .Find() for GORM models
```

---

## 6. Performance Tips

### ✅ Use Preload for relationships you'll access
```go
// Good: Only preload what you need
db.Preload("HomeTeam").Find(&games)
```

### ❌ Don't preload everything blindly
```go
// Bad: Loading data you won't use
db.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").
   Preload("Week.Season").Preload("Picks"). // Overkill!
   Find(&games)
```

### ✅ Use Scan for read-only aggregations
```go
// Good: Efficient for reports/stats
db.Table("picks").
   Select("user_id, COUNT(*) as total").
   Group("user_id").
   Scan(&results)
```

### ✅ Use Preload conditions to filter related data
```go
// Only preload picks that are correct
db.Preload("Picks", "is_correct = ?", true).Find(&users)
```

---

## Summary

| Method | When to Use | Example |
|--------|-------------|---------|
| **Find()** | Standard CRUD on GORM models | `db.Find(&games)` |
| **Scan()** | Custom queries, aggregations | `db.Select("COUNT(*)").Scan(&result)` |
| **Preload()** | Load relationships efficiently | `db.Preload("HomeTeam").Find(&games)` |

**Golden Rule:**
- Use **Find** when you want GORM to manage the query (models, relationships, hooks)
- Use **Scan** when you want full control (custom SELECT, JOINs, aggregates)
- Use **Preload** whenever you'll access relationship fields to avoid N+1 queries
