ifneq (,$(wildcard ./.env))
	include .env
endif
export

run:
	go run -mod=mod cmd/main.go -source ./data/sourceData -target ./data/targetData --delete-missing --deep-search

run-no-delete:
	go run -mod=mod cmd/main.go -source ./data/sourceData -target ./data/targetData

test:
	go test $$(go list ./...)

go-vendor:
	@go mod vendor

go-tidy:
	go mod tidy -v

bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin

lint: bin/golangci-lint
	@if [ "$$(git config --get diff.noprefix)" = "true" ]; then printf "\n\ngolangci-lint has a bug and can't run with the current git configuration: 'diff.noprefix' is set to 'true'. To override this setting for this repository, run the following command:\n\n'git config diff.noprefix false'\n\nFor more details, see https://github.com/golangci/golangci-lint/issues/948.\n\n\n"; exit 1; fi
	bin/golangci-lint run --new --config ./golangci.yaml --new-from-rev=$$(git merge-base $$(cat .git/resource/base_sha 2>/dev/null || echo "origin/master") HEAD) --timeout=3m

lint-all: bin/golangci-lint
	bin/golangci-lint run --config ./golangci.yaml