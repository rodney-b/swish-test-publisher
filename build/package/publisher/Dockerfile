FROM golang:latest AS prebuilder

ENV GOARCH=arm64
ENV GOOS=linux
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod
ENV CGO_ENABLED=0
# disabling cgo is necessary for alpine linux which uses musl libc rather than glibc

WORKDIR /app

# Make it so go modules will pull from our private repos
RUN --mount=type=secret,id=github-token \
  git config --global url.https://$(cat /run/secrets/github-token)@github.com/.insteadOf https://github.com

FROM prebuilder AS builder
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=bind,target=. \
  go build -o /bin/publisher cmd/publisher/main.go && chmod 755 /bin/publisher

FROM alpine:latest AS app
# Create non-root user that matches the container UID/GID (1000:1000)
RUN addgroup -g 1000 appgroup && adduser -D -u 1000 -G appgroup appuser
COPY --from=builder --chown=1000:1000 --chmod=0755 /bin/publisher /usr/local/bin/publisher
# CA certs and a writable tmp
RUN apk add --no-cache ca-certificates && mkdir -p /tmp && chown 1000:1000 /tmp

USER 1000:1000
ENTRYPOINT ["/usr/local/bin/publisher"]