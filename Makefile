all:
	mkdir -p bin
	cd stts && CGO_ENABLED=0 go build -o ../bin/stts 
	cd mailcount && CGO_ENABLED=0 go build -o ../bin/mailcount

