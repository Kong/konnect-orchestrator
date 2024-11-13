.PHONY: test-all
test-all: lint test

.PHONY: lint
lint:
	golangci-lint run -v ./...

.PHONY: clean
clean:
	rm -f ./koctl

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -o ./koctl ./cmd/koctl

# Kept typing this wrong
buld: build

.PHONY: coverage
coverage:
	go test -race -v -count=1 -coverprofile=coverage.out.tmp ./...
	# ignoring generated code for coverage
	grep -E -v 'generated.deepcopy.go' coverage.out.tmp > coverage.out
	rm -f coverage.out.tmp

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: test-integration
test-integration:
	go test -v -count=1 -tags=integration \
		-race \
		${GOTESTFLAGS} \
		./test/integration/...
