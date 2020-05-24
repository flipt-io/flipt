FROM debian:stable-slim

RUN apt-get update && \
	apt-get install bats curl jq -y && \
	apt-get clean -y

EXPOSE 8080
