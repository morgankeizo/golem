.PHONY: build archive clean lint

name := golem

build:
	go build .

archive:
	find . -name '*.go' | cpio -pdm ${name}-archive
	cp go.mod ${name}-archive
	tar -czvf ${name}.tar.gz ${name}-archive
	rm -r ${name}-archive

clean:
	rm ${name} ${name}.tar.gz || true

lint:
	gofmt -s -w .
