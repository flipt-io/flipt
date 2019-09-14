FROM debian:stable-slim

RUN apt-get update && \
	apt-get install curl -y && \
	apt-get clean -y

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
