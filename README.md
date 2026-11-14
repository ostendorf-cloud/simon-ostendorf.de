# simon-ostendorf.de

The personal hub at [www.simon-ostendorf.de](https://www.simon-ostendorf.de): the way to
development or event technology, plus the CV.

## What it does

A Go service serves a Vue 3 single page application built on the organisation's own `govue`
framework. This milestone is a scaffold: one placeholder page, no content yet. See
`AGENTS.md`'s "Scope of this milestone" for what is planned and not yet built.

## Install

```sh
task setup
```

Installs the pinned toolchain (Go, Node, pnpm, and the lint and release tooling) through
`mise`.

## Use

```sh
task web:build   # builds the SPA into cmd/website/dist
task build        # compiles the website binary into bin/website
bin/website
```

The server reads its configuration from the environment; see the Configuration table below.
There is no secret-shaped variable yet. `/healthz`, `/readyz` and `/metrics` answer on a
second, administrative listener (`ADMIN_ADDR`), never on the public one.

## Configuration

| Variable | Purpose | Default | Required |
| --- | --- | --- | --- |
| `ADDR` | Address the public HTTP server listens on | `:8080` | no |
| `ADMIN_ADDR` | Address govue's admin listener listens on: `/healthz`, `/readyz` and `/metrics` | `:8081` | no |
| `LOG_LEVEL` | `debug`, `info`, `warn` or `error` | `info` | no |

## Container

```sh
docker compose up --build
```

Builds the multi-stage `Dockerfile` (frontend, static Go binary, then a minimal non-root
runtime stage with no shell) and serves the site at `http://localhost:8080`.
`docker-compose.yaml` is for local development only; see `charts/simon-ostendorf.de/README.md`
for a cluster installation.

## Development

```sh
task setup
task check
```

`task check` runs every lint, build and test target one after another; `task --list` shows
them individually. `task web:build` has to run before `task build` or `task test:go` can
compile `cmd/website`, since it embeds that output; `task build` and `task test:go` already
declare that dependency, so running either one through Task is enough.

Releases are driven by [release-me](https://github.com/ostendorf-cloud/release-me):
Conventional Commits decide the version, and merging the release pull request it maintains
cuts the release. `CHANGELOG.md` is generated, never hand edited.

## License

Proprietary, all rights reserved. See `LICENSE` and `NOTICE`.
