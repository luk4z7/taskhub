run:
	go run main.go

generate: generate-install generate-run

generate-install:
ifneq (, $(shell which mockgen))
	mockgen -version
else
	go install go.uber.org/mock/mockgen@latest
endif

generate-run:
	go generate ./...

test:
	go test -v -race ./...
