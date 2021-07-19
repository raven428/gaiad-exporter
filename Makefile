GO := go
BUILD ?= .build
BALLS ?= .balls
TGTS ?= \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	windows/amd64

all: build balls release

build:
	@echo ">> building binaries"
	@for tgt in $(TGTS); do \
		echo " > target $$tgt" ; \
		GOOS=$${tgt/\/*} GOARCH=$${tgt/*\/} promu build --prefix="$(BUILD)/$${tgt}" ; \
	done

balls:
	@echo ">> building release balls"
	@for tgt in $(TGTS); do \
		echo " > target $$tgt" ; \
		GOOS=$${tgt/\/*} GOARCH=$${tgt/*\/} promu tarball --prefix="$(BALLS)" "$(BUILD)/$${tgt}" ; \
	done

release: promu github-release
	@echo ">> pushing binary to github"
	@promu release "$(BALLS)"

promu:
	$(GO) get github.com/prometheus/promu

github-release:
	$(GO) get github.com/aktau/github-release

clean:
	@rm -rfv .build
