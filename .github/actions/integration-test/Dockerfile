FROM debian:stable-slim

RUN apt-get update && \
	apt-get install -y bats curl jq gnupg \
	# chromium dependencies
	libgbm1 \
    # webkit dependencies
	libwoff1 libopus0 libwebp6 libwebpdemux2 libenchant1c2a \
	libgudev-1.0-0 libsecret-1-0 libhyphen0 libgdk-pixbuf2.0-0 libegl1 libgles2 \
	libevent-2.1-6 libnotify4 libvpx5 libxslt1.1 \
    # ffmpeg
    ffmpeg

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - && \
	echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

RUN apt-get update && \
	apt-get install -y yarn npm && \
	apt-get clean -y

EXPOSE 8080
