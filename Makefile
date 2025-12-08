BINARY_NAME=app.exe
COVER_PROFILE=cover.out

.PHONY: build run test cover clean

build:
	go build -o $(BINARY_NAME) .

run:
	$(BINARY_NAME)

test:
	go test ./...

cover:
	go test ./... -coverprofile=$(COVER_PROFILE)
	go tool cover -html=$(COVER_PROFILE) -o cover.html

clean:
	del /f $(BINARY_NAME) 2>nul || true
	del /f $(COVER_PROFILE) 2>nul || true
	del /f cover.html 2>nul || true
