BINARY_NAME := atomic-banquet

.PHONY: all clean readme

all: $(BINARY_NAME)

$(BINARY_NAME):
	go build -o $(BINARY_NAME)

readme: $(BINARY_NAME)
	./$(BINARY_NAME) readme > README.md 2>&1

clean:
	rm $(BINARY_NAME)
