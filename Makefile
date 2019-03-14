GO_PKGS=$(shell go list ./... | grep -v -e "/scripts")
toadlester_BUILD_DATE_TIME=$(shell date -u "+%Y.%m.%d %H:%M:%S %Z")
toadlester_VERSION ?= UNSET
toadlester_BRANCH ?= UNSET
toadlester_COMMIT ?= UNSET

format: check-gofmt test

build: go-build

go-build:
	@echo "Building for native..."
	@CGO_ENABLED=0 go build -i -ldflags='-X "git.target.com/api-platform/toadlester/api.version=$(TOADLESTER_VERSION)" -X "git.target.com/api-platform/toadlester/api.buildDateTime=$(TOADLESTER_BUILD_DATE_TIME)" -X "git.target.com/api-platform/toadlester/api.branch=$(TOADLESTER_BRANCH)" -X "git.target.com/api-platform/toadlester/api.revision=$(TOADLESTER_COMMIT)"' -o toadlester .

check-gofmt:
	@echo "Checking formatting..."
	@FMT="0"; \
	for pkg in $(GO_PKGS); do \
		OUTPUT=`gofmt -l $(GOPATH)/src/$$pkg/*.go`; \
		if [ -n "$$OUTPUT" ]; then \
			echo "$$OUTPUT"; \
			FMT="1"; \
		fi; \
	done ; \
	if [ "$$FMT" -eq "1" ]; then \
		echo "Problem with formatting in files above."; \
		exit 1; \
	else \
		echo "Success - way to run gofmt!"; \
	fi

test:
	@go test -coverprofile=c.out $(GO_PKGS) && go tool cover -func=c.out

functional:
	@go test -v

benchmark:
	@go test -bench=.
