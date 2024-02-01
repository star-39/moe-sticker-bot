#!/usr/bin/bash
GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

#AArch64
c1=$(buildah from --arch=arm64 docker://arm64v8/nginx:latest)

# Copy nginx template and app link validation json.
buildah copy $c1 web/nginx/default.conf.template /etc/nginx/templates/
buildah copy $c1 web/nginx/assetlinks.json /www/.well-known/assetlinks.json

# Build react app
cd web/webapp3/
npm install
PUBLIC_URL=/webapp REACT_APP_HOST=msb.cloudns.asia npm run build
buildah copy $c1 build/ /webapp
cd ../..

buildah commit $c1 moe-sticker-bot:msb_nginx_aarch64

buildah push moe-sticker-bot:msb_nginx_aarch64 ghcr.io/star-39/moe-sticker-bot:msb_nginx_aarch64
