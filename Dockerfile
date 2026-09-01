# The merge gate: `docker build .` fails if vet, formatting or tests fail, and CI
# runs exactly this. Day-to-day, prefer `docker compose run` (bind-mounts source,
# no rebuilds). GO_VERSION defaults to the latest stable Go; override it to test
# the lowest supported version: docker build --build-arg GO_VERSION=1.22.12 .
ARG GO_VERSION=1.26.4
FROM golang:${GO_VERSION}

WORKDIR /app

# Module layer cached separately from source.
COPY go.mod ./
RUN go mod download

COPY . .
RUN go vet ./... && \
	{ unformatted="$(gofmt -l .)"; test -z "$unformatted" || \
	  { echo "unformatted files:"; echo "$unformatted"; exit 1; }; } && \
	go test -race ./...
