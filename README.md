[![Build Status](https://ci.centos.org/job/devtools-kubernetes-image-puller-build-master/badge/icon)](https://ci.centos.org/job/devtools-kubernetes-image-puller-build-master/)

## Requirements
This is an upstream version of the [kubernetes-image-puller](https://github.com/redhat-developer/kubernetes-image-puller).  Where the downstream puller requires integrations with [fabric-oso-proxy](https://github.com/fabric8-services/fabric8-oso-proxy) and [fabric8-auth](https://github.com/fabric8-services/fabric8-auth), and impersonates users in multiple clusters, this application is meant to run on a single cluster, and pre-pull Eclipse Che images.

To cache images, this app creates a Daemonset on the desired cluster, which in turn creates a pod on each node in the cluster consisting of a list of containers with command `sleep infinity`. This ensures that all nodes in the cluster have those images cached. We also periodically check the health of the daemonset and re-create it if necessary.

The application can be deployed via Helm or by processing and applying OpenShift Templates.

## Configuration
Configuration is done via env vars pulled from `./deploy/helm/templates/configmap.yaml`, or `./deploy/openshift/configmap.yaml`, depending on the deployment method.
The config values to be set are:

| Env Var | Usage | Default |
| -- | -- | -- |
| `CACHING_INTERVAL_HOURS` | Interval, in hours, between checking health of daemonsets | `"1"` |
| `CACHING_MEMORY_REQUEST` | The container memory request | `10Mi` |
| `CACHING_MEMORY_LIMIT` | The container memory limit | `20Mi` |
| `DAEMONSET_NAME`         | Name of daemonset to be created | `kubernetes-image-puller` |
| `NAMESPACE`              | Namespace where daemonset is to be created. Shared for all users | `k8s-image-puller` |
| `IMAGES`                 | List of images to be cached, in the format `<name>=<image>;...` | Contains a default list of images, but should be configured when deploying |
| `NODE_SELECTOR` | The node selector to run the daemonsets on particular nodes. | `'{}'` |

### Configuration - Helm 

The following values can be set:

| Value                            | Usage                                                        | Default                                               |
| -------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------- |
| `appName`                        | The value of `DAEMONSET_NAME` to be set in the ConfigMap     | `kubernetes-image-puller`                             |
| `image.repository`               | The repository to pull the image from                        | `quay.io/eclpise/kubernetes-image-puller`             |
| `image.tag`                      | The image tag to pull                                        | `latest`                                              |
| `serviceAccount.name`            | The name of the ServiceAccount to create                     | `k8s-image-puller`                                    |
| `configMap.name`                 | The name of the ConfigMap to create                          | `k8s-image-puller`                                    |
| `configMap.images`               | The value of `IMAGES` to be set in the ConfigMap             | // TODO create a reasonable set of default containers |
| `configMap.cachingIntervalHours` | The value of `CACHING_INTERVAL_HOURS` to be set in the ConfigMap | `"1"`                                                 |
| `configMap.cachingMemoryRequest` | The value of `CACHING_MEMORY_REQUEST` to be set in the ConfigMap | `"10Mi"`                                              |
| `configMap.cachingMemeryLimit`   | The value of `CACHING_MEMORY_LIMIT` to be set in the ConfigMap | `"20Mi"`                                              |
| `configMap.nodeSelector`         | The value of `NODE_SELECTOR` to be set in the ConfigMap      | `"{}"`                                                |

### Configuration - Openshift

The following values can be set:

| Parameter | Usage | Default |
| -- | -- | -- |
| `SERVICEACCOUNT_NAME`             | Name of service account used by main pod | `k8s-image-puller` |
| `IMAGE`                           | Name of image used for main pod | `quay.io/eclpise/kubernetes-image-puller` |
| `IMAGE_TAG`                       | Tag of image used for main pod | `latest` |
| `DAEMONSET_NAME` | The value of `DAEMONSET_NAME` to be set in the ConfigMap | `"kubernetes-image-puller"` |
| `CACHING_INTERVAL_HOURS` | The value of `CACHING_INTERVAL_HOURS` to be set in the ConfigMap | `"1"` |
| `CACHING_MEMORY_REQUEST` | The value of `CACHING_MEMORY_REQUEST` to be set in the ConfigMap | `"10Mi"` |
| `CACHING_MEMORY_LIMIT` | The value of `CACHING_MEMORY_LIMIT` to be set in the ConfigMap | `"20Mi"` |
| `NODE_SELECTOR` | The value of `NODE_SELECTOR` to be set in the ConfigMap | `"{}"` |

### Installation - Helm

`kubectl create namespace k8s-image-puller`

`helm install kubernetes-image-puller -n k8s-image-puller deploy/helm`

To set values, changes `deploy/helm/values.yaml` or use `--set property.name=value`

### Installation - Openshift

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
GOOS=linux go build -v -o ./bin/che-image-caching ./cmd/main.go
```
Make docker image:
```bash
docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} .
```

## Testing locally
It's possible to run a simplified version of kubernetes-image-puller locally in minishift. This version avoids most of the complexity in the `oso-proxy` version, so its usefulness is limited.

Note: to run the commands below, you will need to be an admin user.

```
oc adm policy add-cluster-role-to-user cluster-admin admin
oc login -u admin -p any
oc new-project k8s-image-puller
make docker
make local-setup
make local-deploy
```

This uses the `yaml` files in the `./deploy` directory to create a kubernetes image puller locally, that, in turn, creates a daemonset in the current namespace.