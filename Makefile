GO           ?= go
GOFMT        ?= $(GO)fmt
GOBUILD       = $(GO) build
GOCLEAN       = $(GO) clean
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
STATICCHECK  := $(FIRST_GOPATH)/bin/staticcheck
DEP          := $(FIRST_GOPATH)/bin/dep
pkgs          = ./...
BINARY_NAME   = beanstalkd_exporter

.PHONY: all
all: style staticcheck dep clean build

.PHONY: style
style:
	@echo ">> checking code style"
	! $(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: format
format:
	@echo ">> formatting code"
	$(GO) fmt $(pkgs)

.PHONY: vet
vet:
	@echo ">> vetting code"
	$(GO) vet $(pkgs)

.PHONY: staticcheck
staticcheck: $(STATICCHECK)
	@echo ">> running staticcheck"
	$(STATICCHECK) -ignore "$(STATICCHECK_IGNORE)" $(pkgs)

.PHONY: dep
dep: $(DEP)
	@echo ">> running dependency check"
	$(DEP) ensure

.PHONY: build
build:
	@echo ">> building binaries"
	$(GOBUILD) -o $(BINARY_NAME) -v

.PHONY: clean
clean:
	@echo ">> cleaning build files"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: $(STATICCHECK)
$(STATICCHECK):
	@echo ">> running static check"
	test -f $(STATICCHECK) || GOOS= GOARCH= $(GO) get -u honnef.co/go/tools/cmd/staticcheck

.PHONY: $(DEP)
$(DEP):
	@echo ">> installing dep"
	test -f $(DEP) || curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
