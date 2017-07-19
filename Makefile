PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
LDFLAGS := $(shell go run buildscripts/gen-ldflags.go)

BUILD_LDFLAGS := '$(LDFLAGS)'

all: build

checks:
	@echo "checks --- Check deps"
	@(env bash $(PWD)/buildscripts/checkdeps.sh)
	@echo "Checking project is in GOPATH"
	@(env bash $(PWD)/buildscripts/checkgopath.sh)
	@echo "checks ended"

getdeps: checks
	@echo "Installing golint" && go get -u github.com/golang/lint/golint
	@echo "Installing gocyclo" && go get -u github.com/fzipp/gocyclo
	@echo "Installing deadcode" && go get -u github.com/remyoudompheng/go-misc/deadcode
	@echo "Installing misspell" && go get -u github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign" && go get -u github.com/gordonklaus/ineffassign

verifiers: getdeps vet fmt lint cyclo spelling deadcode

vet:
	@echo "Running $@ suspicious constructs"
	@go tool vet -atomic -bool -copylocks -nilfunc -printf -shadow -rangeloops -unreachable -unsafeptr -unusedresult main.go template.go utils.go
	@go tool vet -atomic -bool -copylocks -nilfunc -printf -shadow -rangeloops -unreachable -unsafeptr -unusedresult pkg

fmt:
	@echo "Running $@ indentation and blanks for alignment"
	@gofmt -d main.go template.go utils.go
	@gofmt -d pkg

lint:
	@echo "Running $@ style mistakes"
	@${GOPATH}/bin/golint -set_exit_status github.com/patrickalin/bloomsky-client-go/pkg...

ineffassign:
	@echo "Running $@"
	@${GOPATH}/bin/ineffassign .

cyclo:
	@echo "Running $@"
	@${GOPATH}/bin/gocyclo -over 100 main.go
	@${GOPATH}/bin/gocyclo -over 100 template.go
	@${GOPATH}/bin/gocyclo -over 100 utils.go
	@${GOPATH}/bin/gocyclo -over 100 pkg

deadcode:
	@${GOPATH}/bin/deadcode

spelling:
	@${GOPATH}/bin/misspell -error main.go
	@${GOPATH}/bin/misspell -error `find pkg/`

# Builds, runs the verifiers then runs the tests.
check: test
test: verifiers build
	@echo "Running all testing"
	@go test $(GOFLAGS) .
	@go test $(GOFLAGS) github.com/patrickalin/bloomsky-client-go/pkg...

coverage: build
	@echo "Running all coverage"
	@./buildscripts/go-coverage.sh

# Builds locally.
build:
	@echo "Building to $(PWD)/ ..."
	@CGO_ENABLED=0 go build --ldflags $(BUILD_LDFLAGS) -o $(PWD)/bloomsky-client

# Builds and installs it to $GOPATH/bin.
install: build
	@echo "Installing at $(GOPATH)/bin/ ..."
	@cp $(PWD)/minio $(GOPATH)/bin/bloomsky-client

release: verifiers
	@MINIO_RELEASE=RELEASE ./buildscripts/build.sh

experimental: verifiers
	@MINIO_RELEASE=EXPERIMENTAL ./buildscripts/build.sh

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@rm -rf build
	@rm -rf release
	@rm -rf coverage.txt