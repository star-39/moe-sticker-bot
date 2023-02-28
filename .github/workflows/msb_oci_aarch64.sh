#!/bin/sh

GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

# AArch64
#################################
if false ; then

c1=$(buildah from docker://lopsided/archlinux-arm64v8:latest)

buildah run $c1 -- pacman -Sy
buildah run $c1 -- pacman --noconfirm -S libwebp libheif imagemagick curl gifsicle libarchive python python-pip make gcc

buildah run $c1 -- pip3 install emoji rlottie-python

buildah run $c1 -- pacman --noconfirm -Rsc make gcc python-pip
buildah run $c1 -- sh -c 'yes | pacman -Scc'


buildah config --cmd '/moe-sticker-bot' $c1

buildah commit $c1 moe-sticker-bot:base_aarch64

buildah push moe-sticker-bot:base_aarch64 ghcr.io/star-39/moe-sticker-bot:base_aarch64

fi
#################################

# Build container image.
c1=$(buildah from ghcr.io/star-39/moe-sticker-bot:base_aarch64)


# Install static build of ffmpeg.
curl -JOL "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz"
tar -xvf ffmpeg-master-latest-linuxarm64-gpl.tar.xz
buildah copy $c1 ffmpeg-master-latest-linuxarm64-gpl/bin/ffmpeg /usr/local/bin/ffmpeg


# Build MSB go bin
GOOS=linux GOARCH=arm64 go build -o moe-sticker-bot cmd/moe-sticker-bot/main.go 
buildah copy $c1 moe-sticker-bot /moe-sticker-bot

# Copy tools.
buildah copy $c1 tools/msb_kakao_decrypt.py /usr/local/bin/msb_kakao_decrypt.py
buildah copy $c1 tools/msb_emoji.py /usr/local/bin/msb_emoji.py
buildah copy $c1 tools/msb_rlottie.py /usr/local/bin/msb_rlottie.py

buildah commit $c1 moe-sticker-bot:aarch64

buildah push moe-sticker-bot:aarch64 ghcr.io/star-39/moe-sticker-bot:aarch64
