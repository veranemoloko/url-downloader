APP_NAME = url-downloader
MAIN_FILE = cmd/server/main.go

build:
	go build -o $(APP_NAME) $(MAIN_FILE)

run:
	go run $(MAIN_FILE)

clean:
	rm -rf storage tmp $(APP_NAME)

# ------------------ tests
test:
	go test ./... -v -cover

# ------------------ checks

fmt:
	go fmt ./...

vet:
	go vet ./...

imports:
	goimports -w .

lint:
	golangci-lint run ./...

check: fmt vet imports lint 

.PHONY: build run fmt vet imports lint test check
