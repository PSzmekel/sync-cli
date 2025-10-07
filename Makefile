ifneq (,$(wildcard ./.env))
	include .env
endif
export

run-delete-missing-deep:
	go run -mod=mod cmd/main.go -source ./testdata/manual/source -target ./testdata/manual/target --delete-missing --deep-search

run-delete-missing-shallow:
	go run -mod=mod cmd/main.go -source ./testdata/manual/source -target ./testdata/manual/target --delete-missing

run-no-delete-deep:
	go run -mod=mod cmd/main.go -source ./testdata/manual/source -target ./testdata/manual/target --deep-search

run-no-delete-shallow:
	go run -mod=mod cmd/main.go -source ./testdata/manual/source -target ./testdata/manual/target 

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

# Manual test data setup
MANUAL_DIR ?= testdata/manual
SRC_DIR := $(MANUAL_DIR)/source
TGT_DIR := $(MANUAL_DIR)/target

.PHONY: manual-setup manual-clean manual-tree

manual-setup: manual-clean
	@echo ">> Tworzenie struktury katalogów w $(MANUAL_DIR)"
	mkdir -p "$(SRC_DIR)/sub" "$(TGT_DIR)/sub" "$(SRC_DIR)/sub/sub2" "$(TGT_DIR)/sub/sub2"

	@echo ">> Source files"
	printf "hello\n" > "$(SRC_DIR)/a.txt"
	printf "same1\n"  > "$(SRC_DIR)/same.txt"
	printf "nested\n" > "$(SRC_DIR)/sub/new.txt"
	printf "nested-level2-1\n" > "$(SRC_DIR)/sub/sub2/new.txt"

	@echo ">> Target files"
	printf "same2\n"        > "$(TGT_DIR)/same.txt"
	printf "only-target\n" > "$(TGT_DIR)/old.txt"
	printf "old-nested\n"  > "$(TGT_DIR)/sub/old.txt"
	printf "old-nested-level2-2\n"  > "$(TGT_DIR)/sub/sub2/old.txt"

	@echo ">> setting time marks (same.txt newer in source)"
	# target/same.txt older
	touch -t 202410010101 "$(TGT_DIR)/same.txt"
	# source/same.txt newer
	touch -t 202510010101 "$(SRC_DIR)/same.txt"

	@echo ">> Ready. Use 'make manual-tree' to see the structure."

manual-tree:
	@echo "Source:"
	@find "$(SRC_DIR)" -print | sed 's|^|  |'
	@echo "Target:"
	@find "$(TGT_DIR)" -print | sed 's|^|  |'

manual-clean:
	@echo ">> Delete $(MANUAL_DIR)"
	rm -rf "$(MANUAL_DIR)"