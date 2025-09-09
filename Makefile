dependencies-up:
	cd main-service && \
	docker compose -f docker-compose.local.osx.yaml up -d

run:
	cd main-service && \
	PORT=8080 \
	POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/app?sslmode=disable \
	REDIS_ADDR=localhost:6379 \
	REDIS_PASSWORD= \
	REDIS_DB=0 \
	KAFKA_BROKERS=localhost:29092 \
	KAFKA_CLIENT_ID=main-service \
	go run main.go

test:
	cd main-service && go test -v ./...

cover:
	cd main-service && \
	go test -coverprofile=coverage.out ./... -tags=test && \
	go tool cover -html=coverage.out

gomock:
	cd main-service && \
	rm -rf mocks && \
	./scripts/gen_mocks.sh
