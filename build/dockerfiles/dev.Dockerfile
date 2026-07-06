FROM gcr.io/distroless/base-debian12:nonroot@sha256:4ae8d0163a6f04d96f36e41324d76f00744f0db7545b6d04039c9e6fa1df77f3
COPY "./bin/kubernetes-image-puller" "/"
COPY ./bin/sleep /bin/sleep
CMD ["/kubernetes-image-puller"]
