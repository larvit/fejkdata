ARG GO_VERSION=1.27.4
ARG LOWEST_GO_VERSION=1.22.12
FROM golang:${GO_VERSION} AS latest
ENV GOCACHE=/cache/build GOMODCACHE=/cache/mod
WORKDIR /app
COPY . .
RUN unformatted="$(gofmt -l .)"; test -z "$unformatted" || \
	{ echo "unformatted files:"; echo "$unformatted"; exit 1; }
RUN --mount=type=cache,target=/cache \
	go vet ./... && \
	go run github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0 -over 7 -ignore _test . && \
	go test -race ./... && touch /tested

# The coverage run counts across packages, since the root package's tests reach most of the engine. It is a
# run of its own: under the race detector, package data's tests take longer than go test's timeout of ten minutes.
FROM latest AS covered
RUN --mount=type=cache,target=/cache \
	go test -coverpkg=./... -coverprofile=/coverage.out ./...

FROM scratch AS coverage
COPY --from=covered /coverage.out /

FROM golang:${LOWEST_GO_VERSION} AS lowest
ENV GOCACHE=/cache/build GOMODCACHE=/cache/mod
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/cache \
	lowest="go$(sed -n 's/^go \([0-9]*\.[0-9]*\).*/\1/p' go.mod)" && \
	{ test "$(go env GOVERSION | cut -d. -f1-2)" = "$lowest" || \
		{ echo "LOWEST_GO_VERSION is not $lowest, go.mod's go line"; exit 1; }; } && \
	go vet ./... && go test ./... && touch /tested

FROM golang:${GO_VERSION} AS vulnerabilities
WORKDIR /app
COPY . .
RUN go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 -scan module -format json > /report.json

FROM scratch AS vulnerability-report
COPY --from=vulnerabilities /report.json /

# The full image ships git, which patch_release's tests drive.
FROM python:3.15.2 AS release-tooling
COPY release-tooling/ /release-tooling/
COPY .github/workflows/packages.yml /.github/workflows/packages.yml
COPY renovate.json /renovate.json
COPY Dockerfile /Dockerfile
COPY compose.yaml /compose.yaml
RUN python3 -m unittest discover -s /release-tooling -p '*_test.py' && touch /tested

FROM python:3.15.2-slim AS data-import
COPY data-import/ /data-import/
COPY data/ /data/
RUN python3 -m unittest discover -s /data-import -p '*_test.py' && touch /tested

FROM golang:${GO_VERSION} AS cli
ENV GOCACHE=/cache/build GOMODCACHE=/cache/mod
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/cache \
	CGO_ENABLED=0 go build -o /fejkdata ./cmd/fejkdata

FROM python:3.15.2-slim AS python-package
COPY README.md /README.md
COPY python-package/ /python-package/
COPY --from=cli /fejkdata /python-package/fejkdata/bin/fejkdata
RUN cd /python-package && python3 -m unittest discover -p '*_test.py' && touch /tested

FROM golang:${GO_VERSION} AS npm-wasm
ENV GOCACHE=/cache/build GOMODCACHE=/cache/mod
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/cache \
	GOOS=js GOARCH=wasm go vet ./npm-package/wasm && \
	mkdir /npm-package && \
	GOOS=js GOARCH=wasm go build -trimpath -ldflags='-s -w' -o /npm-package/fejkdata.wasm ./npm-package/wasm && \
	cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" /npm-package/

FROM scratch AS npm-package-built
COPY npm-package/ /npm-package/
COPY --from=npm-wasm /npm-package/ /npm-package/

FROM node:26.13.1-slim AS npm-package
COPY --from=npm-package-built /npm-package/ /npm-package/
COPY README.md /README.md
COPY --from=cli /fejkdata /fejkdata
RUN cd /npm-package && FEJKDATA_CLI=/fejkdata node --test && touch /tested

# The lowest Node that package.json's engines names.
FROM node:26.13.1-slim AS npm-package-lowest
COPY --from=npm-package-built /npm-package/ /npm-package/
COPY README.md /README.md
COPY --from=cli /fejkdata /fejkdata
RUN cd /npm-package && FEJKDATA_CLI=/fejkdata node --test && touch /tested

FROM oven/bun:1.4.5 AS npm-package-bun
COPY --from=npm-package-built /npm-package/ /npm-package/
COPY README.md /README.md
COPY --from=cli /fejkdata /fejkdata
RUN cd /npm-package && FEJKDATA_CLI=/fejkdata bun test --timeout 60000 render.test.js && touch /tested

# Deno imports the package over HTTP, as a browser does, and from disk.
FROM denoland/deno:2.9.8 AS npm-package-deno
COPY --from=npm-package-built /npm-package/ /npm-package/
RUN cd /npm-package && deno test --no-check --no-lock --allow-net --allow-read --allow-import load.deno.js && touch /tested

FROM python:3.15.2-slim AS python-lint
COPY --from=ghcr.io/astral-sh/ruff:0.17.6-alpine /usr/local/bin/ruff /usr/local/bin/ruff
RUN pip install --no-cache-dir mypy==2.4.1
COPY data-import/ /python/data-import/
COPY release-tooling/ /python/release-tooling/
COPY python-package/ /python/python-package/
RUN ruff check --no-cache /python
# mypy runs inside each folder, since a script imports its siblings by bare name, as `import geo`.
RUN for folder in data-import release-tooling python-package; do (cd /python/$folder && mypy --strict --cache-dir /dev/null .) || exit 1; done && touch /tested

FROM goreleaser/goreleaser:v2.20.3 AS packaging
WORKDIR /check
COPY release-tooling/goreleaser.yaml ./
RUN git init -q && git remote add origin https://github.com/larvit/fejkdata.git && goreleaser check --config goreleaser.yaml && touch /tested

FROM scratch AS lowest-and-tooling
# BuildKit skips a stage nothing copies from.
COPY --from=lowest /tested /tmp/lowest-tested
COPY --from=release-tooling /tested /tmp/release-tooling-tested
COPY --from=data-import /tested /tmp/data-import-tested
COPY --from=packaging /tested /tmp/packaging-tested
COPY --from=python-lint /tested /tmp/python-lint-tested
COPY --from=python-package /tested /tmp/python-package-tested
COPY --from=npm-package /tested /tmp/npm-package-tested
COPY --from=npm-package-lowest /tested /tmp/npm-package-lowest-tested
COPY --from=npm-package-bun /tested /tmp/npm-package-bun-tested
COPY --from=npm-package-deno /tested /tmp/npm-package-deno-tested

# A pull request's CI builds only latest and lowest-and-tooling, so a check goes in one of them.
FROM lowest-and-tooling
COPY --from=latest /tested /tmp/latest-tested
