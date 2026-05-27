plextraccli: *.go */*.go
	go build -v

.PHONY: test
test: plextraccli
	plextraccli --client demo reports
	plextraccli --client demo reports --cols status,name
	plextraccli --client demo reports --cols +tags
	plextraccli --client demo reports --cols err

.PHONY: lint
lint: plextraccli
	go tool golangci-lint run

.PHONY: lint-fix
lint-fix: plextraccli
	go tool golangci-lint run --fix

.PHONY: watch
watch:
	while true; do \
		m="$$(find . -iname '*.go' | sort | while read -r f; do cat "$$f"; done | md5sum)"; \
		if [ "$$m" != "$$n" ]; then \
			make test; \
		fi; \
		n="$$m"; \
	done

.PHONY: pre-commit
pre-commit: plextraccli
	go test -v ./...
	go tool golangci-lint run --disable godox

.PHONY: release
release:
	~/go/bin/goreleaser release --snapshot --clean
