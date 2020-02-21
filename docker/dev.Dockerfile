FROM gcr.io/distroless/base
COPY "./bin/kubernetes-image-puller" "/"
CMD ["/kubernetes-image-puller"]
