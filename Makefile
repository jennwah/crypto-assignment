# lint
lint:
	golangci-lint run

# generate swagger API docs
swag:
	~/go/bin/swag init -g ./cmd/main.go

# database migrations
create_migration:
	migrate create -dir ./migrations -ext sql $(NAME)
migrate_up:
	migrate -path ./migrations -database postgres://root:pass@localhost:5432/database?sslmode=disable up $(N)
migrate_down:
	migrate -path ./migrations -database postgres://root:pass@localhost:5432/database?sslmode=disable down $(N)
migrate_version:
	migrate -path ./migrations -database postgres://root:pass@localhost:5432/database?sslmode=disable version
seed_data:
	@echo "Seeding the database..."
	@for file in seed/*.sql; do \
		echo "Running seed file: $$file"; \
		docker exec database bash -c "PGPASSWORD=pass psql -U root -d database -f /seed/`basename $$file`"; \
	done

# spin up services
run:
	docker compose up -d
# build
build:
	go build -o dist/api cmd/main.go
stop:
	docker compose down

# tests
unit_test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
unit_test_vis:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out