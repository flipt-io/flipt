FROM alpine:latest
LABEL maintainer="mark.aaron.phelps@gmail.com"

RUN apk update && apk add postgresql-client

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

COPY flipt /
COPY config /etc/flipt/config

EXPOSE 8080
EXPOSE 9000

CMD ["./flipt"]
