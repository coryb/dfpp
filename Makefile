PLATFORMS= \
	freebsd-amd64 \
	linux-amd64 \
	windows-amd64 \
	darwin-amd64 \
	$(NULL)

NAME     = dfpp
DIST     = $(shell pwd)/dist
GOPATH   = $(shell pwd)
GOBIN   ?= $(shell pwd)
BIN     ?= $(GOBIN)/$(NAME)
CURVER  ?= $(patsubst v%,%,$(shell git describe --abbrev=0 --tags))
NEWVER  ?= $(shell echo $(CURVER) | awk -F. '{print $$1"."$$2"."$$3+1}')
TODAY   := $(shell date +%Y-%m-%d)

export GOPATH

build: src/gopkg.in/coryb/dfpp.v1
	go build -ldflags "-w -s" -o $(BIN) main/main.go

src/%:
	mkdir -p $(@D)
	test -L $@ || ln -sf ../../.. $@
	go get -v $* $*/main

install:
	${MAKE} GOBIN=~/bin build

vet:
	@go tool vet *.go main/*.go

clean:
	rm -rf pkg dist bin src ./$(NAME)

cross-setup:
	for p in $(PLATFORMS); do \
        echo "Building for $$p"; \
		cd $(GOROOT)/src && sudo GOROOT_BOOTSTRAP=$(GOROOT) GOOS=$${p/-*/} GOARCH=$${p/*-/} bash ./make.bash --no-clean; \
   done

all:
	rm -rf $(DIST); \
	mkdir -p $(DIST); \
	for p in $(PLATFORMS); do \
        echo "Building for $$p"; \
        ${MAKE} build GOOS=$${p/-*/} GOARCH=$${p/*-/} BIN=$(DIST)/$(NAME)-$$p; \
    done

fmt:
	gofmt -s -w *.go main/*.go

changes:
	@git log --pretty=format:"* %s [%cn] [%h]" --no-merges ^v$(CURVER) HEAD main/*.go *.go | grep -vE 'gofmt|go fmt'

update-changelog: 
	@echo "# Changelog" > CHANGELOG.md.new; \
	echo >> CHANGELOG.md.new; \
	echo "## $(NEWVER) - $(TODAY)" >> CHANGELOG.md.new; \
	echo >> CHANGELOG.md.new; \
	$(MAKE) --no-print-directory --silent changes | \
	perl -pe 's{\[([a-f0-9]+)\]}{[[$$1](https://github.com/coryb/dfpp/commit/$$1)]}g' | \
	perl -pe 's{\#(\d+)}{[#$$1](https://github.com/coryb/dfpp/issues/$$1)}g' >> CHANGELOG.md.new; \
	tail -n +2 CHANGELOG.md >> CHANGELOG.md.new; \
	mv CHANGELOG.md.new CHANGELOG.md; \
	git commit -m "Updated Changelog" CHANGELOG.md; \
	git tag v$(NEWVER)

version:
	@echo $(CURVER)

docker:
	mkdir -p docker-root/bin docker-root/etc/ssl/certs
	/usr/bin/security find-certificate -a -p /System/Library/Keychains/SystemRootCertificates.keychain > docker-root/etc/ssl/certs/ca-certificates.crt
	${MAKE} BIN=./docker-root/bin/$(NAME) GOOS=linux GOARCH=amd64 build
	docker build -t coryb/$(NAME):$(CURVER) .
	docker tag coryb/$(NAME):$(CURVER) coryb/$(NAME):latest

release: docker
	docker push coryb/$(NAME):$(CURVER)
	docker push coryb/$(NAME):latest
