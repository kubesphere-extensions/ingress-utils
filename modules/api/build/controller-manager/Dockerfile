FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
ADD . /workspace

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o controller-manager cmd/controller-manager/main.go

FROM alpine:3.18.3
WORKDIR /
COPY --from=builder /workspace/controller-manager /usr/local/bin/
USER 65532:65532

CMD ["sh"]