.PHONY: all
all:
	go build -v -mod vendor -o tool src/main.go

.PHONY: clean
clean:
	go clean -x -cache
