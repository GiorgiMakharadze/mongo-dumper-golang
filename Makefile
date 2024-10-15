build:
	go build -o ./bin/mongo-dump-scheduler

run: build	
	./bin/mongo-dump-scheduler

test:
	go test -v ./...

format:
	go fmt ./...	