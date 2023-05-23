FROM docker.io/golang:1.20.2 as builder
ARG VERSION=devel

WORKDIR /build
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s -X 'main.version=${VERSION}'" ./cmd/image-registry-metrics-exporter

FROM docker.io/alpine:3.18.0
# renovate: datasource=repology depName=alpine_3_18/ca-certificates versioning=loose
ARG CA_CERTIFICATES_VERSION=20230506-r0

COPY --from=builder /build/image-registry-metrics-exporter /image-registry-metrics-exporter

RUN apk add --no-cache ca-certificates=${CA_CERTIFICATES_VERSION}

EXPOSE 9252
USER 65534

ENTRYPOINT ["/image-registry-metrics-exporter"]
