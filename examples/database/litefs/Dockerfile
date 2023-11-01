FROM alpine

COPY --from=flyio/litefs:0.5 /usr/local/bin/litefs /usr/local/bin/litefs
COPY --from=flipt/flipt:latest /flipt /usr/local/bin/flipt

RUN mkdir -p /etc/flipt/config && \
    mkdir -p /var/opt/flipt

COPY --from=flipt/flipt:latest /etc/flipt/config/default.yml /etc/flipt/config/default.yml
RUN chown -R root:root /var/opt/flipt /etc/flipt

ADD litefs.yml /etc/litefs.yml

RUN apk add bash fuse3 sqlite ca-certificates

ENTRYPOINT litefs mount
