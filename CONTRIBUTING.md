# Contributing

Contributions to the Kubernetes Image Puller are welcome!

## Code of Conduct

This project is governed by the [Eclipse Foundation Community Code of Conduct](https://www.eclipse.org/org/documents/Community_Code_of_Conduct.php). By participating, you are expected to uphold this code. Please report unacceptable behaviour to [conduct@eclipse-foundation.org](mailto:conduct@eclipse-foundation.org).

## Eclipse Contributor Agreement (ECA)

Before your contribution can be accepted, you must sign the [Eclipse Contributor Agreement (ECA)](https://www.eclipse.org/legal/ECA.php). This is a one-time process.

1. Create an [Eclipse Foundation account](https://accounts.eclipse.org/) if you don't have one
2. Sign the ECA using the same email address as your Git commits
3. More details are in the [ECA FAQ](https://www.eclipse.org/legal/eca/faq/)

## Certificate of Origin

All commits must be signed off to indicate agreement with the [Developer Certificate of Origin (DCO)](https://developercertificate.org/). Add the following line to the end of each commit message:

```
Signed-off-by: Your Name <your.email@example.com>
```

You can do this automatically by committing with `git commit -s`.

## How to Contribute

### Reporting Issues

- Search [existing issues](https://github.com/che-incubator/kubernetes-image-puller/issues) before creating a new one
- Include steps to reproduce, expected behaviour, and actual behaviour

### Submitting Pull Requests

1. Fork the repository and create a feature branch from `main`
2. Keep PRs focused — one concern per PR
3. Link the related issue in the PR description
4. Ensure all tests pass before requesting review
5. Address review feedback and mention the reviewer when ready for re-review

### Prerequisites

- Go (version specified in `go.mod`)
- Docker or Podman
- `make`
- [kind](https://github.com/kubernetes-sigs/kind) (for end-to-end tests)

### Building

```bash
# Run tests and build both binaries (kubernetes-image-puller + sleep)
make build

# Build the production container image (multi-stage, UBI8-based)
make docker

# Build the dev container image (requires local build first, distroless-based)
make docker-dev
```

### Running Tests

```bash
# Run unit tests
make test

# Run linter (requires golangci-lint)
make lint

# Run vulnerability scanner (requires govulncheck)
govulncheck ./...
```

End-to-end tests require a running Kubernetes cluster. Use the provided script to set up a [kind](https://github.com/kubernetes-sigs/kind) cluster and run e2e tests:

```bash
./hack/run-e2e.sh
```

To remove the cluster afterward, pass `--rm` or run `kind delete cluster --name k8s-image-puller-e2e`.

### Dependency Management

This project uses Go module vendoring. After modifying dependencies:

```bash
go mod tidy
go mod vendor
```

Both `go.sum` and the `vendor/` directory must be committed.

## Contact

For questions, open a [discussion](https://github.com/che-incubator/kubernetes-image-puller/issues) or reach out on the [Eclipse Che mailing list](https://accounts.eclipse.org/mailing-list/che-dev).
