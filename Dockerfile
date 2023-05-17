FROM golang:1.20 AS builder

LABEL maintainer="darmiel <hi@d2a.io>"
LABEL org.opencontainers.image.source = "https://github.com/RALF-Life/engine"

WORKDIR /usr/src/app
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN go install github.com/goreleaser/goreleaser@latest

# Install and cache dependencies (by @montanaflynn)
# https://github.com/montanaflynn/golang-docker-cache
COPY go.mod go.sum ./
RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

# Copy remaining source
COPY /pkg ./pkg
COPY /internal ./internal
COPY /cmd ./cmd
COPY go.mod .
COPY go.sum .
COPY .goreleaser.yaml .

RUN goreleaser build --snapshot --single-target

FROM alpine:3.15

ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

RUN addgroup -S nonroot \
    && adduser -S nonroot -G nonroot \
    && chown nonroot:nonroot /zoneinfo.zip

USER nonroot

COPY --from=builder /usr/src/app/dist/server-build_linux_arm64/engine-server .

EXPOSE 80

ENTRYPOINT [ "/engine-server" ]
