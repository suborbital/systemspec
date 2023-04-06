.PHONY: lint lintfix test

lint:
	golangci-lint run ./...

lintfix:
	golangci-lint run ./... --fix

test:
	 go test -v -count=1 ./...
