# Agent Guidelines

This file provides context for AI coding agents working on the Kubernetes Image Puller. It complements `README.md` and `dev-guide.adoc` with non-obvious project knowledge that is difficult to derive from code alone.

## Quick Reference

```bash
make test       # Run unit tests (cfg, pkg, sleep, utils)
make build      # Run tests, then build binaries (kubernetes-image-puller + sleep)
make lint       # Run golangci-lint (must be installed separately)
make docker     # Build production container image (multi-stage, UBI8-based)
make docker-dev # Build dev container image (requires local build first, distroless-based)
make clean      # Remove build artifacts
```

### Additional Checks

```bash
go vet ./...              # Static analysis
govulncheck ./...         # Vulnerability scanner (install: go install golang.org/x/vuln/cmd/govulncheck@latest)
golangci-lint run -v      # Extended linting (install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest)
```

## Architecture

The Kubernetes Image Puller caches container images on cluster nodes by creating a DaemonSet where each pod runs `sleep 720h` containers for the configured images. An initContainer copies a Go-based `sleep` binary into a shared volume so that even scratch-based images (which lack a shell) can be cached.

The application runs as a Deployment. On startup it creates a DaemonSet, waits for it to become ready, then periodically recreates it to refresh images with mutable tags.

### How It Works

1. `cmd/main.go` — Entrypoint, calls `singlecluster.CacheImages()`
2. `pkg/single-cluster/puller.go` — Sets up the Kubernetes client (KUBECONFIG, ~/.kube/config, or in-cluster), creates the DaemonSet, handles SIGTERM for clean shutdown, and periodically refreshes the cache
3. `utils/operations.go` — High-level operations (CacheImages, RefreshCache, DeleteDaemonsetIfExists) called by the puller
4. `utils/clusterutils.go` — Low-level Kubernetes API interactions: DaemonSet construction, watch-based readiness checks, container spec generation
5. `cfg/` — Configuration from environment variables with defaults

### Key Directories

| Path | Purpose |
|------|---------|
| `cmd/` | Application entrypoint |
| `cfg/` | Environment variable configuration with defaults |
| `pkg/single-cluster/` | Core puller logic (client setup, lifecycle, signal handling) |
| `utils/` | Kubernetes API operations and DaemonSet construction |
| `sleep/` | Standalone Go sleep binary (replaces coreutils sleep for scratch images) |
| `e2e/` | End-to-end tests (require a running Kubernetes cluster) |
| `deploy/helm/` | Helm chart templates |
| `deploy/openshift/` | OpenShift templates |
| `build/dockerfiles/` | Production (`Dockerfile`) and dev (`dev.Dockerfile`) container builds |
| `vendor/` | Vendored Go dependencies (committed to repo) |

## Critical Gotchas

### Dependencies Use Vendoring

This project vendors all Go dependencies. After any `go get` or `go mod tidy`, you **must** run `go mod vendor` and commit the resulting `vendor/` changes. CI builds from the vendor directory, not the module cache.

### Two Dockerfiles, Two Base Images

- **`build/dockerfiles/Dockerfile`** (production): Multi-stage build on UBI8. Creates `appuser` with UID 65532 to match the PodSecurityContext. Used by CI for the `quay.io/eclipse/kubernetes-image-puller` image.
- **`build/dockerfiles/dev.Dockerfile`** (dev): Single-stage, copies pre-built binaries into `gcr.io/distroless/static-debian12:nonroot` (which also uses UID 65532). Requires running `make build` first.

If you change the UID in PodSecurityContext (`utils/clusterutils.go`), you must also update the `adduser` command in `Dockerfile`. The dev.Dockerfile inherits UID 65532 from the `:nonroot` distroless tag.

### Configuration Is Environment-Variable Driven

All runtime configuration comes from environment variables (see `cfg/envvars.go`). There are no config files or flags. The `cfg.GetConfig()` function is called from multiple packages — it reads env vars fresh each time (no caching). Tests that call functions using config **must** set the required env vars (at minimum `IMAGES` and `CACHING_INTERVAL_HOURS`).

### Test Patterns

- **Unit tests** (`cfg/`, `utils/`, `sleep/`): Run locally with `make test`. Tests use `os.Setenv`/`os.Clearenv` to configure the environment. When adding tests, always `defer os.Clearenv()` to avoid leaking state between tests.
- **E2e tests** (`e2e/`): Require a running cluster with `NAMESPACE`, `KUBECONFIG`, and `DAEMONSET_NAME` env vars set. These are **not** run by `make test` — they must be invoked manually against a cluster.
- The `utils/` package tests only cover functions that don't require a Kubernetes client. Functions that need a client are covered by e2e tests.

### RBAC Templates Must Stay In Sync

RBAC rules are defined in two places that must match:
- `deploy/helm/templates/serviceaccount.yaml` (Helm)
- `deploy/openshift/serviceaccount.yaml` (OpenShift)

When changing RBAC permissions, update both files identically.

### The Sleep Binary

The `sleep/` directory contains a standalone Go implementation of `sleep` that accepts Go duration strings (e.g. `720h`). It exists because scratch-based container images lack coreutils. The binary is built separately (`./bin/sleep`) and copied into containers. It is not a library — it has its own `package main`.

### CGO and Static Linking

The Makefile default is `CGO_ENABLED=1`. This is required for FIPS compliance on Red Hat UBI images, which link against the FIPS-certified OpenSSL library instead of Go's native crypto. The dev Dockerfile uses a distroless image and may need `CGO_ENABLED=0` for static binaries — see `build/dockerfiles/dev.Dockerfile`.

### Error Handling Style

The codebase uses `log.Fatalf()` for unrecoverable errors (failed client creation, missing env vars, failed API calls). This is intentional — the application runs as a Kubernetes Deployment, so a crash triggers a restart. Do not convert `log.Fatalf` calls to error returns without considering the restart-based recovery model.

## CI Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `next-build.yml` | Push to `main`, pull requests, manual | Runs test job (unit tests + govulncheck), then builds and pushes container image to quay.io |

The test job uses `go-version-file: go.mod` to install the Go version specified in the module. When bumping the Go version in `go.mod`, CI automatically picks up the new version.

### CI Security Practices

- All GitHub Actions should be pinned to SHA digests (not mutable version tags)
- Checkout steps should use `persist-credentials: false`
- Workflow permissions should be explicitly scoped (e.g. `contents: read`)
- govulncheck should be pinned to a specific version, not `@latest`

## Contribution Workflow

- Fork-based contributions — do not push directly to the upstream repo
- Eclipse Public License 2.0 (EPL-2.0)
- Squash commits into clean, atomic units before requesting review
- One concern per PR, link related issues
- Run `make test`, `go vet ./...`, and `govulncheck ./...` before submitting
- The project has no CONTRIBUTING.md — follow patterns from existing PRs

## Related Projects

- [kubernetes-image-puller-operator](https://github.com/che-incubator/kubernetes-image-puller-operator) — Operator that manages the image puller via a `KubernetesImagePuller` custom resource. Changes to config fields or RBAC in this repo may require corresponding operator updates.
