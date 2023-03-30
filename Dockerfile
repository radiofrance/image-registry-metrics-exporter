FROM docker.io/golang:1.20.2 as builder
ARG VERSION=devel

WORKDIR /build
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s -X 'main.version=${VERSION}'" ./cmd/image-registry-metrics-exporter

FROM scratch

COPY --from=builder /build/image-registry-metrics-exporter /image-registry-metrics-exporter

EXPOSE 9252
USER 65534

ENTRYPOINT ["/image-registry-metrics-exporter"]
