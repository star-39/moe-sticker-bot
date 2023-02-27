#!/bin/sh

GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

# AMD64
#################################
if true ; then

c1=$(buildah from docker://archlinux:latest)

buildah run $c1 -- pacman -Sy
buildah run $c1 -- pacman --noconfirm -S libheif imagemagick curl gifsicle exiv2 libarchive python python-pip
buildah run $c1 -- sh -c 'yes | pacman -Scc'

buildah run $c1 -- pip3 install lottie[GIF] cairosvg emoji
 
buildah config --cmd '/moe-sticker-bot' $c1

buildah commit $c1 moe-sticker-bot:base

buildah push moe-sticker-bot:base ghcr.io/star-39/moe-sticker-bot:base

fi
#################################

# Build container image.
c1=$(buildah from ghcr.io/star-39/moe-sticker-bot:base)

# Install static build of ffmpeg.
curl -JOL "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz"
unzip ffmpeg-master-latest-linux64-gpl.tar.xz
buildah copy $c1 ffmpeg-master-latest-linux64-gpl/bin/ffmpeg /usr/local/bin/ffmpeg


# Build MSB go bin
go build -o moe-sticker-bot cmd/moe-sticker-bot/main.go 
buildah copy $c1 moe-sticker-bot /moe-sticker-bot

# Copy tools.
buildah copy $c1 tools/msb_kakao_decrypt.py /usr/local/bin/msb_kakao_decrypt.py
buildah copy $c1 tools/msb_emoji.py /usr/local/bin/msb_emoji.py

buildah commit $c1 moe-sticker-bot:latest

buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:amd64
buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:latest
