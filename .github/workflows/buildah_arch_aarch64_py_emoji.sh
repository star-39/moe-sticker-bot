
GITHUB_TOKEN=$1

buildah login -u star-39 -p $GITHUB_TOKEN ghcr.io

c1=$(buildah from docker://lopsided/archlinux-arm64v8:latest)

buildah run $c1 -- pacman -Sy
buildah run $c1 -- pacman --noconfirm -S python python-pip
buildah run $c1 -- sh -c 'yes | pacman -Scc'

buildah run $c1 -- pip3 install emoji Flask waitress
 
buildah copy $c1 microservices/py_emoji/main.py /main.py

buildah config --cmd './main.py' $c1

buildah commit $c1 moe-sticker-bot:py_emoji

buildah push moe-sticker-bot:py_emoji ghcr.io/star-39/moe-sticker-bot:py_emoji
