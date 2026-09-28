.PHONY: test build install-host

test:
	cd daemon && go test ./...

build:
	cd daemon && mkdir -p bin && go build -o bin/cloakwork ./cmd/cloakwork

install-host: build
	./install/install-host-macos.sh
