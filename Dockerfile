FROM golang:1.18 AS builder

LABEL maintainer="darmiel <hi@d2a.io>"
LABEL org.opencontainers.image.source = "https://github.com/RALF-Life/engine"

WORKDIR /usr/src/app
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install and cache dependencies (by @montanaflynn)
# https://github.com/montanaflynn/golang-docker-cache
COPY go.mod go.sum ./
RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

# Copy remaining source
COPY /actions ./actions
COPY /engine ./engine
COPY /model ./model
COPY /server ./server
COPY /internal ./internal
COPY /cmd ./cmd
COPY go.mod .
COPY go.sum .

# Build from sources
RUN GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0 \
    go build \
    -o engine-server \
    ./cmd/server/main.go


FROM alpine:3.15


ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

RUN addgroup -S nonroot \
    && adduser -S nonroot -G nonroot \
    && chown nonroot:nonroot /zoneinfo.zip

USER nonroot

COPY --from=builder /usr/src/app/engine-server .

EXPOSE 1887

ENTRYPOINT [ "/engine-server" ]
