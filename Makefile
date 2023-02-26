.PHONY: build clean test test-race lint gen

.DEFAULT_GOAL := build

# Call `make V=1` in order to print commands verbosely.
ifeq ($(V),1)
    Q =
else
    Q = @
endif

BUILD_ENV := GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on
GO_PKG := $(shell go list ./...)
GO_BUILD_TAGS := 'osusergo netgo static_build' 
LDFLAGS := -extldflags -static
GO_BUILD_ARGS := -ldflags "$(LDFLAGS)" -tags "$(GO_BUILD_TAGS)"
GO_GEN_FLAGS :=
GO_TEST_FLAGS := -tags $(GO_BUILD_TAGS)
GO_VET_FLAGS :=
SHELL_ARGS :=
ifeq ($(V),1)
	GO_BUILD_ARGS += -v
	SHELL_ARGS += -v
	GO_TEST_FLAGS += -v
	GO_VET_FLAGS += -v
	GO_GEN_FLAGS += -v
endif

build: gen
	${Q}${BUILD_ENV} go build $(GO_BUILD_ARGS) -o /dev/null $(GO_PKG)

gen:
	${Q}go generate $(GO_GEN_FLAGS) ./...

test: build
	${Q}${BUILD_ENV} go test $(GO_TEST_FLAGS) $(GO_PKG)

test-race: test
	${Q}${BUILD_ENV} CGO_ENABLED=1 go test $(GO_TEST_FLAGS) -race $(GO_PKG)

lint:
	${Q}go vet $(GO_VET_FLAGS) $(GO_PKG)
	golangci-lint run

clean:

