SOURCES :=  $(shell find . -name '*.go')
BINARY_NAME := rss-banquet

.PHONY: all clean readme docker-dev test

all: $(BINARY_NAME)

dist: linux-arm64 linux-amd64 macos-amd64 macos-arm64

$(BINARY_NAME): $(SOURCES)
	go build -o dist/$(BINARY_NAME)

linux-arm64: $(SOURCES)
	GOOS=linux GOARCH=arm64 go build -trimpath -o dist/$(BINARY_NAME)-arm64

linux-amd64: $(SOURCES)
	GOOS=linux GOARCH=amd64 go build -trimpath -o dist/$(BINARY_NAME)-amd64

macos-amd64: $(SOURCES)
	GOOS=darwin GOARCH=amd64 go build -trimpath -o dist/$(BINARY_NAME)-macos-amd64

macos-arm64: $(SOURCES)
	GOOS=darwin GOARCH=arm64 go build -trimpath -o dist/$(BINARY_NAME)-macos-arm64

readme: $(BINARY_NAME)
	@PATH=./dist:${PATH} $(BINARY_NAME) readme > README.md 2>&1

clean:
	rm -f $(BINARY_NAME) dist/$(BINARY_NAME)-*

docker-dev:
	docker build -t rss-banquet-dev --target dev-server . && \
	docker run --rm -v $$PWD:/build -p 8080:8080 rss-banquet-dev

test:
	@go test -v ./... | sed '/PASS/s//\x1b[32m&\x1b[0m/' | sed '/FAIL/s//\x1b[31m&\x1b[0m/'
