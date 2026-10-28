ARG GO_VERSION=1.27.3
ARG LOWEST_GO_VERSION=1.22.12
FROM golang:${GO_VERSION} AS portable

WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
	go vet ./... && \
	go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 -over 14 -ignore _test . && \
	go test -race ./...

FROM golang:${LOWEST_GO_VERSION} AS lowest
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
	lowest="go$(sed -n 's/^go \([0-9]*\.[0-9]*\).*/\1/p' go.mod)" && \
	{ test "$(go env GOVERSION | cut -d. -f1-2)" = "$lowest" || \
		{ echo "LOWEST_GO_VERSION is not $lowest, go.mod's go line"; exit 1; }; } && \
	go vet ./... && go test ./... && touch /tested

FROM golang:${GO_VERSION} AS vulnerabilities
WORKDIR /app
COPY . .
RUN go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -scan module -format json > /report.json

FROM scratch AS vulnerability-report
COPY --from=vulnerabilities /report.json /

# The full image ships git, which patch_release's tests drive; patch_release runs the filer.
FROM python:3.15.2 AS release-tooling
COPY release-tooling/ /release-tooling/
COPY vulnerability-filing/ /vulnerability-filing/
COPY .github/workflows/packages.yml /.github/workflows/packages.yml
COPY renovate.json /renovate.json
COPY Dockerfile /Dockerfile
RUN python3 -m unittest discover -s /release-tooling -p '*_test.py' && touch /tested

FROM python:3.15.2-slim AS vulnerability-filing
COPY vulnerability-filing/ /vulnerability-filing/
# The filer's test files an item into the plan itself.
COPY todo.md /todo.md
RUN python3 -m unittest discover -s /vulnerability-filing -p '*_test.py' && touch /tested

FROM goreleaser/goreleaser:v2.18.3 AS packaging
WORKDIR /check
COPY release-tooling/goreleaser.yaml ./
RUN git init -q && git remote add origin https://github.com/larvit/fejkdata.git && goreleaser check --config goreleaser.yaml && touch /tested

FROM portable
# BuildKit skips a stage nothing copies from.
COPY --from=lowest /tested /tmp/lowest-tested
COPY --from=release-tooling /tested /tmp/release-tooling-tested
COPY --from=vulnerability-filing /tested /tmp/vulnerability-filing-tested
COPY --from=packaging /tested /tmp/packaging-tested
RUN unformatted="$(gofmt -l .)"; test -z "$unformatted" || \
	{ echo "unformatted files:"; echo "$unformatted"; exit 1; }
