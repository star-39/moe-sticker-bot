#!/bin/bash

GITHUB_TOKEN=$1

echo "Building moe-sticker-bot for Github Container Registry!"

c1=$(buildah from debian:11)

# install system dependencies
buildah run $c1 -- apt update -y
buildah run $c1 -- apt install python3 python3-pip imagemagick curl libarchive-tools -y

buildah run $c1 -- curl -Lo /ffmpeg.tar.xz https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz
buildah run $c1 -- tar --strip-components=1 --wildcards -xvf /ffmpeg.tar.xz '*/ffmpeg'
buildah run $c1 -- mv /ffmpeg /usr/bin/ffmpeg
buildah run $c1 -- rm /ffmpeg.tar.xz

# grab sources
buildah run $c1 -- curl -Lo /moe-sticker-bot.zip https://github.com/star-39/moe-sticker-bot/archive/refs/heads/master.zip
buildah run $c1 -- bsdtar -xvf /moe-sticker-bot.zip -C /

# install python dependencies
buildah run $c1 -- pip3 install wheel setuptools
buildah run $c1 -- pip3 install -r /moe-sticker-bot-master/requirements.txt

# finish
buildah config --cmd '' $c1
buildah config --entrypoint "cd /moe-sticker-bot-master && /usr/bin/python3 main.py" $c1

# Fix python3.8+'s problem.
buildah config --env COLUMNS=80 $c1

# clean up
buildah run $c1 -- apt autoremove python3-pip -y
buildah run $c1 -- apt autoclean

buildah commit $c1 moe-sticker-bot

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:amd64
buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:latest
