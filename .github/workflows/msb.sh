GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

# AMD64
#################################
# Control building base image.
if false ; then

c1=$(buildah from docker://archlinux:latest)

buildah run $c1 -- pacman -Sy
buildah run $c1 -- pacman --noconfirm -S libheif ffmpeg imagemagick curl libarchive python python-pip
buildah run $c1 -- sh -c 'yes | pacman -Scc'

buildah run $c1 -- pip3 install lottie[GIF] cairosvg
 
buildah config --cmd '/moe-sticker-bot' $c1

buildah commit $c1 moe-sticker-bot:base

buildah push moe-sticker-bot:base ghcr.io/star-39/moe-sticker-bot:base

fi
# End building base image.

# Build container image.
c1=$(buildah from ghcr.io/star-39/moe-sticker-bot:base)

# Build MSB go bin
go version
go build -o moe-sticker-bot cmd/moe-sticker-bot/main.go 
buildah copy $c1 moe-sticker-bot /moe-sticker-bot

# Copy tools.
buildah copy $c1 tools/msb_kakao_decrypt.py /usr/local/bin/msb_kakao_decrypt.py

buildah commit $c1 moe-sticker-bot:latest

buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:amd64
buildah push moe-sticker-bot ghcr.io/star-39/moe-sticker-bot:latest

# AArch64
#################################
if false ; then

c1=$(buildah from docker://lopsided/archlinux-arm64v8:latest)

buildah run $c1 -- pacman -Sy
buildah run $c1 -- pacman --noconfirm -S libheif ffmpeg imagemagick curl libarchive python python-pip
buildah run $c1 -- sh -c 'yes | pacman -Scc'

buildah run $c1 -- pip3 install lottie[GIF] cairosvg

buildah config --cmd '/moe-sticker-bot' $c1

buildah commit $c1 moe-sticker-bot:base_aarch64

buildah push moe-sticker-bot:base_aarch64 ghcr.io/star-39/moe-sticker-bot:base_aarch64

fi
#################################

c1=$(buildah from ghcr.io/star-39/moe-sticker-bot:base_aarch64)

# Build MSB go bin
go version
GOOS=linux GOARCH=arm64 go build -o moe-sticker-bot cmd/moe-sticker-bot/main.go 
buildah copy $c1 moe-sticker-bot /moe-sticker-bot

# Copy tools.
buildah copy $c1 tools/msb_kakao_decrypt.py /usr/local/bin/msb_kakao_decrypt.py

buildah commit $c1 moe-sticker-bot:aarch64

buildah push moe-sticker-bot:aarch64 ghcr.io/star-39/moe-sticker-bot:aarch64
