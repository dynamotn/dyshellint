BINARY_FOLDER  = $(CURDIR)/bin
INSTALL_PATH  ?= $(HOME)/.local/bin

# The version is of the format Major.Minor.Patch[-Prerelease][+BuildMetadata]
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
# A build that sits on no tag is not a release, and says so rather than
# claiming an empty version. The leading `v` is dropped to match what
# goreleaser stamps from the same tag.
VERSION    = $(if $(GIT_TAG),$(patsubst v%,%,$(GIT_TAG)),dev)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
BUILD_TIME = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS += -X gitlab.com/dynamo-tools/dyshellint/internal/version.gitCommit=${GIT_COMMIT}
LDFLAGS += -X gitlab.com/dynamo-tools/dyshellint/internal/version.gitTreeState=${GIT_DIRTY}
LDFLAGS += -X gitlab.com/dynamo-tools/dyshellint/internal/version.version=${VERSION}
LDFLAGS += -X gitlab.com/dynamo-tools/dyshellint/internal/version.buildTime=${BUILD_TIME}
LDFLAGS += $(EXT_LDFLAGS)

.PHONY: dyshellint
dyshellint:
	@mkdir -p '$(BINARY_FOLDER)'
	go build -ldflags '$(LDFLAGS)' -o '$(BINARY_FOLDER)'/dyshellint $(CURDIR)/cmd/dyshellint/main.go

.PHONY: build
build: dyshellint

.PHONY: test
test:
	go test -race ./...

# The Go side, then the linter on its own shell scripts: this repository is
# held to the guide it enforces.
.PHONY: lint
lint:
	gofmt -l .
	go vet ./...
	@bash $(CURDIR)/scripts/lint.sh

.PHONY: clean
clean:
	@rm -rf '$(BINARY_FOLDER)' '$(CURDIR)/dist'

.PHONY: install
install: build
	@install '$(BINARY_FOLDER)'/* '$(INSTALL_PATH)'

.PHONY: release
release:
	@bash $(CURDIR)/scripts/release.sh $(RELEASE_ARGS)

.PHONY: all
all: install clean

.PHONY: info
info:
	 @echo "Version:           ${VERSION}"
	 @echo "Git Tag:           ${GIT_TAG}"
	 @echo "Git Commit:        ${GIT_COMMIT}"
	 @echo "Git Tree State:    ${GIT_DIRTY}"
