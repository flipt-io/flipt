FROM debian:stable-slim

RUN apt-get update && \
	apt-get install curl jq -y && \
	apt-get clean -y

EXPOSE 8080

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
