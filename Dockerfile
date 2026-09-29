ARG GO_VERSION=1.27.1
FROM golang:${GO_VERSION} AS portable

WORKDIR /app

# Module layer cached separately from source.
COPY go.mod ./
RUN go mod download

COPY . .
RUN go vet ./... && \
	go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 -over 14 -ignore _test . && \
	go test -race ./...

FROM portable
RUN unformatted="$(gofmt -l .)"; test -z "$unformatted" || \
	{ echo "unformatted files:"; echo "$unformatted"; exit 1; }
