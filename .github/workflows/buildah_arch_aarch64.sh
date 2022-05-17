#!/bin/bash

GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

#################################

# c1=$(buildah from docker://lopsided/archlinux-arm64v8:latest)

# buildah run $c1 -- pacman -Sy
# buildah run $c1 -- pacman --noconfirm -S libheif ffmpeg imagemagick curl libarchive python python-pip
# buildah run $c1 -- sh -c 'yes | pacman -Scc'

# buildah run $c1 -- pip3 install lottie[GIF] cairosvg

# buildah config --cmd '/moe-sticker-bot' $c1

# buildah commit $c1 moe-sticker-bot:base_aarch64

# buildah push moe-sticker-bot:base_aarch64 ghcr.io/star-39/moe-sticker-bot:base_aarch64

#################################

c1=$(buildah from ghcr.io/star-39/moe-sticker-bot:base_aarch64)

go version
GOOS=linux GOARCH=arm64 go build
buildah copy $c1 moe-sticker-bot /moe-sticker-bot

buildah commit $c1 moe-sticker-bot:aarch64

buildah push moe-sticker-bot:aarch64 ghcr.io/star-39/moe-sticker-bot:aarch64
