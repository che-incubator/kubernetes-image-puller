[![Build Status](https://ci.centos.org/job/devtools-kubernetes-image-puller-build-master/badge/icon)](https://ci.centos.org/job/devtools-kubernetes-image-puller-build-master/)

## Requirements
The application is meant to be used in conjunction with [fabric-oso-proxy](https://github.com/fabric8-services/fabric8-oso-proxy) and [fabric8-auth](https://github.com/fabric8-services/fabric8-auth).

Fabric8-auth is used to obtain a service account token, which is used along with an impersonate header to create resources in other clusters/namespaces as directed by fabric8-oso-proxy.

For everything to function, it is necessary to have
1. A user in fabric8-auth for the serviceaccount used by this app
1. Users in fabric8-oso-proxy that can be impersonated to create resources on other clusters
  - These users require `create,delete,watch,get` permissions for `daemonset.apps` in their respective clusters and the specified namespace

To cache images, this app goes through oso-proxy to create daemonsets on desired clusters, which in turn create a pod on each node in the cluster consisting of a list of containers with command `sleep infinity`. This ensures that all nodes in the cluster have those images cached. We also periodically check the health of the daemonset and re-create it if necessary.

## Configuration
Configuration is done via env vars pulled from `./openshift/configmap.yaml`.
The config values to be set are

| Env Var | Usage |
| -- | -- |
| `CACHING_INTERVAL_HOURS` | Interval, in hours, between checking health of daemonsets |
| `DAEMONSET_NAME`         | Name of daemonset to be created |
| `NAMESPACE`              | Namespace where daemonset is to be created. Shared for all users |
| `IMPERSONATE_USERS`      | Comma-separated list of users to impersonate when creating daemonsets |
| `OPENSHIFT_PROXY_URL`    | URL of oso-proxy |
| `IMAGES`                 | List of images to be cached, in the format `<name>=<image>;...` |
| `OIDC_PROVIDER`          | URL of token provider for service account |
| `MULTICLUSTER`           | Run in multi cluster mode; default is true |

Additionally, `./openshift/app.yaml` has a few parameters:

| Parameter | Usage |
| -- | -- |
| `SERVICEACCOUNT_NAME`             | Name of service account used by main pod |
| `SERVICE_ACCT_CREDENTIALS_SECRET` | Name of secret storing service account details (see below) |
| `IMAGE`                           | Name of image used for main pod |
| `IMAGE_TAG`                       | Tag of image used for main pod |

Finally, a secret containing the pod's serviceaccount's secret and id should be created with the data

| Key | Value |
| --- | ----- |
| `service.account.secret` | Service account token |
| `service.account.id` | User id for service account |


## Building

### Makefile
```bash
# Build Go binary:
make build
# Make docker image:
make docker
# The above:
make
# Clean:
make clean
```
The provided Makefile has two parameters:
- `DOCKERIMAGE_NAME`: name for docker image
- `DOCKERIMAGE_TAG`: tag for docker image

### Manual
Build:
```bash
GOOS=linux go build -v -o ./bin/che-image-caching ./cmd/main.go
```
Make docker image:
```bash
docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} .
```

## Testing

```shell
make test
```

Will run unit tests.

End to end tests require [kind](https://github.com/kubernetes-sigs/kind):

```shell
GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0

./hack/run-e2e.sh
```

Will start a kind cluster and run the end-to-end tests in `./e2e`.  To remove the cluster after running the tests, pass the `--rm` argument to the script, or run `kind delete cluster --name k8s-image-puller-e2e`.