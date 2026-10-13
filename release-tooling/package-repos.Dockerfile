# The apt, dnf and apk repositories GitHub Pages serves, signed, from the .deb, .rpm and
# .apk files of the `packages` stage: a snapshot built from the source here, or a release's
# when `--build-context packages=<dir>` replaces it. The checks target installs fejkdata
# from each repository, and where EXPECTED_VERSION is set, checks it prints that tag; the
# README's Development section shows the commands.
ARG BASE_URL=https://larvit.github.io/fejkdata
ARG EXPECTED_VERSION

FROM goreleaser/goreleaser:v2.18.3 AS snapshot
WORKDIR /w/src
COPY . .
RUN git init -q && git add -A && git -c user.name=snapshot -c user.email=snapshot@example.invalid commit -qm snapshot \
	&& goreleaser release --snapshot --clean --skip=publish,docker,homebrew,scoop,aur \
		--config release-tooling/goreleaser.yaml

FROM scratch AS packages
COPY --from=snapshot /w/dist/*.deb /w/dist/*.rpm /w/dist/*.apk /

FROM debian:trixie-20260918-slim AS throwaway-keys-maker
RUN apt-get update && apt-get install -y --no-install-recommends gnupg openssl && rm -rf /var/lib/apt/lists/*
RUN export GNUPGHOME=/tmp/gnupg && mkdir -m 700 $GNUPGHOME \
	&& gpg --batch --quiet --passphrase '' --quick-gen-key 'fejkdata throwaway <throwaway@example.invalid>' ed25519 sign never \
	&& mkdir /keys && gpg --batch --armor --export-secret-keys > /keys/gpg.asc \
	&& openssl genrsa -out /keys/rsa.pem 2048 2>/dev/null

FROM scratch AS throwaway-keys
COPY --from=throwaway-keys-maker /keys /

FROM debian:trixie-20260918-slim AS deb-rpm
ARG BASE_URL
RUN apt-get update && apt-get install -y --no-install-recommends apt-utils createrepo-c dpkg-dev gnupg \
	&& rm -rf /var/lib/apt/lists/*
COPY --from=packages *.deb /site/deb/pool/
COPY --from=packages *.rpm /site/rpm/
WORKDIR /site
RUN --mount=type=secret,id=gpg,required=true --mount=type=tmpfs,target=/root/.gnupg \
	gpg --batch --quiet --import /run/secrets/gpg \
	&& gpg --armor --export > fejkdata.asc \
	&& cd deb \
	&& for arch in amd64 arm64; do \
		mkdir -p dists/stable/main/binary-$arch \
		&& dpkg-scanpackages --arch $arch pool > dists/stable/main/binary-$arch/Packages \
		&& gzip -9k dists/stable/main/binary-$arch/Packages; \
	done \
	&& apt-ftparchive -o APT::FTPArchive::Release::Origin=fejkdata \
		-o APT::FTPArchive::Release::Suite=stable -o APT::FTPArchive::Release::Codename=stable \
		-o APT::FTPArchive::Release::Components=main \
		-o "APT::FTPArchive::Release::Architectures=amd64 arm64" \
		release dists/stable > dists/stable/Release \
	&& gpg --batch --yes --clearsign -o dists/stable/InRelease dists/stable/Release \
	&& gpg --batch --yes --armor --detach-sign -o dists/stable/Release.gpg dists/stable/Release \
	&& cd ../rpm \
	&& createrepo_c --quiet . \
	&& gpg --batch --yes --armor --detach-sign repodata/repomd.xml \
	&& printf '[fejkdata]\nname=fejkdata\nbaseurl=%s/rpm\nrepo_gpgcheck=1\ngpgcheck=0\ngpgkey=%s/fejkdata.asc\n' \
		"$BASE_URL" "$BASE_URL" > fejkdata.repo

FROM alpine:3.22.2 AS apk
RUN apk add --no-cache abuild openssl
COPY --from=packages *.apk /packages/
RUN --mount=type=secret,id=rsa,required=true,target=/keys/fejkdata.rsa \
	mkdir -p /site/apk/x86_64 /site/apk/aarch64 \
	&& openssl rsa -in /keys/fejkdata.rsa -pubout -out /site/apk/fejkdata.rsa.pub 2>/dev/null \
	&& cp /packages/*_linux_amd64.apk /site/apk/x86_64/ \
	&& cp /packages/*_linux_arm64.apk /site/apk/aarch64/ \
	&& for arch in x86_64 aarch64; do \
		cd /site/apk/$arch \
		&& apk index --quiet --allow-untrusted -o APKINDEX.tar.gz *.apk \
		&& mv *.apk "fejkdata-$(tar -xzOf APKINDEX.tar.gz APKINDEX | sed -n 's/^V://p').apk" \
		&& abuild-sign -q -k /keys/fejkdata.rsa APKINDEX.tar.gz; \
	done

FROM scratch AS site
COPY --from=deb-rpm /site /
COPY --from=apk /site /

FROM debian:trixie-20260918-slim AS check-deb
ARG EXPECTED_VERSION
COPY --from=site / /site
RUN echo 'deb [signed-by=/site/fejkdata.asc] file:/site/deb stable main' > /etc/apt/sources.list.d/fejkdata.list \
	&& apt-get update -o Dir::Etc::SourceList=/etc/apt/sources.list.d/fejkdata.list -o Dir::Etc::SourceParts=- \
	&& apt-get install -y fejkdata \
	&& fejkdata --version > /checked \
	&& { [ -z "$EXPECTED_VERSION" ] || [ "$(cat /checked)" = "fejkdata $EXPECTED_VERSION" ]; }

FROM fedora:43@sha256:a651ddf48ea28a06ed4e1e6519f51c9f47e7a5a138722ade87369b8fbb7e5b42 AS check-rpm
ARG EXPECTED_VERSION
COPY --from=site / /site
RUN sed -e 's#^baseurl=.*#baseurl=file:///site/rpm#' -e 's#^gpgkey=.*#gpgkey=file:///site/fejkdata.asc#' \
		/site/rpm/fejkdata.repo > /etc/yum.repos.d/fejkdata.repo \
	&& dnf install -y --repo fejkdata fejkdata \
	&& fejkdata --version > /checked \
	&& { [ -z "$EXPECTED_VERSION" ] || [ "$(cat /checked)" = "fejkdata $EXPECTED_VERSION" ]; }

FROM alpine:3.22.2 AS check-apk
ARG EXPECTED_VERSION
COPY --from=site / /site
RUN cp /site/apk/fejkdata.rsa.pub /etc/apk/keys/ \
	&& apk add --no-cache --repositories-file /dev/null -X /site/apk fejkdata \
	&& fejkdata --version > /checked \
	&& { [ -z "$EXPECTED_VERSION" ] || [ "$(cat /checked)" = "fejkdata $EXPECTED_VERSION" ]; }

FROM scratch AS checks
COPY --from=check-deb /checked /deb
COPY --from=check-rpm /checked /rpm
COPY --from=check-apk /checked /apk
