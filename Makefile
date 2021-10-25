.PHONY: build archive clean lint

name := golem

build:
	go build .

archive:
	mkdir ${name}-archive
	cp go.mod ${name}-archive
	cp main.go ${name}-archive
	cp -r protocol ${name}-archive
	cp -r proxy ${name}-archive
	cp -r server ${name}-archive
	tar -czvf ${name}.tar.gz ${name}-archive
	rm -r ${name}-archive

clean:
	rm ${name}

lint:
	gofmt -s -w .
