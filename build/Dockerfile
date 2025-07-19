# https://goreleaser.com/docker/

ARG BINARY=flipt

FROM alpine:3.22.1

LABEL maintainer="dev@flipt.io"
LABEL org.opencontainers.image.name="flipt"
LABEL org.opencontainers.image.source="https://github.com/flipt-io/flipt"

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

COPY $BINARY /

RUN addgroup flipt && \
    adduser -S -D -g '' -G flipt -s /bin/sh flipt && \
    chown -R flipt:flipt /etc/flipt /var/opt/flipt

EXPOSE 8080
EXPOSE 9000

USER flipt

CMD ["./flipt", "server"]
