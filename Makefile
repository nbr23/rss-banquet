SOURCES :=  $(shell find . -name '*.go')
BINARY_NAME := rss-banquet
VERSION := $(shell bash scripts/version.sh)
COMMIT := $(shell git rev-parse HEAD)
COMMIT_TIME := $(shell git show -s --format=%cI HEAD)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.commitTime=$(COMMIT_TIME)

.PHONY: all clean readme docker-dev test

all: $(BINARY_NAME)

dist: linux-arm64 linux-amd64 macos-amd64 macos-arm64

$(BINARY_NAME): $(SOURCES) scripts/version.sh
	go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)

linux-arm64: $(SOURCES) scripts/version.sh
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-arm64

linux-amd64: $(SOURCES) scripts/version.sh
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-amd64

macos-amd64: $(SOURCES) scripts/version.sh
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-macos-amd64

macos-arm64: $(SOURCES) scripts/version.sh
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-macos-arm64

readme: $(BINARY_NAME)
	@PATH="./dist:${PATH}" $(BINARY_NAME) readme > README.md 2>&1

clean:
	rm -f $(BINARY_NAME) dist/$(BINARY_NAME)-*

docker-dev:
	docker build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(COMMIT)" --build-arg COMMIT_TIME="$(COMMIT_TIME)" -t rss-banquet-dev --target dev-server . && \
	docker run --rm -v $$PWD:/build -p 8080:8080 rss-banquet-dev

test:
	@BANQUET_GLOBAL_USER_AGENT="Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:159.0) Gecko/20100101 Firefox/159.0" go test -v ./... | sed '/PASS/s//\x1b[32m&\x1b[0m/' | sed '/FAIL/s//\x1b[31m&\x1b[0m/'
