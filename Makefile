all:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/fileren fileren/fileren.go
	cd stts && CGO_ENABLED=0 go build -o ../bin/stts 
	cd ipinfolookup && CGO_ENABLED=0 go build -o ../bin/ipinfolookup
	cd mailcount && CGO_ENABLED=0 go build -o ../bin/mailcount
	cd airvpn && CGO_ENABLED=0 go build -o ../bin/airvpn
	cd string_normalize && CGO_ENABLED=0 go build -o ../bin/string_normalize

format:
	~/go/bin/goimports -w */*.go */*/*.go
	gofmt -w */*.go */*/*.go

