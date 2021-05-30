# [@moe_sticker_bot](https://t.me/moe_sticker_bot)
A Telegram bot doing sticker stuffs, it imports LINE sticker set to Telegram, downloads LINE and Telegram stickers, creates new sticker set. 

Telegram用萌萌貼圖BOT, 可以從LINE Store導入貼圖包到Telegram, 可以下載LINE和telegram的貼圖包, 可以創建新的貼圖包.

Telegram用萌え萌えのスタンプBOTです。LINEストアからスタンプをTelegramにインポートしたり、LINEとTelegramのスタンプをダウンロードしたり、新しいスタンプセットを作ったり

![](https://user-images.githubusercontent.com/75669297/120078508-deeebc00-c0ea-11eb-8fe1-f0a51dae4267.png)

## Deployment
### Deploy with pre-built containers
A pre-built OCI container is available at https://github.com/users/star-39/packages/container/package/moe-sticker-bot

Simply run:
```
podman run -dt -e BOT_TOKEN=your_bot_token ghcr.io/star-39/moe-sticker-bot:latest
```
Of course, you can use docker as well, just replace `podman` with `docker`.

### Python Dependencies
* [python-telegram-bot](https://github.com/python-telegram-bot/python-telegram-bot)
* requests
* bs4
* emoji
* lottie

### System Dependencies
* python 3.9+
* ImageMagick
* libwebp
* ffmpeg (with libwebp and libx264)
* bsdtar (libarchive-tools)
* curl

Specify `BOT_TOKEN` environment variable and run:
```
BOT_TOKEN=your_bot_token python3 main.py
```

This software supports all platforms python supports, including Linux, Windows and Mac.

### Step by step manual deployment
#### Fedora / RHEL / CentOS
You may need to add [RPM Fusion](https://rpmfusion.org/Configuration) first.
```
dnf install git ffmpeg ImageMagick libwebp bsdtar curl 
pip3 install python-telegram-bot requests beautifulsoup4 emoji lottie[GIF]
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
BOT_TOKEN=your_bot_token python3 main.py
```

#### Ubuntu / Debian
```
apt install git imagemagick libwebp6 ffmpeg libarchive-tools curl 
pip3 install python-telegram-bot requests beautifulsoup4 emoji lottie[GIF]
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
BOT_TOKEN=your_bot_token python3 main.py
```

#### Windows
Using Windows Powershell:
```
# install scoop if you have not
Set-ExecutionPolicy RemoteSigned -scope CurrentUser
iwr -useb get.scoop.sh | iex
scoop install git
scoop install ffmpeg-nightly-shared imagemagick python sudo
pip install python-telegram-bot requests beautifulsoup4 emoji lottie[GIF]
sudo New-Item -ItemType symboliclink -path C:\Windows\system32\bsdtar.exe -target C:\Windows\system32\tar.exe
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
$ENV:BOT_TOKEN=your_bot_token ; python main.py
```


## Known issue
### No response?
The bot might have encountered some unhandled exception, try sending `/cancel` or report to GitHub issues.

### LINE animated stickers will never be supported due to Telegram's restrictions
LINE's animated stickers are in APNG bitmap, however,
Telegram's animated sticker only allow vector images, which is a completely different.
Telegram even sets the size limit to 64KB, hence
converting stickers to Telegram animated format is impossible.

## License
The GPL V3 License

![image](https://www.gnu.org/graphics/gplv3-with-text-136x68.png)
