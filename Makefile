build:
	@go build -o bin/gobank ./cmd/gobank
run: build
	@./bin/gobank
test: 
	@go test -v ./...
