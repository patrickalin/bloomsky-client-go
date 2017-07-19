PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
LDFLAGS := $(shell go run scripts/build/gen-ldflags.go)

BUILD_LDFLAGS := '$(LDFLAGS)'

all: build

checks:
	@echo "checks --- Check deps"
	@(env bash $(PWD)/scripts/checkdeps.sh)
	@echo "Checking project is in GOPATH"
	@(env bash $(PWD)/scripts/checkgopath.sh)
	@echo "checks ended"

getdeps: checks
	@echo "Installing golint" && go get -u github.com/golang/lint/golint
	@echo "Installing gocyclo" && go get -u github.com/fzipp/gocyclo
	@echo "Installing deadcode" && go get -u github.com/remyoudompheng/go-misc/deadcode
	@echo "Installing misspell" && go get -u github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign" && go get -u github.com/gordonklaus/ineffassign
	@echo "Installing errcheck" && go get -u github.com/kisielk/errcheck

verifiers: getdeps vet fmt lint cyclo spelling deadcode errcheck

vet:
	@echo "Running $@ suspicious constructs"
	@go tool vet -atomic -bool -copylocks -nilfunc -printf -shadow -rangeloops -unreachable -unsafeptr -unusedresult main.go template.go utils.go
	@go tool vet -atomic -bool -copylocks -nilfunc -printf -shadow -rangeloops -unreachable -unsafeptr -unusedresult pkg

fmt:
	@echo "Running $@ indentation and blanks for alignment"
	@gofmt -d *.go
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

errcheck:
	@echo "Running $@"
	@${GOPATH}/bin/errcheck github.com/patrickalin/bloomsky-client-go

# Builds, runs the verifiers then runs the tests.
check: test
test: verifiers build
	@echo "Running all testing"
	@go test $(GOFLAGS) .
	@go test $(GOFLAGS) github.com/patrickalin/bloomsky-client-go/pkg...

coverage: build
	@echo "Running all coverage"
	@./scripts/go-coverage.sh

# Builds locally.
build:
	@echo "Building to $(PWD)/ ..."
	@go list -f '{{ .Name }}: {{ .Doc }}'
	@CGO_ENABLED=0 go build --ldflags $(BUILD_LDFLAGS) -o $(PWD)/bloomsky-client

# Builds and installs it to $GOPATH/bin.
install: build
	@echo "Installing at $(GOPATH)/bin/ ..."
	@cp $(PWD)/minio $(GOPATH)/bin/bloomsky-client

release: verifiers
	@MINIO_RELEASE=RELEASE ./scripts/build.sh

experimental: verifiers
	@BLOOMSKY_RELEASE=EXPERIMENTAL ./scripts/build.sh

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@rm -rf build
	@rm -rf release
	@rm -rf coverage.txt

doc:
	@echo "listen on http://localhost:8081 ctrl+c stop"
	@(env bash $(PWD)/scripts/doc.sh)

bench:
	@echo "Running $@ generate prof.cpu"
	@go test -bench . -cpuprofile prof.cpu

travisGihtub:
    @travis encrypt GITHUB_SECRET_TOKEN=$(GITHUB_SECRET_TOKEN) -a


