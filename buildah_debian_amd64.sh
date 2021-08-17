#!/bin/bash

GITHUB_TOKEN=$1

echo "Building moe-sticker-bot for Github Container Registry!"

c1=$(buildah from debian:sid)

# install system dependencies
buildah run $c1 -- apt update -y
buildah run $c1 -- apt install ffmpeg python3 python3-pip imagemagick curl libarchive-tools -y

# grab sources
buildah run $c1 -- curl -Lo /moe-sticker-bot.zip https://github.com/star-39/moe-sticker-bot/archive/refs/heads/master.zip
buildah run $c1 -- bsdtar -xvf /moe-sticker-bot.zip -C /

# install python dependencies
buildah run $c1 -- pip3 install wheel
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
