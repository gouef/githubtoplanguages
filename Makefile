.PHONY: install tests coverage build

install:
	go mod tidy && go mod vendor

tests:
	go test -covermode=set ./... -coverprofile=coverage.txt && go tool cover -func=coverage.txt
coverage:
	go test -v -covermode=set ./... -coverprofile=coverage.txt && go tool cover -html=coverage.txt -o coverage.html && xdg-open coverage.html
build:
	go build .