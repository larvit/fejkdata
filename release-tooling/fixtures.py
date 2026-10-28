import carried

COMPOSE = "x-go: &go\n  image: golang:${{GO_VERSION:-{}}}\n"
DOCKERFILE = "ARG GO_VERSION={}\nFROM golang:${{GO_VERSION}} AS latest\n"
GO_MOD = "module example.com/m\n\ngo {}\n\nrequire (\n\tgolang.org/x/text {}\n\texample.com/gone v1.0.0 // indirect\n)\n"
GORELEASER = "builds:\n  - main: ./cmd/fejkdata\n    goos: [linux, darwin, windows]\n    goarch: [amd64, arm64]\n"
MAIN = """# Changelog

## [Unreleased]

### Added

- A thing.

## [0.1.0] - 2026-10-13

The first release.
"""
NEW_GO_MOD = "module example.com/m\n\ngo 1.22\n\nrequire golang.org/x/text v0.30.0\n\nrequire example.com/new v0.1.0 // indirect\n"
NOTES = """# Go 1.28 Release Notes

## Ports {#ports}

### Darwin {#darwin}

<!-- go.dev/issue/1 -->
As [announced](go1.27#darwin) in the Go 1.27 release notes,
Go 1.28 requires macOS 14 Sonoma or later;
support for previous versions has been discontinued.

### Linux {#linux}

Go 1.28 is the last release that requires Linux kernel version 3.2 or later. Go 1.29 will require Linux kernel version 4.4 or later.

### Windows

Go 1.28 now requires at least Windows 10.1 or Windows Server 2019.

### FreeBSD

Go 1.28 requires FreeBSD 14 or later.

## Standard library

Go 1.28 requires nothing new of Darwin.
"""
PACKAGES = "    steps:\n      - uses: goreleaser/goreleaser-action@v7\n        with:\n          version: {}\n"
RELEASED = """# Changelog

## [0.1.0] - 2026-10-13

The first release.
"""
SECTION = (
	"## [0.1.1] - 2026-10-14\n\n### Changed\n\n- The binaries and image are built with Go 1.27.2.\n\n"
	"### Security\n\n- The binaries no longer carry 0.1.0's known vulnerabilities: GO-2026-0001.\n\n"
)


def carrying(go="1.27.1", goreleaser="v2.18.3", go_mod=GO_MOD.format("1.22", "v0.29.0"), go_sum=""):
	return carried.Carried(tuple(map(int, go.split("."))), goreleaser, go_mod, go_sum)
