all:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/stts stts/main.go
