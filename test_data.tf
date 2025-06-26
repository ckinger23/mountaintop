# Users - Primary records
resource "aws_dynamodb_table_item" "users" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    "1" = { username = "Carter", email = "carter@example.com" }
    "2" = { username = "Paul", email = "paul@example.com" }
    "3" = { username = "Cal", email = "cal@example.com" }
    "4" = { username = "Nathan", email = "nathan@example.com" }
    "5" = { username = "Nick", email = "nick@example.com" }
  }

  item = jsonencode({
    "PK"           = { "S" = "USER#${each.key}" },
    "SK"           = { "S" = "PROFILE" },
    "entity_type"  = { "S" = "USER" },
    "GSI1_PK"      = { "S" = "USER" },
    "GSI1_SK"      = { "S" = each.key },
    "data" = {
      "M" = {
        "user_id"   = { "S" = each.key },
        "username"  = { "S" = each.value.username },
        "email"     = { "S" = each.value.email },
        "role"      = { "S" = "player" }
      }
    }
  })
}

# User lookup by username
resource "aws_dynamodb_table_item" "user_username_lookup" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    "1" = { username = "Carter", email = "carter@example.com" }
    "2" = { username = "Paul", email = "paul@example.com" }
    "3" = { username = "Cal", email = "cal@example.com" }
    "4" = { username = "Nathan", email = "nathan@example.com" }
    "5" = { username = "Nick", email = "nick@example.com" }
  }

  item = jsonencode({
    "PK"           = { "S" = "USERNAME#${each.value.username}" },
    "SK"           = { "S" = "LOOKUP" },
    "GSI1_PK"      = { "S" = "USERNAME_LOOKUP" },
    "GSI1_SK"      = { "S" = each.value.username },
    "user_id"      = { "S" = each.key }
  })
}

# User lookup by email
resource "aws_dynamodb_table_item" "user_email_lookup" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    "1" = { username = "Carter", email = "carter@example.com" }
    "2" = { username = "Paul", email = "paul@example.com" }
    "3" = { username = "Cal", email = "cal@example.com" }
    "4" = { username = "Nathan", email = "nathan@example.com" }
    "5" = { username = "Nick", email = "nick@example.com" }
  }

  item = jsonencode({
    "PK"           = { "S" = "EMAIL#${each.value.email}" },
    "SK"           = { "S" = "LOOKUP" },
    "GSI1_PK"      = { "S" = "EMAIL_LOOKUP" },
    "GSI1_SK"      = { "S" = each.value.email },
    "user_id"      = { "S" = each.key }
  })
}

  item = jsonencode({
    "PK"           = { "S" = "USER#${each.key}" },
    "SK"           = { "S" = "PROFILE" },
    "entity_type"  = { "S" = "USER" },
    "GSI1_PK"      = { "S" = "USER" },
    "GSI1_SK"      = { "S" = each.key },
    "data" = {
      "M" = {
        "user_id"   = { "S" = each.key },
        "username"  = { "S" = each.value.username },
        "email"     = { "S" = each.value.email },
        "role"      = { "S" = "player" }
      }
    }
  })
}

# Conferences
resource "aws_dynamodb_table_item" "conferences" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    "1" = { name = "SEC", abbreviation = "SEC" }
    "2" = { name = "Big Ten", abbreviation = "B1G" }
    "3" = { name = "Big 12", abbreviation = "B12" }
    "4" = { name = "ACC", abbreviation = "ACC" }
  }

  item = jsonencode({
    "PK"          = { "S" = "CONFERENCE#${each.key}" },
    "SK"          = { "S" = "METADATA" },
    "entity_type" = { "S" = "CONFERENCE" },
    "GSI1_PK"     = { "S" = "CONFERENCE" },
    "GSI1_SK"     = { "S" = each.key },
    "data" = {
      "M" = {
        "conference_id" = { "S" = each.key },
        "name"          = { "S" = each.value.name },
        "abbreviation"  = { "S" = each.value.abbreviation }
      }
    }
  })
}

# Conference lookup by name
resource "aws_dynamodb_table_item" "conference_name_lookup" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    "1" = { name = "SEC", abbreviation = "SEC" }
    "2" = { name = "Big Ten", abbreviation = "B1G" }
    "3" = { name = "Big 12", abbreviation = "B12" }
    "4" = { name = "ACC", abbreviation = "ACC" }
  }

  item = jsonencode({
    "PK"           = { "S" = "CONFERENCE_NAME#${each.value.name}" },
    "SK"           = { "S" = "LOOKUP" },
    "GSI1_PK"      = { "S" = "CONFERENCE_NAME_LOOKUP" },
    "GSI1_SK"      = { "S" = each.value.name },
    "conference_id" = { "S" = each.key }
  })
}

# Teams
resource "aws_dynamodb_table_item" "teams" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    # SEC
    "1"  = { name = "Georgia", conference_id = "1" }
    "2"  = { name = "Alabama", conference_id = "1" }
    "3"  = { name = "Auburn", conference_id = "1" }
    "4"  = { name = "Tennessee", conference_id = "1" }
    # Big Ten
    "5"  = { name = "Ohio State", conference_id = "2" }
    "6"  = { name = "Oregon", conference_id = "2" }
    "7"  = { name = "Michigan", conference_id = "2" }
    "8"  = { name = "Illinois", conference_id = "2" }
    # Big 12
    "9"  = { name = "Utah", conference_id = "3" }
    "10" = { name = "Oklahoma State", conference_id = "3" }
    "11" = { name = "Arizona State", conference_id = "3" }
    "12" = { name = "Arizona", conference_id = "3" }
    # ACC
    "13" = { name = "Clemson", conference_id = "4" }
    "14" = { name = "Florida State", conference_id = "4" }
    "15" = { name = "SMU", conference_id = "4" }
    "16" = { name = "Miami", conference_id = "4" }
  }

  item = jsonencode({
    "PK"           = { "S" = "TEAM#${each.key}" },
    "SK"           = { "S" = "METADATA" },
    "entity_type"  = { "S" = "TEAM" },
    "GSI1_PK"      = { "S" = "TEAM" },
    "GSI1_SK"      = { "S" = each.key },
    "data" = {
      "M" = {
        "team_id"       = { "S" = each.key },
        "name"          = { "S" = each.value.name },
        "conference_id" = { "S" = each.value.conference_id }
      }
    }
  })
}

# Team lookup by name
resource "aws_dynamodb_table_item" "team_name_lookup" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    # SEC
    "1"  = { name = "Georgia", conference_id = "1" }
    "2"  = { name = "Alabama", conference_id = "1" }
    "3"  = { name = "Auburn", conference_id = "1" }
    "4"  = { name = "Tennessee", conference_id = "1" }
    # Big Ten
    "5"  = { name = "Ohio State", conference_id = "2" }
    "6"  = { name = "Oregon", conference_id = "2" }
    "7"  = { name = "Michigan", conference_id = "2" }
    "8"  = { name = "Illinois", conference_id = "2" }
    # Big 12
    "9"  = { name = "Utah", conference_id = "3" }
    "10" = { name = "Oklahoma State", conference_id = "3" }
    "11" = { name = "Arizona State", conference_id = "3" }
    "12" = { name = "Arizona", conference_id = "3" }
    # ACC
    "13" = { name = "Clemson", conference_id = "4" }
    "14" = { name = "Florida State", conference_id = "4" }
    "15" = { name = "SMU", conference_id = "4" }
    "16" = { name = "Miami", conference_id = "4" }
  }

  item = jsonencode({
    "PK"           = { "S" = "TEAM_NAME#${each.value.name}" },
    "SK"           = { "S" = "LOOKUP" },
    "GSI1_PK"      = { "S" = "TEAM_NAME_LOOKUP" },
    "GSI1_SK"      = { "S" = each.value.name },
    "team_id"      = { "S" = each.key }
  })
}

# Teams by conference relationship
resource "aws_dynamodb_table_item" "teams_by_conference" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    # SEC
    "1"  = { name = "Georgia", conference_id = "1" }
    "2"  = { name = "Alabama", conference_id = "1" }
    "3"  = { name = "Auburn", conference_id = "1" }
    "4"  = { name = "Tennessee", conference_id = "1" }
    # Big Ten
    "5"  = { name = "Ohio State", conference_id = "2" }
    "6"  = { name = "Oregon", conference_id = "2" }
    "7"  = { name = "Michigan", conference_id = "2" }
    "8"  = { name = "Illinois", conference_id = "2" }
    # Big 12
    "9"  = { name = "Utah", conference_id = "3" }
    "10" = { name = "Oklahoma State", conference_id = "3" }
    "11" = { name = "Arizona State", conference_id = "3" }
    "12" = { name = "Arizona", conference_id = "3" }
    # ACC
    "13" = { name = "Clemson", conference_id = "4" }
    "14" = { name = "Florida State", conference_id = "4" }
    "15" = { name = "SMU", conference_id = "4" }
    "16" = { name = "Miami", conference_id = "4" }
  }

  item = jsonencode({
    "PK"           = { "S" = "CONFERENCE#${each.value.conference_id}" },
    "SK"           = { "S" = "TEAM#${each.key}" },
    "entity_type"  = { "S" = "CONFERENCE_TEAM" },
    "GSI1_PK"      = { "S" = "CONFERENCE_TEAM" },
    "GSI1_SK"      = { "S" = "${each.value.conference_id}#${each.key}" },
    "data" = {
      "M" = {
        "conference_id" = { "S" = each.value.conference_id },
        "team_id"       = { "S" = each.key },
        "team_name"     = { "S" = each.value.name }
      }
    }
  })
}

# League
resource "aws_dynamodb_table_item" "league" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  item = jsonencode({
    "PK"          = { "S" = "LEAGUE#1" },
    "SK"          = { "S" = "METADATA" },
    "entity_type" = { "S" = "LEAGUE" },
    "GSI1_PK"     = { "S" = "LEAGUE" },
    "GSI1_SK"     = { "S" = "1" },
    "data" = {
      "M" = {
        "league_id"   = { "S" = "1" },
        "name"        = { "S" = "LocalLeague" },
        "description" = { "S" = "Local test league" },
        "season"      = { "S" = "2023" }
      }
    }
  })
}

# League Members
resource "aws_dynamodb_table_item" "league_members" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = toset(["1", "2", "3", "4", "5"]) # User IDs

  item = jsonencode({
    "PK"          = { "S" = "LEAGUE#1" },
    "SK"          = { "S" = "MEMBER#${each.key}" },
    "entity_type" = { "S" = "LEAGUE_MEMBER" },
    "GSI1_PK"     = { "S" = "LEAGUE_MEMBER" },
    "GSI1_SK"     = { "S" = "1#${each.key}" },
    "data" = {
      "M" = {
        "league_id"  = { "S" = "1" },
        "user_id"    = { "S" = each.key },
        "joined_at"  = { "S" = "2023-09-01T00:00:00Z" }
      }
    }
  })
}

# Games (15 games across 5 weeks)
resource "aws_dynamodb_table_item" "games" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    # Week 1
    "1" = { home_team = "1", away_team = "2", week = 1, date = "2023-09-02T19:30:00Z" }  # Georgia vs Alabama
    "2" = { home_team = "5", away_team = "6", week = 1, date = "2023-09-02T20:00:00Z" }  # Ohio State vs Oregon
    "3" = { home_team = "9", away_team = "10", week = 1, date = "2023-09-02T20:30:00Z" } # Utah vs Oklahoma State
    
    # Week 2
    "4" = { home_team = "3", away_team = "4", week = 2, date = "2023-09-09T19:30:00Z" }  # Auburn vs Tennessee
    "5" = { home_team = "7", away_team = "8", week = 2, date = "2023-09-09T20:00:00Z" }  # Michigan vs Illinois
    "6" = { home_team = "11", away_team = "12", week = 2, date = "2023-09-09T20:30:00Z" } # Arizona State vs Arizona
    
    # Week 3
    "7" = { home_team = "13", away_team = "14", week = 3, date = "2023-09-16T19:30:00Z" } # Clemson vs Florida State
    "8" = { home_team = "15", away_team = "16", week = 3, date = "2023-09-16T20:00:00Z" } # SMU vs Miami
    "9" = { home_team = "1", away_team = "3", week = 3, date = "2023-09-16T20:30:00Z" }   # Georgia vs Auburn
    
    # Week 4
    "10" = { home_team = "6", away_team = "7", week = 4, date = "2023-09-23T19:30:00Z" }  # Oregon vs Michigan
    "11" = { home_team = "10", away_team = "11", week = 4, date = "2023-09-23T20:00:00Z" } # Oklahoma State vs Arizona State
    "12" = { home_team = "14", away_team = "15", week = 4, date = "2023-09-23T20:30:00Z" } # Florida State vs SMU
    
    # Week 5
    "13" = { home_team = "2", away_team = "4", week = 5, date = "2023-09-30T19:30:00Z" }  # Alabama vs Tennessee
    "14" = { home_team = "5", away_team = "7", week = 5, date = "2023-09-30T20:00:00Z" }  # Ohio State vs Michigan
    "15" = { home_team = "13", away_team = "16", week = 5, date = "2023-09-30T20:30:00Z" } # Clemson vs Miami
  }

  item = jsonencode({
    "PK"           = { "S" = "GAME#${each.key}" },
    "SK"           = { "S" = "METADATA" },
    "entity_type"  = { "S" = "GAME" },
    "GSI1_PK"      = { "S" = "GAME" },
    "GSI1_SK"      = { "S" = each.key },
    "GSI2_PK"      = { "S" = "LEAGUE#1#WEEK#${each.value.week}" },
    "GSI2_SK"      = { "S" = "GAME#${each.key}" },
    "data" = {
      "M" = {
        "game_id"      = { "S" = each.key },
        "league_id"    = { "S" = "1" },
        "home_team_id" = { "S" = each.value.home_team },
        "away_team_id" = { "S" = each.value.away_team },
        "game_time"    = { "S" = each.value.date },
        "game_date"    = { "S" = substr(each.value.date, 0, 10) },
        "week"         = { "N" = tostring(each.value.week) },
        "season"       = { "N" = "2023" },
        "status"       = { "S" = "SCHEDULED" },
        "home_score"   = { "N" = "0" },
        "away_score"   = { "N" = "0" }
      }
    }
  })
}

# Games by date lookup
resource "aws_dynamodb_table_item" "games_by_date" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  for_each = {
    # Week 1
    "1" = { home_team = "1", away_team = "2", week = 1, date = "2023-09-02T19:30:00Z" }  # Georgia vs Alabama
    "2" = { home_team = "5", away_team = "6", week = 1, date = "2023-09-02T20:00:00Z" }  # Ohio State vs Oregon
    "3" = { home_team = "9", away_team = "10", week = 1, date = "2023-09-02T20:30:00Z" } # Utah vs Oklahoma State
    
    # Week 2
    "4" = { home_team = "3", away_team = "4", week = 2, date = "2023-09-09T19:30:00Z" }  # Auburn vs Tennessee
    "5" = { home_team = "7", away_team = "8", week = 2, date = "2023-09-09T20:00:00Z" }  # Michigan vs Illinois
    "6" = { home_team = "11", away_team = "12", week = 2, date = "2023-09-09T20:30:00Z" } # Arizona State vs Arizona
    
    # Week 3
    "7" = { home_team = "13", away_team = "14", week = 3, date = "2023-09-16T19:30:00Z" } # Clemson vs Florida State
    "8" = { home_team = "15", away_team = "16", week = 3, date = "2023-09-16T20:00:00Z" } # SMU vs Miami
    "9" = { home_team = "1", away_team = "3", week = 3, date = "2023-09-16T20:30:00Z" }   # Georgia vs Auburn
    
    # Week 4
    "10" = { home_team = "6", away_team = "7", week = 4, date = "2023-09-23T19:30:00Z" }  # Oregon vs Michigan
    "11" = { home_team = "10", away_team = "11", week = 4, date = "2023-09-23T20:00:00Z" } # Oklahoma State vs Arizona State
    "12" = { home_team = "14", away_team = "15", week = 4, date = "2023-09-23T20:30:00Z" } # Florida State vs SMU
    
    # Week 5
    "13" = { home_team = "2", away_team = "4", week = 5, date = "2023-09-30T19:30:00Z" }  # Alabama vs Tennessee
    "14" = { home_team = "5", away_team = "7", week = 5, date = "2023-09-30T20:00:00Z" }  # Ohio State vs Michigan
    "15" = { home_team = "13", away_team = "16", week = 5, date = "2023-09-30T20:30:00Z" } # Clemson vs Miami
  }

  item = jsonencode({
    "PK"           = { "S" = "DATE#${substr(each.value.date, 0, 10)}#LEAGUE#1" },
    "SK"           = { "S" = "GAME#${each.key}" },
    "entity_type"  = { "S" = "GAME_DATE" },
    "GSI1_PK"      = { "S" = "GAME_DATE" },
    "GSI1_SK"      = { "S" = "${substr(each.value.date, 0, 10)}#1#${each.key}" },
    "data" = {
      "M" = {
        "game_id"    = { "S" = each.key },
        "league_id"  = { "S" = "1" },
        "game_date"  = { "S" = substr(each.value.date, 0, 10) },
        "game_time"  = { "S" = each.value.date }
      }
    }
  })
}

# Picks (each user picks a random team for each game)
resource "aws_dynamodb_table_item" "picks" {
  provider = aws.localstack
  table_name = aws_dynamodb_table.football_league.name
  hash_key   = "PK"
  range_key  = "SK"

  # This creates a pick for each user for each game
  for_each = {
    for idx, pick in flatten([
      for user_id in ["1", "2", "3", "4", "5"] : [
        for game_id in range(1, 16) : {
          user_id    = user_id
          game_id    = tostring(game_id)
          pick       = random_shuffle.teams[game_id].result[tonumber(user_id) % 2 == 0 ? 0 : 1] # Alternate picks
          created_at = "2023-09-01T00:00:00Z"
        }
      ]
    ]) : "${pick.user_id}-${pick.game_id}" => pick
  }

  item = jsonencode({
    "PK"           = { "S" = "PICK#1#2023#${each.value.user_id}" },
    "SK"           = { "S" = "GAME#${each.value.game_id}" },
    "entity_type"  = { "S" = "PICK" },
    "GSI1_PK"      = { "S" = "PICK" },
    "GSI1_SK"      = { "S" = "1#2023#${each.value.user_id}#${each.value.game_id}" },
    "GSI2_PK"      = { "S" = "LEAGUE#1#WEEK#${local.game_weeks[each.value.game_id]}" },
    "GSI2_SK"      = { "S" = "USER#${each.value.user_id}#GAME#${each.value.game_id}" },
    "GSI3_PK"      = { "S" = "USER#${each.value.user_id}#2023" },
    "GSI3_SK"      = { "S" = "WEEK#${local.game_weeks[each.value.game_id]}#GAME#${each.value.game_id}" },
    "GSI4_PK"      = { "S" = "GAME#${each.value.game_id}" },
    "GSI4_SK"      = { "S" = "USER#${each.value.user_id}" },
    "data" = {
      "M" = {
        "user_id"    = { "S" = each.value.user_id },
        "game_id"    = { "S" = each.value.game_id },
        "league_id"  = { "S" = "1" },
        "season"     = { "S" = "2023" },
        "week"       = { "N" = tostring(local.game_weeks[each.value.game_id]) },
        "pick"       = { "S" = each.value.pick },
        "status"     = { "S" = "PENDING" },
        "points"     = { "N" = "0" },
        "created_at" = { "S" = each.value.created_at },
        "updated_at" = { "S" = each.value.created_at }
      }
    }
  })
}

# Helper to get game weeks
locals {
  game_weeks = {
    "1" = 1, "2" = 1, "3" = 1,
    "4" = 2, "5" = 2, "6" = 2,
    "7" = 3, "8" = 3, "9" = 3,
    "10" = 4, "11" = 4, "12" = 4,
    "13" = 5, "14" = 5, "15" = 5
  }
}

# Random team selector for picks
resource "random_shuffle" "teams" {
  for_each = toset([for i in range(1, 16) : tostring(i)]) # For each game
  input = ["home", "away"]
  result_count = 2
}
