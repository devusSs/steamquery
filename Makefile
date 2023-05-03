# Update the version to your needs.
BUILD_VERSION = $(STEAMQUERY_VERSION) # Set this via env.
BUILD_DATE=$$(date +%Y.%m.%d-%H:%M:%S)

# DO NOT CHANGE.
build:
	@echo "Building app for Windows (AMD64), Linux (AMD64) & MacOS (ARM64)..."
	@go mod tidy
	@GOOS=windows GOARCH=amd64 go build -v -trimpath -ldflags="-s -w -X 'main.buildVersion=${BUILD_VERSION}' -X 'main.buildDate=${BUILD_DATE}'" -o release/steamquery_win_amd64/ ./...
	@GOOS=linux GOARCH=amd64 go build -v -trimpath -ldflags="-s -w -X 'main.buildVersion=${BUILD_VERSION}' -X 'main.buildDate=${BUILD_DATE}'" -o release/steamquery_lin_amd64/ ./...
	@GOOS=darwin GOARCH=arm64 go build -v -trimpath -ldflags="-s -w -X 'main.buildVersion=${BUILD_VERSION}' -X 'main.buildDate=${BUILD_DATE}'" -o release/steamquery_mac_arm64/ ./...
	@echo "Done building app"

# DO NOT CHANGE.
clean:
	@clear
	@go mod tidy
	@rm -rf ./debug/
	@rm -rf ./release/
	@rm -rf ./dist/
	@rm -rf ./logs/
	@rm -rf ./tmp/
	@rm -rf ./testing/

# DO NOT CHANGE.
dev: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/steamquery_mac_arm64/steamquery ./testing
	@cd ./testing && ./steamquery

# DO NOT CHANGE.
test: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/steamquery_mac_arm64/steamquery ./testing
	@cd ./testing && ./steamquery -t

# DO NOT CHANGE.
beta: build
	@clear
	@-mkdir ./testing
	@-mkdir ./testing/files
	@cp -R ./files ./testing
	@cp ./release/steamquery_mac_arm64/steamquery ./testing
	@cd ./testing && ./steamquery -b