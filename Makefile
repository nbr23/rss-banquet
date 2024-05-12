BINARY_NAME := rss-banquet

.PHONY: all clean readme docker-dev test

all: $(BINARY_NAME)

$(BINARY_NAME):
	go build -o $(BINARY_NAME)

readme: $(BINARY_NAME)
	./$(BINARY_NAME) readme > README.md 2>&1

clean:
	rm $(BINARY_NAME)

docker-dev:
	docker build -t rss-banquet-dev --target dev-server . && \
	docker run --rm -v $$PWD:/build -p 8080:8080 rss-banquet-dev

test:
	@go test -v ./... | sed '/PASS/s//\x1b[32m&\x1b[0m/' | sed '/FAIL/s//\x1b[31m&\x1b[0m/'
