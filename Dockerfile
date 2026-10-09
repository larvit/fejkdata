ARG GO_VERSION=1.27.1
FROM golang:${GO_VERSION} AS portable

WORKDIR /app

COPY . .
# docs/decisions.md#the-modules-develop-in-one-committed-gowork-and-no-published-gomod-carries-a-replace
RUN pkgs="$(go list -m -f '{{.Dir}}/...')" && \
	go vet $pkgs && \
	go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 -over 14 -ignore _test . && \
	go test -race $pkgs

FROM portable
RUN unformatted="$(gofmt -l .)"; test -z "$unformatted" || \
	{ echo "unformatted files:"; echo "$unformatted"; exit 1; }
