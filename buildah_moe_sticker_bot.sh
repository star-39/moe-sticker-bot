#!/bin/bash

echo "Building moe-sticker-bot for Github Container Registry!"

c1=$(buildah from ubi8)
crootfs=$(buildah mount $c1)

# prepare repos
buildah run $c1 -- dnf config-manager --enable codeready-builder-for-rhel-8-x86_64-rpms
buildah run $c1 -- dnf install https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm https://mirrors.rpmfusion.org/free/el/rpmfusion-free-release-8.noarch.rpm https://mirrors.rpmfusion.org/nonfree/el/rpmfusion-nonfree-release-8.noarch.rpm -y

# install dependencies
buildah run $c1 -- dnf install python39 bsdtar ImageMagick ffmpeg libwebp curl -y
buildah run $c1 -- pip3 wheel install python-telegram-bot requests beautifulsoup4 emoji lottie[GIF]

buildah run $c1 -- dnf clean all

# grab sources
cd $crootfs
curl -Lo moe-sticker-bot.zip https://github.com/star-39/moe-sticker-bot/archive/refs/heads/master.zip
bsdtar -xvf moe-sticker-bot.zip
cd -

# finish
buildah config --cmd '' $c1
buildah config --entrypoint "cd /moe-sticker-bot-master && /usr/bin/python3 main.py" $c1

buildah commit $c1 moe-sticker-bot

buildah rm $c1
