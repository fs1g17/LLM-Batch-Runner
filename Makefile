build:
	go build -o batch .

test:
	go test ./...

clear: 
	find ./data/out -mindepth 1 -delete
