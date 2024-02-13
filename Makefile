BINARY_NAME := atomic-banquet

.PHONY: all clean readme docker-dev

all: $(BINARY_NAME)

$(BINARY_NAME):
	go build -o $(BINARY_NAME)

readme: $(BINARY_NAME)
	./$(BINARY_NAME) readme > README.md 2>&1

clean:
	rm $(BINARY_NAME)

docker-dev:
	docker build -t atomic-banquet-dev --target dev-server . && \
	docker run --rm -p 8080:8080 atomic-banquet-dev
