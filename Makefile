.PHONY: api-test api-run api-build infra-up infra-down
api-test:
	cd services/api && go test ./...
api-run:
	cd services/api && go run ./cmd/api
api-build:
	cd services/api && go build ./cmd/api
infra-up:
	docker compose up -d postgres
infra-down:
	docker compose down
