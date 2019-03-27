FROM alpine:latest

RUN apk update && \
    apk --no-cache add \
      ca-certificates \
      curl \
      git \
      make \
      bash \
      jq

RUN git --git-dir=/dev/null clone --depth=1 https://github.com/robwhitby/shakedown.git /vendor/shakedown
RUN git --git-dir=/dev/null clone --depth=1 https://github.com/vishnubob/wait-for-it.git /vendor/wait-for-it

WORKDIR /app

COPY ./test.sh .
