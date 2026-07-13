TXTMAN=txt2man
GO_BIN ?= "go"

man: docs/uchess.txt
	$(TXTMAN) -s1 -p -P uchess -t uchess docs/uchess.txt > docs/uchess.man

install: tidy
	cd cmd/uchess && $(GO_BIN) install
	make tidy

tidy:
	$(GO_BIN) mod tidy -v

test:
	$(GO_BIN) test ./...

build: tidy
	cd cmd/uchess && $(GO_BIN) build -v .
	make tidy

clean:
	rm -f cmd/uchess/uchess
	rm -rf dist

release:
	goreleaser release --clean

release_test:
	goreleaser release --snapshot --skip=publish --clean
