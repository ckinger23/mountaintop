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
    name = "id"
    type = "S"
  }

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "week"
    type = "S"
  }

  attribute {
    name = "game_id"
    type = "S"
  }

  attribute {
    name = "pick"
    type = "S"
  }

  # GSI-EntityType: entity_type as the partition key, id as the sort key
  global_secondary_index {
    name               = "GSI-EntityType"
    hash_key           = "entity_type"
    range_key          = "id"
    projection_type    = "ALL"
    read_capacity     = 5
    write_capacity    = 5
  }

  # GSI-UserPicks: user_id as the partition key, week as the sort key
  global_secondary_index {
    name               = "GSI-UserPicks"
    hash_key           = "user_id"
    range_key          = "week"
    projection_type    = "ALL"
    read_capacity     = 5
    write_capacity    = 5
  }

  # GSI-GamePicks: game_id as the partition key, pick as the sort key
  global_secondary_index {
    name               = "GSI-GamePicks"
    hash_key           = "game_id"
    range_key          = "pick"
    projection_type    = "ALL"
    read_capacity     = 5
    write_capacity    = 5
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