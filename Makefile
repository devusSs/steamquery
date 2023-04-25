dev:
	@rm -rf debug/
	@go mod tidy
	@go build -o debug/steamquery .
	@./debug/steamquery

pub:
	@echo "Building app for Windows (AMD64), Linux (AMD64) & MacOS (ARM64)..."
	@go mod tidy
	@GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o release/steamquery_win_amd64.exe .
	@GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o release/steamquery_lin_amd64 .
	@GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o release/steamquery_mac_arm64 .
	@echo "Done building app"

build:
	@echo "Building app for Windows (AMD64), Linux (AMD64) & MacOS (ARM64)..."
	@go mod tidy
	@GOOS=windows GOARCH=amd64 go build -o release/steamquery_win_amd64.exe .
	@GOOS=linux GOARCH=amd64 go build -o release/steamquery_lin_amd64 .
	@GOOS=darwin GOARCH=arm64 go build -o release/steamquery_mac_arm64 .
	@echo "Done building app"

clean:
	@go mod tidy
	@rm -rf debug/
	@rm -rf release/
	@rm -rf dist/