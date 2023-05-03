version = v0.4.0

build:
	@echo "Building app for Windows (AMD64), Linux (AMD64) & MacOS (ARM64)..."
	@go mod tidy
	@GOOS=windows GOARCH=amd64 go build -v -trimpath -ldflags="-s -w -X main.buildVersion=$(version) -X main.buildDate=$$(date +%Y.%m.%d-%H:%M:%S)" -o release/steamquery_win_amd64/ ./...
	@GOOS=linux GOARCH=amd64 go build -v -trimpath -ldflags="-s -w -X main.buildVersion=$(version) -X main.buildDate=$$(date +%Y.%m.%d-%H:%M:%S)" -o release/steamquery_lin_amd64/ ./...
	@GOOS=darwin GOARCH=arm64 go build -v -trimpath -ldflags="-s -w -X main.buildVersion=$(version) -X main.buildDate=$$(date +%Y.%m.%d-%H:%M:%S)" -o release/steamquery_mac_arm64/ ./...
	@echo "Done building app"

clean:
	@clear
	@go mod tidy
	@rm -rf ./debug/
	@rm -rf ./release/
	@rm -rf ./dist/
	@rm -rf ./logs/
	@rm -rf ./tmp/
	@rm -rf ./testing/

dev: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/steamquery_mac_arm64/steamquery ./testing
	@cd ./testing && ./steamquery

test: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/steamquery_mac_arm64/steamquery ./testing
	@cd ./testing && ./steamquery -t

beta: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/steamquery_mac_arm64/steamquery ./testing
	@cd ./testing && ./steamquery -b