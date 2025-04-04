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
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2 run

.PHONY: lint-fix
lint-fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2 run --fix

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
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2 run --disable godox
