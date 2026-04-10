.DEFAULT_GOAL := test

.PHONY: test
test:
	go test -race -count=2 ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf *.out *.test
