provider "aws" {
  alias                   = "localstack"
  region                  = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  
  endpoints {
    dynamodb = "http://localhost:4566"
  }
}

resource "aws_dynamodb_table" "football_league" {
  provider     = aws.localstack
  name         = "FootballLeague"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  # Define attributes used in the table
  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  # Attributes for GSIs
  attribute {
    name = "entity_type"
    type = "S"
  }

  attribute {
    name = "GSI1_PK"
    type = "S"
  }

  attribute {
    name = "GSI1_SK"
    type = "S"
  }

  attribute {
    name = "GSI2_PK"
    type = "S"
  }

  attribute {
    name = "GSI2_SK"
    type = "S"
  }

  attribute {
    name = "GSI3_PK"
    type = "S"
  }

  attribute {
    name = "GSI3_SK"
    type = "S"
  }

  attribute {
    name = "GSI4_PK"
    type = "S"
  }

  attribute {
    name = "GSI4_SK"
    type = "S"
  }

  # GSI1: Entity lookup by type and alternate keys (username, email, name)
  global_secondary_index {
    name               = "GSI1-EntityLookup"
    hash_key           = "GSI1_PK"
    range_key          = "GSI1_SK"
    projection_type    = "ALL"
  }

  # GSI2: League and Week based queries (games by week, picks by league/week)
  global_secondary_index {
    name               = "GSI2-LeagueWeek"
    hash_key           = "GSI2_PK"
    range_key          = "GSI2_SK"
    projection_type    = "ALL"
  }

  # GSI3: User-based queries (user picks by season/week)
  global_secondary_index {
    name               = "GSI3-UserQueries"
    hash_key           = "GSI3_PK"
    range_key          = "GSI3_SK"
    projection_type    = "ALL"
  }

  # GSI4: Game-based queries (all picks for a game)
  global_secondary_index {
    name               = "GSI4-GameQueries"
    hash_key           = "GSI4_PK"
    range_key          = "GSI4_SK"
    projection_type    = "ALL"
  }

  # Enable TTL if needed
  ttl {
    attribute_name = "TTL"
    enabled        = false
  }

  tags = {
    Environment = "local"
    Name        = "FootballLeague"
  }
}