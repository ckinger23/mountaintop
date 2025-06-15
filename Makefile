.PHONY: localstack-up localstack-down localstack-logs init-localstack test

# Docker Compose command (supports both 'docker compose' and 'docker-compose')
DOCKER_COMPOSE = $(shell command -v docker-compose 2> /dev/null || echo "docker compose")

# Start LocalStack
localstack-up:
	$(DOCKER_COMPOSE) up -d

# Stop LocalStack
localstack-down:
	$(DOCKER_COMPOSE) down

# View LocalStack logs
localstack-logs:
	$(DOCKER_COMPOSE) logs -f

# Initialize LocalStack with tables
init-localstack:
	chmod +x ./scripts/init-localstack.sh
	$(DOCKER_COMPOSE) up -d
	sleep 5  # Wait for LocalStack to start

	# Run the init script in a temporary container
	docker run --rm \
		--network host \
		-v $(PWD)/scripts:/scripts \
		-v $(HOME)/.aws:/root/.aws \
		-e AWS_ACCESS_KEY_ID=test \
		-e AWS_SECRET_ACCESS_KEY=test \
		-e AWS_DEFAULT_REGION=us-east-1 \
		amazon/aws-cli \
		--endpoint-url=http://localhost:4566 \
		./scripts/init-localstack.sh

# Run tests
test:
	go test ./... -v
