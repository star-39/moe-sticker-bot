#!/bin/bash

GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

#################################

c1=$(buildah from docker://arm64v8/archlinux:latest)

buildah run $c1 -- pacman -Sy
buildah run $c1 -- pacman --noconfirm -S libheif ffmpeg imagemagick curl libarchive 

buildah config --cmd '/moe-sticker-bot' $c1

buildah commit $c1 moe-sticker-bot:base

buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:base

#################################

c1=$(buildah from ghcr.io/star-39/moe-sticker-bot:base_aarch64)

go build
buildah copy $c1 moe-sticker-bot /moe-sticker-bot

buildah commit $c1 moe-sticker-bot:latest

buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:aarch64
