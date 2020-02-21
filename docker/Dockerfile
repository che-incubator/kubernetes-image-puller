FROM golang:1.13.8 AS dependencies

WORKDIR /kubernetes-image-puller

# Populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS builder

COPY . .
# RUN go test -v ./cfg... ./pkg... ./utils... \
#     && GOOS=linux go build -v -o ./bin/${BINARY_NAME} ./cmd/main.go
RUN make build


FROM gcr.io/distroless/base
COPY --from=builder "/kubernetes-image-puller/bin/kubernetes-image-puller" "/"
CMD ["/kubernetes-image-puller"]
