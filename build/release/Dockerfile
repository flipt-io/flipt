# https://goreleaser.com/docker/

FROM alpine:3.9
LABEL maintainer="mark.aaron.phelps@gmail.com"

RUN apk add --no-cache postgresql-client \
    openssl \
    ca-certificates

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

COPY flipt /
COPY config /etc/flipt/config

EXPOSE 8080
EXPOSE 9000

CMD ["./flipt"]
