[![master](https://ci.centos.org/buildStatus/icon?subject=master&job=devtools-che-incubator-kubernetes-image-puller-build-master/)](https://ci.centos.org/job/devtools-che-incubator-kubernetes-image-puller-build-master/)
[![nightly](https://ci.centos.org/buildStatus/icon?subject=nightly&job=devtools-kubernetes-image-puller-nightly/)](https://ci.centos.org/job/devtools-kubernetes-image-puller-nightly/)

[![Contribute](https://camo.githubusercontent.com/7ca4f6be43fb5eb61a73ba6d40b3481d93ef5813/68747470733a2f2f6368652e6f70656e73686966742e696f2f666163746f72792f7265736f75726365732f666163746f72792d636f6e747269627574652e737667)](https://che.openshift.io/f?url=https://github.com/che-incubator/kubernetes-image-puller)

## Requirements
This is an upstream version of the [kubernetes-image-puller](https://github.com/redhat-developer/kubernetes-image-puller).  Where the downstream puller requires integrations with [fabric-oso-proxy](https://github.com/fabric8-services/fabric8-oso-proxy) and [fabric8-auth](https://github.com/fabric8-services/fabric8-auth), and impersonates users in multiple clusters, this application is meant to run on a single cluster, and pre-pull Eclipse Che images.

To cache images, this app creates a Daemonset on the desired cluster, which in turn creates a pod on each node in the cluster consisting of a list of containers with command `sleep 30d`. This ensures that all nodes in the cluster have those images cached. We also periodically check the health of the daemonset and re-create it if necessary.

The application can be deployed via Helm or by processing and applying OpenShift Templates.

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

### Configuration - Helm 

The following values can be set:

| Value                            | Usage                                                        | Default                                               |
| -------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------- |
| `deploymentName`                 | The value of `DAEMONSET_NAME` to be set in the ConfigMap, as well as the name of the deployment     | `kubernetes-image-puller`                             |
| `image.repository`               | The repository to pull the image from                        | `quay.io/eclpise/kubernetes-image-puller`             |
| `image.tag`                      | The image tag to pull                                        | `latest`                                              |
| `serviceAccount.name`            | The name of the ServiceAccount to create                     | `k8s-image-puller`                                    |
| `configMap.name`                 | The name of the ConfigMap to create                          | `k8s-image-puller`                                    |
| `configMap.images`               | The value of `IMAGES` to be set in the ConfigMap             | // TODO create a reasonable set of default containers |
| `configMap.cachingIntervalHours` | The value of `CACHING_INTERVAL_HOURS` to be set in the ConfigMap | `"1"`                                                 |
| `configMap.cachingMemoryRequest` | The value of `CACHING_MEMORY_REQUEST` to be set in the ConfigMap | `"10Mi"`                                              |
| `configMap.cachingMemeryLimit`   | The value of `CACHING_MEMORY_LIMIT` to be set in the ConfigMap | `"20Mi"`                                              |
| `configMap.cachingCpuRequest`    | The value of `CACHING_CPU_REQUEST` to be set in the ConfigMap | `.05`                                                 |
| `configMap.cachingCpuLimit`      | The value of `CACHING_CPU_LIMIT` to be set in the ConfigMap  | `.2`                                                  |
| `configMap.nodeSelector`         | The value of `NODE_SELECTOR` to be set in the ConfigMap      | `"{}"`                                                |

### Configuration - Openshift

The following values can be set:

| Parameter | Usage | Default |
| -- | -- | -- |
| `SERVICEACCOUNT_NAME`             | Name of service account used by main pod | `k8s-image-puller` |
| `IMAGE`                           | Name of image used for main pod | `quay.io/eclpise/kubernetes-image-puller` |
| `IMAGE_TAG`                       | Tag of image used for main pod | `latest` |
| `DAEMONSET_NAME` | The value of `DAEMONSET_NAME` to be set in the ConfigMap | `"kubernetes-image-puller"` |
| `DEPLOYMENT_NAME` | The name of the image puller deployment | `"kubernetes-image-puller"` |
| `CACHING_INTERVAL_HOURS` | The value of `CACHING_INTERVAL_HOURS` to be set in the ConfigMap | `"1"` |
| `CACHING_MEMORY_REQUEST` | The value of `CACHING_MEMORY_REQUEST` to be set in the ConfigMap | `"10Mi"` |
| `CACHING_MEMORY_LIMIT` | The value of `CACHING_MEMORY_LIMIT` to be set in the ConfigMap | `"20Mi"` |
| `CACHING_CPU_REQUEST` | The value of `CACHING_CPU_REQUEST` to be set in the ConfigMap | `.05` |
| `CACHING_CPU_LIMIT` | The value of `CACHING_CPU_LIMIT` to be set in the ConfigMap | `.2` |
| `NODE_SELECTOR` | The value of `NODE_SELECTOR` to be set in the ConfigMap | `"{}"` |

### Installation - Helm

`kubectl create namespace kubernetes-image-puller`

`helm install kubernetes-image-puller -n k8s-image-puller deploy/helm`

To set values, changes `deploy/helm/values.yaml` or use `--set property.name=value`

### Installation - Openshift

#### Openshift special consideration - Project Quotas

OpenShift has a notion of [project quotas](https://docs.openshift.com/container-platform/4.3/applications/quotas/quotas-setting-per-project.html) to limit the aggregate resource consumption per project/namespace.  The namespace that the image puller is deployed in must have enough memory and CPU to run each container for each node in the cluster:

```
(memory/CPU limit) * (number of images) * (number of nodes in cluster)
```

For example, running the image puller that caches 5 images on 20 nodes, with a container memory limit of `20Mi`, your namespace would need a quota of `2000Mi`.

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
