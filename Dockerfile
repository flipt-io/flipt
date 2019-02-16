FROM alpine:latest
LABEL maintainer="mark.aaron.phelps@gmail.com"

RUN mkdir -p /etc/flipt && \
    mkdir -p /var/opt/flipt

COPY flipt /
COPY config /etc/flipt/config

EXPOSE 8080
EXPOSE 9000

CMD ["./flipt"]
