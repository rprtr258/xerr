.PHONY: test
test:
	@go run gotest.tools/gotestsum@latest

.PHONY: bump
bump:
	@go get -u ./...
	@go mod tidy
