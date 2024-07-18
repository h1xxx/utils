all:
	go build fileren.go

test: all
	rm -rf test/tmp
	cp -a test/template test/tmp
	./fileren test/tmp
	find test/tmp -type f | sort
