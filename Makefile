.PHONY: default
default: build

.PHONY: build
build:
	go build .

.PHONY: test
test:
	go test ./...

.PHONY: doc
doc:
	godoc -http=:6060

.PHONY: release
release:
	goreleaser release --snapshot --rm-dist
