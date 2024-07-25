FROM golang:1.22.5 AS builder
WORKDIR /go/src/vacant.sh/
COPY go.mod go.sum ./
RUN go mod download
ADD . vmanager
RUN cd vmanager && make webhook-manager

FROM alpine:latest
COPY --from=builder /go/src/vacant.sh/vmanager/_output/webhook-manager /webhook-manager
ENTRYPOINT ["/webhook-manager"]
