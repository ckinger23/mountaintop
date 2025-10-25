# GORM Practical Examples - Find vs Scan vs Preload

## Scenario: User Picks Report

Let's walk through building a "user picks report" feature using different GORM methods.

---

## Example 1: Basic User Query (Find)

### Goal: Get all users

```go
func GetAllUsers(db *gorm.DB) ([]models.User, error) {
    var users []models.User

    // Find loads complete User models
    if err := db.Find(&users).Error; err != nil {
        return nil, err
    }

    return users, nil
}
```

**SQL Generated:**
```sql
SELECT * FROM users WHERE deleted_at IS NULL;
```

**Result:**
```json
[
  {
    "id": 1,
    "username": "alice",
    "email": "alice@example.com",
    "is_admin": false
    // Note: picks field is empty (not loaded)
  }
]
```

---

## Example 2: Users with Picks (Find + Preload)

### Goal: Get all users AND their picks

```go
func GetUsersWithPicks(db *gorm.DB) ([]models.User, error) {
    var users []models.User

    // Preload eagerly loads the Picks relationship
    if err := db.Preload("Picks").Find(&users).Error; err != nil {
        return nil, err
    }

    return users, nil
}
```

**SQL Generated:**
```sql
-- Query 1: Get users
SELECT * FROM users WHERE deleted_at IS NULL;

-- Query 2: Get all picks for those users (1 efficient query!)
SELECT * FROM picks WHERE user_id IN (1, 2, 3, 4, 5);
```

**Result:**
```json
[
  {
    "id": 1,
    "username": "alice",
    "picks": [
      {"id": 101, "game_id": 1, "points_earned": 1},
      {"id": 102, "game_id": 2, "points_earned": 1}
    ]
  }
]
```

**Performance:** 2 queries total (vs N+1 queries without Preload)

---

## Example 3: Nested Preload

### Goal: Users → Picks → Game → Teams

```go
func GetUsersWithFullPickDetails(db *gorm.DB) ([]models.User, error) {
    var users []models.User

    // Nested preloads - load picks and their related games/teams
    if err := db.
        Preload("Picks").                       // Load user's picks
        Preload("Picks.Game").                  // Load game for each pick
        Preload("Picks.Game.HomeTeam").         // Load home team for each game
        Preload("Picks.Game.AwayTeam").         // Load away team for each game
        Preload("Picks.PickedTeam").            // Load the team user picked
        Find(&users).Error; err != nil {
        return nil, err
    }

    return users, nil
}
```

**SQL Generated:**
```sql
-- Query 1: Users
SELECT * FROM users WHERE deleted_at IS NULL;

-- Query 2: Picks
SELECT * FROM picks WHERE user_id IN (1, 2, 3);

-- Query 3: Games
SELECT * FROM games WHERE id IN (1, 2, 3, 4);

-- Query 4: Home Teams
SELECT * FROM teams WHERE id IN (1, 3, 5);

-- Query 5: Away Teams
SELECT * FROM teams WHERE id IN (2, 4, 6);

-- Query 6: Picked Teams
SELECT * FROM teams WHERE id IN (1, 1, 2, 3);
```

**Result:**
```json
[
  {
    "id": 1,
    "username": "alice",
    "picks": [
      {
        "id": 101,
        "game": {
          "id": 1,
          "home_team": {"name": "Alabama"},
          "away_team": {"name": "Georgia"}
        },
        "picked_team": {"name": "Alabama"}
      }
    ]
  }
]
```

**Performance:** 6 queries total (regardless of how many users/picks!)

---

## Example 4: Conditional Preload

### Goal: Users with only CORRECT picks

```go
func GetUsersWithCorrectPicks(db *gorm.DB) ([]models.User, error) {
    var users []models.User

    // Preload with a condition
    if err := db.
        Preload("Picks", "is_correct = ?", true).
        Find(&users).Error; err != nil {
        return nil, err
    }

    return users, nil
}
```

**SQL Generated:**
```sql
-- Query 1: Users
SELECT * FROM users WHERE deleted_at IS NULL;

-- Query 2: Only correct picks
SELECT * FROM picks WHERE user_id IN (1, 2, 3) AND is_correct = true;
```

---

## Example 5: Aggregation with Scan

### Goal: Count picks per user

```go
type UserPickCount struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    PickCount int   `json:"pick_count"`
}

func GetUserPickCounts(db *gorm.DB) ([]UserPickCount, error) {
    var results []UserPickCount

    // Scan is required for aggregate functions
    if err := db.
        Table("users u").
        Select("u.id as user_id, u.username, COUNT(p.id) as pick_count").
        Joins("LEFT JOIN picks p ON u.id = p.user_id").
        Group("u.id").
        Scan(&results).Error; err != nil {
        return nil, err
    }

    return results, nil
}
```

**SQL Generated:**
```sql
SELECT u.id as user_id, u.username, COUNT(p.id) as pick_count
FROM users u
LEFT JOIN picks p ON u.id = p.user_id
GROUP BY u.id;
```

**Result:**
```json
[
  {"user_id": 1, "username": "alice", "pick_count": 15},
  {"user_id": 2, "username": "bob", "pick_count": 12}
]
```

**Why Scan?** Can't use Find because we're using COUNT() aggregate function.

---

## Example 6: Complex Report with Multiple Joins (Scan)

### Goal: Weekly leaderboard with game details

```go
type WeeklyReport struct {
    Username      string  `json:"username"`
    WeekNumber    int     `json:"week_number"`
    CorrectPicks  int     `json:"correct_picks"`
    TotalPicks    int     `json:"total_picks"`
    WinPercentage float64 `json:"win_percentage"`
}

func GetWeeklyReport(db *gorm.DB, seasonID uint) ([]WeeklyReport, error) {
    var results []WeeklyReport

    if err := db.
        Table("users u").
        Select(`
            u.username,
            w.week_number,
            SUM(CASE WHEN p.is_correct = 1 THEN 1 ELSE 0 END) as correct_picks,
            COUNT(p.id) as total_picks,
            CAST(SUM(CASE WHEN p.is_correct = 1 THEN 1 ELSE 0 END) AS FLOAT) /
                NULLIF(COUNT(p.id), 0) as win_percentage
        `).
        Joins("LEFT JOIN picks p ON u.id = p.user_id").
        Joins("LEFT JOIN games g ON p.game_id = g.id").
        Joins("LEFT JOIN weeks w ON g.week_id = w.id").
        Where("w.season_id = ?", seasonID).
        Group("u.id, w.week_number").
        Order("w.week_number ASC, correct_picks DESC").
        Scan(&results).Error; err != nil {
        return nil, err
    }

    return results, nil
}
```

**Why Scan?**
- Multiple JOINs
- Aggregate functions (SUM, COUNT)
- Custom GROUP BY
- Result doesn't map to a single GORM model

---

## Performance Comparison

### The N+1 Problem in Action

#### ❌ WITHOUT Preload (Bad)

```go
var users []models.User
db.Find(&users) // 1 query

for _, user := range users {
    // Each iteration triggers a new query!
    fmt.Printf("%s has %d picks\n", user.Username, len(user.Picks))
}

// Total queries: 1 + N (where N = number of users)
// For 100 users: 101 queries! 🐌
```

#### ✅ WITH Preload (Good)

```go
var users []models.User
db.Preload("Picks").Find(&users) // 2 queries total

for _, user := range users {
    // No additional queries - data already loaded!
    fmt.Printf("%s has %d picks\n", user.Username, len(user.Picks))
}

// Total queries: 2 (users + picks)
// For 100 users: Still just 2 queries! ⚡
```

---

## When to Use What?

### Use `.Find()` when:
- ✅ Querying a GORM model (`User`, `Game`, `Pick`)
- ✅ You want soft delete filtering
- ✅ You need GORM hooks to fire
- ✅ You're doing basic CRUD operations

**Example:** Fetch all games for a week
```go
db.Where("week_id = ?", weekID).Find(&games)
```

---

### Use `.Find()` + `.Preload()` when:
- ✅ You'll access relationship fields
- ✅ You want to avoid N+1 queries
- ✅ You're serializing to JSON with nested objects

**Example:** Game listing with team names
```go
db.Preload("HomeTeam").Preload("AwayTeam").Find(&games)
json.Marshal(games) // Needs HomeTeam.Name, AwayTeam.Name
```

---

### Use `.Scan()` when:
- ✅ Using aggregate functions (SUM, COUNT, AVG, MIN, MAX)
- ✅ Custom SELECT with computed fields
- ✅ Complex JOINs
- ✅ Mapping to a DTO (non-model struct)
- ✅ Raw SQL or complex queries

**Example:** Leaderboard with aggregations
```go
db.Table("users").
   Select("username, SUM(points) as total").
   Joins("LEFT JOIN picks ON...").
   Group("username").
   Scan(&results)
```

---

## Common Mistakes

### ❌ Mistake 1: Using Find for Aggregates

```go
// This will ERROR!
var result struct{ Total int }
db.Select("COUNT(*)").Find(&result)
// Error: Can't map to model fields
```

**Fix:** Use Scan
```go
db.Model(&models.User{}).Select("COUNT(*)").Scan(&result)
```

---

### ❌ Mistake 2: Not Preloading When Needed

```go
// This causes N+1 queries!
db.Find(&games)
for _, game := range games {
    log.Println(game.HomeTeam.Name) // New query each time!
}
```

**Fix:** Add Preload
```go
db.Preload("HomeTeam").Find(&games)
```

---

### ❌ Mistake 3: Over-Preloading

```go
// Loading too much data!
db.Preload("Picks").              // 100 picks
   Preload("Picks.Game").          // 100 games
   Preload("Picks.Game.Week").     // 10 weeks
   Preload("Picks.Game.Week.Season"). // 1 season
   Find(&users)
// Memory bloat for data you don't need!
```

**Fix:** Only preload what you'll use
```go
db.Preload("Picks", "is_correct = ?", true).
   Preload("Picks.PickedTeam").
   Find(&users)
```

---

## Testing Tips

### Mock Database Queries

```go
func TestGetUsersWithPicks(t *testing.T) {
    // Setup in-memory database
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&models.User{}, &models.Pick{})

    // Seed test data
    user := models.User{Username: "test"}
    db.Create(&user)
    db.Create(&models.Pick{UserID: user.ID, PointsEarned: 1})

    // Test with Preload
    var users []models.User
    db.Preload("Picks").Find(&users)

    assert.Equal(t, 1, len(users[0].Picks))
}
```

---

## Summary Table

| Operation | Method | Preload? | Use Case |
|-----------|--------|----------|----------|
| Get all games | `.Find()` | No | Basic listing |
| Games with teams | `.Find()` | Yes | JSON with nested data |
| Leaderboard stats | `.Scan()` | No | Aggregations (SUM/COUNT) |
| Weekly report | `.Scan()` | No | Complex JOINs + GROUP BY |
| User with picks | `.Find()` | Yes | Avoid N+1 queries |

**Key Principle:**
- **Find** = GORM manages everything (models, relationships)
- **Scan** = You manage everything (custom queries, DTOs)
- **Preload** = Eager load relationships to avoid N+1 queries
