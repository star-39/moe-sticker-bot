#!/bin/bash

GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

#################################

c1=$(buildah from debian:sid)

buildah run $c1 -- apt update -y
buildah run $c1 -- apt install python3 python3-pip imagemagick curl libarchive-tools libmariadb-dev -y

buildah run $c1 -- curl -Lo /usr/bin/ffmpeg "https://github.com/star-39/ffmpeg-nano-static/releases/download/v5fcf2d7b5f818d68f77e56ff2b2b3d50c50b90be/ffmpeg"
buildah run $c1 -- chmod +x /usr/bin/ffmpeg

buildah run $c1 -- curl -Lo /requirements.txt https://github.com/star-39/moe-sticker-bot/raw/master/requirements.txt

buildah run $c1 -- pip3 install wheel setuptools
buildah run $c1 -- pip3 install -r /requirements.txt
buildah run $c1 -- pip3 cache purge

buildah run $c1 -- apt autoremove python3-pip -y
buildah run $c1 -- apt install python3-setuptools -y
buildah run $c1 -- apt autoclean

buildah config --cmd '' $c1
buildah config --entrypoint "cd /moe-sticker-bot-master && /usr/bin/python3 main.py" $c1

# Fix python3.8+'s problem.
buildah config --env COLUMNS=80 $c1

buildah commit $c1 moe-sticker-bot:base

buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:base

#################################

c1=$(buildah from ghcr.io/star-39/moe-sticker-bot:base)

buildah run $c1 -- curl -Lo /moe-sticker-bot.zip https://github.com/star-39/moe-sticker-bot/archive/refs/heads/master.zip
buildah run $c1 -- bsdtar -xvf /moe-sticker-bot.zip -C /

buildah commit $c1 moe-sticker-bot:latest

buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:amd64
buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:latest
