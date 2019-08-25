FROM markphelps/flipt:latest

RUN apk update && apk add --no-cache wget ca-certificates nss && \
  wget -O mkcert https://github.com/FiloSottile/mkcert/releases/download/v1.4.0/mkcert-v1.4.0-linux-amd64 && chmod 755 mkcert

RUN ./mkcert -install -cert-file /var/opt/flipt/cert.pem -key-file /var/opt/flipt/key.pem 0.0.0.0 flipt
