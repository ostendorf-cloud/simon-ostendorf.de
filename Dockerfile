# syntax=docker/dockerfile:1@sha256:572cebbc2b15e6f84c077742952674889cb63cd5e9b80abdffcbe587b6f19183

# This file builds the site's two images as named final stages, "api"
# (cmd/api) and "frontend" (cmd/frontend), rather than two Dockerfiles. The
# binaries share go.mod/go.sum and internal/, so one file keeps the pinned
# base image digests below, Node included, in exactly one place. "docker build
# --target <name>" builds only the stages that final stage depends on:
# BuildKit never evaluates a stage that is not an ancestor of the requested
# target, so "--target api" never runs the Node stage below. Only the frontend
# embeds the build. Both images are released under one tag and deployed
# together. A build with no "--target" resolves to the last stage, "frontend".

# Node stage: builds the Vue SPA bundle and writes the prerendered pages into
# it, cmd/frontend/dist, the same path "task web:build" and "task prerender"
# write to locally. The Go build stage below only copies that directory out,
# never runs Node itself. Only the "frontend" stage depends on this one.
#
# The pages, the sitemap and robots.txt carry the origin of SITE_URL in
# web/src/lib/seo.ts, https://simon-ostendorf.de unless VITE_SITE_URL is set
# at build time: the image is built for one production origin.
#
# Digest for node:26.13.1-bookworm-slim is the multi-platform image index
# (linux/amd64, linux/arm64, linux/ppc64le), resolved against Docker Hub's
# API; re-resolve with
# "docker buildx imagetools inspect node:26.13.1-bookworm-slim" if this
# patch version ever moves. That command reports the index digest below,
# the same one a plain "docker pull" resolves against on either platform.
FROM node:26.13.1-bookworm-slim@sha256:0042e17d232a87a2cd485be1da7900fca4d10a4ad11b29e9285da1d2e54798c9 AS web

WORKDIR /src/web

# corepack no longer ships with Node: the Node.js TSC voted to stop bundling it
# by default starting with Node.js 25 (nodejs/corepack#722). npm still ships,
# so npm installs corepack first, then corepack installs the pinned pnpm
# version below, the exact one mise.toml's "pnpm" tool resolves to, so the
# lockfile this stage reads was written by the same pnpm a local "task setup"
# installs. Renovate updates PNPM_VERSION through the regex manager in
# renovate.json, in the same pull request as the mise.toml pin (group "pnpm");
# internal/depsync fails when the two differ.
# renovate: datasource=npm depName=pnpm
ARG PNPM_VERSION=12.23.0
RUN npm install -g corepack@0.36.0 && corepack enable && corepack prepare "pnpm@${PNPM_VERSION}" --activate

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./

# routeTable.ts reads internal/routes/routes.json straight through Vite's
# JSON import, the same file the Go side embeds, so the frontend build needs
# it on disk at the same relative path it has in the real repository tree
# (one level above web/) instead of a second copy under web/.
COPY internal/routes/routes.json /src/internal/routes/routes.json

RUN pnpm run build && pnpm run prerender

# Shared Go stage: downloads the module graph once, for both build stages
# below, so the cache layer and the base image digest exist in one place.
#
# Digest for golang:1.27.4-bookworm is the multi-platform image index
# (linux/amd64, linux/arm64, linux/386, linux/arm/v7, linux/ppc64le),
# matching mise.toml's pinned Go patch exactly; resolved against Docker
# Hub's API. Re-resolve with
# "docker buildx imagetools inspect golang:1.27.4-bookworm" if this patch
# version ever moves; that command reports the index digest below.
FROM golang:1.27.4-bookworm@sha256:ce59b41415e2524ae99926fe1f37ce48ba8c194dcbd2a56a65436cdcefa2bd34 AS go-deps

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

# Build stage for cmd/frontend: a fully static binary (CGO_ENABLED=0), so
# the runtime stage needs no libc. go:embed fails on an empty directory (see
# AGENTS.md's "A build needs the frontend first"), so the Node stage's
# output has to exist before this COPY.
FROM go-deps AS build-frontend

COPY cmd/frontend/ cmd/frontend/
COPY internal/ internal/
COPY --from=web /src/cmd/frontend/dist cmd/frontend/dist

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/frontend ./cmd/frontend

# Build stage for cmd/api: a fully static binary. cmd/api embeds nothing,
# so this stage copies only its own directory and never references the
# Node stage.
FROM go-deps AS build-api

COPY cmd/api/ cmd/api/
COPY internal/ internal/

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Runtime stage for cmd/api: distroless/static-debian12 carries only a
# minimal root filesystem, ca-certificates and tzdata, no shell, no package
# manager. The ":nonroot" tag already runs as uid/gid 65532, not root.
#
# Digest for gcr.io/distroless/static-debian12:nonroot is the multi-platform
# image index (linux/amd64, linux/arm64, linux/arm/v7, linux/s390x,
# linux/ppc64le), resolved against gcr.io's registry API; re-resolve with
# "docker buildx imagetools inspect gcr.io/distroless/static-debian12:nonroot"
# if this needs refreshing; that command reports the index digest below.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS api

COPY --from=build-api /out/api /usr/local/bin/api

# Matches ADDR's and ADMIN_ADDR's defaults (":8080" and ":8081"): the public
# listener and govue's admin listener (health, readiness, metrics).
EXPOSE 8080
EXPOSE 8081

# Numeric, not "nonroot:nonroot": the uid/gid the base image's own nonroot
# user resolves to, in the form every container runtime can resolve without
# reading /etc/passwd.
USER 65532:65532

ENTRYPOINT ["/usr/local/bin/api"]

# Runtime stage for cmd/frontend: same distroless base. ca-certificates is
# what page view counting needs when Umami is reached over https. The last
# stage, so a build with no "--target" builds it.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS frontend

COPY --from=build-frontend /out/frontend /usr/local/bin/frontend

# Matches ADDR's and ADMIN_ADDR's defaults (":8080" and ":8081").
EXPOSE 8080
EXPOSE 8081

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/frontend"]
