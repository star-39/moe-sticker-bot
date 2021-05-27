#!/bin/bash

echo "Building moe-sticker-bot for Github Container Registry!"

c1=$(buildah from fedora:34)
crootfs=$(buildah mount $c1)

# prepare repos
buildah run $c1 -- dnf install https://mirrors.rpmfusion.org/free/fedora/rpmfusion-free-release-34.noarch.rpm -y

# install dependencies
buildah run $c1 -- dnf install python3.9 python-pip bsdtar ImageMagick ffmpeg libwebp curl -y
buildah run $c1 -- pip3 install wheel python-telegram-bot requests beautifulsoup4 emoji lottie[GIF]

buildah run $c1 -- dnf clean all

# grab sources
cd $crootfs
curl -Lo moe-sticker-bot.zip https://github.com/star-39/moe-sticker-bot/archive/refs/heads/master.zip
bsdtar -xvf moe-sticker-bot.zip
cd -

# finish
buildah config --cmd '' $c1
buildah config --entrypoint "cd /moe-sticker-bot-master && /usr/bin/python3 main.py" $c1

# Fix python3.8+'s problem.
buildah config --env COLUMNS=80

buildah commit $c1 moe-sticker-bot:fedora-v2a1

buildah rm $c1

