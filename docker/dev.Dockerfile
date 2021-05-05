FROM gcr.io/distroless/base
COPY "./bin/kubernetes-image-puller" "/"
COPY ./bin/sleep /bin/sleep
CMD ["/kubernetes-image-puller"]
