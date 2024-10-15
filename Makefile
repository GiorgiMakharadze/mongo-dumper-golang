BINARY=./bin/mongo-dump-scheduler

build:
	mkdir -p bin
	go build -o $(BINARY)

run: build
	$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test -v ./...

format:
	go fmt ./...	


.PHONY: build run clean test format
