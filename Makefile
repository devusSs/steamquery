dev:
	@go mod tidy
	@go build -o debug/steamquery_dev/ ./...
	@./debug/steamquery_dev/steamquery

build:
	@echo "Building app for Windows (AMD64), Linux (AMD64) & MacOS (ARM64)..."
	@go mod tidy
	@GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o release/steamquery_win_amd64/ ./...
	@GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o release/steamquery_lin_amd64/ ./...
	@GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o release/steamquery_mac_arm64/ ./...
	@echo "Done building app"

clean:
	@go mod tidy
	@rm -rf debug/
	@rm -rf release/
	@rm -rf dist/
	@rm -rf logs/