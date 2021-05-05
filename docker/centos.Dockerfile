FROM golang:1.13.8 AS dependencies

WORKDIR /kubernetes-image-puller

# Populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS builder

COPY . .
RUN make build

FROM registry.centos.org/centos:7

RUN yum update -y -d 1 \
    && yum clean all -y \
    && rm -rf /var/cache/yum

COPY --from=builder "/kubernetes-image-puller/bin/kubernetes-image-puller" "/"
COPY --from=builder /kubernetes-image-puller/bin/sleep /bin/sleep
CMD ["/kubernetes-image-puller"]
