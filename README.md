# [@moe_sticker_bot](https://t.me/moe_sticker_bot)
A Telegram bot doing sticker stuffs, it imports LINE sticker set to Telegram, downloads LINE and Telegram stickers, creates new sticker set. 

Telegram用萌萌貼圖BOT, 可以從LINE Store導入貼圖包到Telegram, 可以下載LINE和telegram的貼圖包, 可以創建新的貼圖包.

Telegram用萌え萌えのスタンプBOTです。LINEストアからスタンプをTelegramにインポートしたり、LINEとTelegramのスタンプをダウンロードしたり、新しいスタンプセットを作ったり

<img src="https://user-images.githubusercontent.com/75669297/119772095-9f14b280-bef9-11eb-8b99-d13847a26ea7.png" width="500">


## Deployment

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

Fill your Telegram Bot API in config.ini`, and it's ready to fly! Just run with:

```
python3 main.py
```

Or you can specify `BOT_TOKEN` environment variable and run:
```
BOT_TOKEN=your_token python3 main.py
```

This software supports all platforms python supports, including Linux, Windows and Mac.

### Deploy with pre-built containers
A pre-built OCI container is available at https://github.com/users/star-39/packages/container/package/moe-sticker-bot

Follow GitHub's guide on pulling the container, or you can just run with
```
podman run -dt --rm -e BOT_TOKEN=your_bot_token_here ghcr.io/star-39/moe-sticker-bot:latest
```
Of course, you can use docker as well.

### Step by step manual deployment
#### Fedora / RHEL / CentOS
You may need to add [RPM Fusion](https://rpmfusion.org/Configuration) first.
```
dnf install ffmpeg ImageMagick libwebp bsdtar curl 
pip3 install python-telegram-bot requests beautifulsoup4 emoji lottie[GIF]
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
### open config.ini and fill your bot api.
python3 main.py
```

#### Ubuntu / Debian
```
apt install imagemagick libwebp6 ffmpeg libarchive-tools curl 
pip3 install python-telegram-bot requests beautifulsoup4 emoji lottie[GIF]
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
### open config.ini and fill your bot api.
python3 main.py
```

## Known issue

### No response?
The bot might have encountered some unhandled exception, please report to issues.

### LINE animated stickers will never be supported due to Telegram's restrictions
LINE's animated stickers are in APNG bitmap, however,
Telegram's animated sticker only allow vector images, which is a completely different.
Telegram even sets the size limit to 64KB, hence
converting stickers to Telegram animated format is impossible.

## License
The GPL V3 License

![image](https://www.gnu.org/graphics/gplv3-127x51.png)
