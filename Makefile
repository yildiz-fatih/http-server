test:
	go test ./...

test-v:
	go test -v ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	rm coverage.out

cover-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out
