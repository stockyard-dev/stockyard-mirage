build:
	CGO_ENABLED=0 go build -o mirage ./cmd/mirage/

run: build
	./mirage

test:
	go test ./...

clean:
	rm -f mirage

.PHONY: build run test clean
