all: footloose

footloose: bin/footloose

bin/footloose:
	CGO_ENABLED=0 go build -v -o bin/footloose .

install: bin/footloose
	install -D $^ $(shell go env GOPATH)/bin/footloose

.PHONY: bin/footloose install binary footloose

