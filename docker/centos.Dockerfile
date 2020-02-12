FROM registry.centos.org/centos:7

RUN yum update -y -d 1 \
    && yum clean all -y \
    && rm -rf /var/cache/yum

COPY "./bin/kubernetes-image-puller" "/"
CMD ["/kubernetes-image-puller"]
