FROM markphelps/flipt:latest

RUN apk update && apk add --no-cache git bash

RUN git clone https://github.com/vishnubob/wait-for-it.git /tmp && \
    chmod +x /tmp/wait-for-it.sh
