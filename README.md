[![next](https://github.com/che-incubator/kubernetes-image-puller/actions/workflows/next-build.yml/badge.svg)](https://github.com/che-incubator/kubernetes-image-puller/actions/workflows/next-build.yml)

[![Contribute](https://www.eclipse.org/che/contribute.svg)](https://workspaces.openshift.com#https://github.com/che-incubator/kubernetes-image-puller)

## About

To cache images, Kubernetes Image Puller creates a Daemonset on the desired cluster, which in turn creates a pod on each node in the cluster consisting of a list of containers with command `sleep 720h`.
This ensures that all nodes in the cluster have those images cached. The `sleep` binary being used is [golang-based](https://github.com/che-incubator/kubernetes-image-puller/tree/main/sleep) (please see [Scratch Images](#scratch-images)).
We also periodically check the health of the daemonset and re-create it if necessary.

The application can be deployed via Helm or by processing and applying OpenShift Templates. Also, there is a community supported operator available on the [OperatorHub](https://operatorhub.io/operator/kubernetes-imagepuller-operator).

## Configuration
Configuration is done via env vars pulled from `./deploy/helm/templates/configmap.yaml`, or `./deploy/openshift/configmap.yaml`, depending on the deployment method.
The config values to be set are:

| Env Var | Usage | Default |
| -- | -- | -- |
| `CACHING_INTERVAL_HOURS` | Interval, in hours, between checking health of daemonsets | `"1"` |
| `CACHING_MEMORY_REQUEST` | The memory request for each cached image when the puller is running | `10Mi` |
| `CACHING_MEMORY_LIMIT` | The memory limit for each cached image when the puller is running | `20Mi` |
| `CACHING_CPU_REQUEST` | The CPU request for each cached image when the puller is running | `.05` or 50 millicores |
| `CACHING_CPU_LIMIT` | The CPU limit for each cached image when the puller is running | `.2` or 200 millicores |
| `DAEMONSET_NAME`         | Name of daemonset to be created | `kubernetes-image-puller` |
| `NAMESPACE`              | Namespace where daemonset is to be created | `kubernetes-image-puller` |
| `IMAGES`                 | List of images to be cached, in the format `<name>=<image>;...` | Contains a default list of images, but should be configured when deploying |
| `NODE_SELECTOR` | Node selector applied to pods created by the daemonset       | `'{}'` |
| `IMAGE_PULL_SECRETS` | List of image pull secrets, in the format `pullsecret1;...` to add to pods created by the DaemonSet. Those secrets need to be in the image puller's namespace and a cluster administrator must create them.       | `""` |
| `AFFINITY` | Affinity applied to pods created by the daemonset       | `'{}'` |
| `KIP_IMAGE` | The image puller image to copy the `sleep` binary from | `quay.io/eclipse/kubernetes-image-puller:next` |

### Configuration - Helm 

The following values can be set:

| Value                            | Usage                                                        | Default                                               |
| -------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------- |
| `deploymentName`                 | The value of `DAEMONSET_NAME` to be set in the ConfigMap, as well as the name of the deployment     | `kubernetes-image-puller`                             |
| `image.repository`               | The repository to pull the image from                        | `quay.io/eclipse/kubernetes-image-puller`             |
| `image.tag`                      | The image tag to pull                                        | `next`                                                |
| `serviceAccount.name`            | The name of the ServiceAccount to create                     | `k8s-image-puller`                                    |
| `configMap.name`                 | The name of the ConfigMap to create                          | `k8s-image-puller`                                    |
| `configMap.images`               | The value of `IMAGES` to be set in the ConfigMap             | // TODO create a reasonable set of default containers |
| `configMap.cachingIntervalHours` | The value of `CACHING_INTERVAL_HOURS` to be set in the ConfigMap | `"1"`                                                 |
| `configMap.cachingMemoryRequest` | The value of `CACHING_MEMORY_REQUEST` to be set in the ConfigMap | `"10Mi"`                                              |
| `configMap.cachingMemoryLimit`   | The value of `CACHING_MEMORY_LIMIT` to be set in the ConfigMap | `"20Mi"`                                              |
| `configMap.cachingCpuRequest`    | The value of `CACHING_CPU_REQUEST` to be set in the ConfigMap | `.05`                                                 |
| `configMap.cachingCpuLimit`      | The value of `CACHING_CPU_LIMIT` to be set in the ConfigMap  | `.2`                                                  |
| `configMap.nodeSelector`         | The value of `NODE_SELECTOR` to be set in the ConfigMap      | `"{}"`                                                |
| `configMap.imagePullSecrets` | The value of `IMAGE_PULL_SECRETS`       | `""` |
| `configMap.affinity`         | The value of `AFFINITY` to be set in the ConfigMap      | `"{}"`                                                |

### Configuration - OpenShift

The following values can be set:

| Parameter | Usage | Default |
| -- | -- | -- |
| `SERVICEACCOUNT_NAME`             | Name of service account used by main pod | `k8s-image-puller` |
| `IMAGE`                           | Name of image used for main pod | `quay.io/eclipse/kubernetes-image-puller` |
| `IMAGE_TAG`                       | Tag of image used for main pod | `next` |
| `DAEMONSET_NAME` | The value of `DAEMONSET_NAME` to be set in the ConfigMap | `"kubernetes-image-puller"` |
| `DEPLOYMENT_NAME` | The name of the image puller deployment | `"kubernetes-image-puller"` |
| `CACHING_INTERVAL_HOURS` | The value of `CACHING_INTERVAL_HOURS` to be set in the ConfigMap | `"1"` |
| `CACHING_MEMORY_REQUEST` | The value of `CACHING_MEMORY_REQUEST` to be set in the ConfigMap | `"10Mi"` |
| `CACHING_MEMORY_LIMIT` | The value of `CACHING_MEMORY_LIMIT` to be set in the ConfigMap | `"20Mi"` |
| `CACHING_CPU_REQUEST` | The value of `CACHING_CPU_REQUEST` to be set in the ConfigMap | `.05` |
| `CACHING_CPU_LIMIT` | The value of `CACHING_CPU_LIMIT` to be set in the ConfigMap | `.2` |
| `NAMESPACE` | The value of `NAMESPACE` to be set in the ConfigMap | `k8s-image-puller` |
| `NODE_SELECTOR` | The value of `NODE_SELECTOR` to be set in the ConfigMap | `"{}"` |
| `IMAGE_PULL_SECRETS` | The value of `IMAGE_PULL_SECRETS`       | `""` |
| `AFFINITY` | The value of `AFFINITY` to be set in the ConfigMap | `"{}"` |

### Installation - Helm

`kubectl create namespace k8s-image-puller`

`helm install kubernetes-image-puller -n k8s-image-puller deploy/helm`

To set values, change `deploy/helm/values.yaml` or use `--set property.name=value`

### Installation - OpenShift

#### Openshift special consideration - Project Quotas

OpenShift has a notion of [project quotas](https://docs.openshift.com/container-platform/4.3/applications/quotas/quotas-setting-per-project.html) to limit the aggregate resource consumption per project/namespace.  The namespace that the image puller is deployed in must have enough memory and CPU to run each container for each node in the cluster:

```
(memory/CPU limit) * (number of images) * (number of nodes in cluster)
```

For example, running the image puller that caches 5 images on 20 nodes, with a container memory limit of `5Mi`, your namespace would need a quota of `500Mi`.

#### Installing the image puller

`oc new-project k8s-image-puller`

`oc process -f deploy/openshift/serviceaccount.yaml | oc apply -f -`

`oc process -f deploy/openshift/configmap.yaml | oc apply -f -`

`oc process -f deploy/openshift/app.yaml | oc apply -f -`

To change parameters, add `-p PARAM=value` to the `oc process` command, before piping to `oc apply`.

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
GOOS=linux go build -v -o ./bin/kubernetes-image-puller ./cmd/main.go
```
Make docker image:
```bash
docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} -f ./docker/Dockerfile .
```

## Testing
To run the unit tests:
```shell
make test
```

End to end tests require [kind](https://github.com/kubernetes-sigs/kind).
Note that kind should not be installed with `go get` from this repository's directory.

```shell
cd $HOME && GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0 && cd ~-

./hack/run-e2e.sh
```

Will start a kind cluster and run the end-to-end tests in `./e2e`.  To remove the cluster after running the tests, pass the `--rm` argument to the script, or run `kind delete cluster --name k8s-image-puller-e2e`.

## Scratch Images
The image puller now supports pre-pulling scratch images.
Previously the image puller was not able to pull scratch images, as they do not contain a `sleep` command.

However, the daemonset created by the Kubernetes Image Puller now:
1. creates an `initContainer` that copies a golang-based `sleep` binary to a common `kip` volume.
2. creates containers `volumeMounts` set to the `kip` volume, and with `command` set to `/kip/sleep 720h`

As a result, every container (including scratch image containers) uses the provided golang-based `sleep` binary.
