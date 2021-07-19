GO := go
BUILD ?= .build
BALLS ?= .balls
TGTS ?= \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	windows/amd64

all: build balls

build: get-promu
	@echo ">> building binaries"
	@for tgt in $(TGTS); do \
		echo " > target $$tgt" ; \
		GOOS=$${tgt/\/*} GOARCH=$${tgt/*\/} promu build --prefix="$(BUILD)/$${tgt}" ; \
	done

balls: get-promu
	@echo ">> building release balls"
	@for tgt in $(TGTS); do \
		echo " > target $$tgt" ; \
		GOOS=$${tgt/\/*} GOARCH=$${tgt/*\/} promu tarball --prefix="$(BALLS)" "$(BUILD)/$${tgt}" ; \
	done

release: clean build balls get-promu get-github-release
	@echo ">> pushing binary to github"
	@promu release "$(BALLS)"

get-promu:
	$(GO) get github.com/prometheus/promu

get-github-release:
	$(GO) get github.com/aktau/github-release

clean:
	@rm -rfv .build .balls
