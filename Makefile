GO ?= go
VERSION=`git describe --tags`
BINARY=noise

build:
	$(GO) build -o bin/${BINARY}-darwin-amd64 ./cmd/${BINARY}

build_amd64:
	env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags="-w -s" -o ./bin/${BINARY}-amd64-linux ./cmd/${BINARY}

build_emitter_windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO)  build -o ./bin/emitter-win64.exe ./cmd/emitter

build_emitter_amd64:
	env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags="-w -s" -o ./bin/emitter-amd64-linux ./cmd/emitter

run_emitter:
	$(GO) build -o bin/emitter-darwin-amd64 ./cmd/emitter
	bin/emitter-darwin-amd64