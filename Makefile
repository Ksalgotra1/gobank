build:
	@go build -o bin/gobank ./cmd/gobank
run: build
	@./bin/gobank
test: 
	@go test -v ./...
docker-up:
	@docker-compose up -d
docker-down:
	@docker-compose down
