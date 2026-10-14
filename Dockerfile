ARG GO_VERSION=1.27.2
FROM golang:${GO_VERSION} AS portable

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go vet ./... && \
	go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 -over 14 -ignore _test . && \
	go test -race ./...

FROM golang:${GO_VERSION} AS vulnerabilities
WORKDIR /app
COPY . .
RUN go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 -scan module

FROM python:3.14.8-slim AS release-tooling
COPY release-tooling/ /release-tooling/
RUN python3 -m unittest discover -s /release-tooling -p '*_test.py' && touch /tested

FROM goreleaser/goreleaser:v2.18.3 AS packaging
WORKDIR /check
COPY release-tooling/goreleaser.yaml ./
RUN git init -q && git remote add origin https://github.com/larvit/fejkdata.git && goreleaser check --config goreleaser.yaml && touch /tested

FROM portable
# BuildKit skips a stage nothing copies from.
COPY --from=release-tooling /tested /tmp/release-tooling-tested
COPY --from=packaging /tested /tmp/packaging-tested
RUN unformatted="$(gofmt -l .)"; test -z "$unformatted" || \
	{ echo "unformatted files:"; echo "$unformatted"; exit 1; }
