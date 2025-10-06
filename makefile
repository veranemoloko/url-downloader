APP_NAME = url-downloader
MAIN_FILE = cmd/server/main.go

build:
	go build -o $(APP_NAME) $(MAIN_FILE)

run:
	go run $(MAIN_FILE)

run_jq:
	go run $(MAIN_FILE) | jq

clean:
	rm -rf downloads $(APP_NAME)

test:
	go test ./... -v -cover

test_cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

check:
	go fmt ./...
	go vet ./... || exit 1
	goimports -w .
	golangci-lint run ./... || exit 1
	govulncheck ./... || exit 1
	go test -race ./... || exit 1


.PHONY: build run test test_cover check clean
