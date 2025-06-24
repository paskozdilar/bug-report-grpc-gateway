.PHONY: help latest previous

help:
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@echo "    latest   - Install version 2.27.0 of grpc-gateway and build"
	@echo "    previous - Install version 2.26.3 of grpc-gateway and build"
	@echo "    both     - Execute 'previous' target, then 'latest' target"
	@echo "    help     - Show this help message"

latest:
	@echo "-- Installing grpc-gateway version 2.27.0 and building..."
	go mod download
	go get -tool github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.0
	rm -rf src/proto
	go tool buf dep update
	go tool buf generate
	go build ./...

previous:
	@echo "-- Installing grpc-gateway version 2.26.3 and building..."
	go mod download
	go get -tool github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.26.3
	rm -rf src/proto
	go tool buf dep update
	go tool buf generate
	go build ./...

both: previous latest
