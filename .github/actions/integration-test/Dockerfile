FROM debian:stable-slim

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add - && \
	echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list

RUN apt-get update && \
	apt-get install -y bats curl jq yarn \
	# chromium dependencies
	libgbm1 \
    # webkit dependencies
	libwoff1 libopus0 libwebp6 libwebpdemux2 libenchant1c2a \
	libgudev-1.0-0 libsecret-1-0 libhyphen0 libgdk-pixbuf2.0-0 libegl1 libgles2 \
	libevent-2.1-6 libnotify4 libvpx5 libxslt1.1 \
    # ffmpeg
    ffmpeg && \
	apt-get clean -y

EXPOSE 8080
