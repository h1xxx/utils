all:
	mkdir -p bin
	cd stts && CGO_ENABLED=0 go build -o ../bin/stts 
