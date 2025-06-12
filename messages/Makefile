test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

test-coverage-html:
	go tool cover -html=coverage.out -o coverage.html

test-coverage: test
	go tool cover -func=coverage.out
