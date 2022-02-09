# [@moe_sticker_bot](https://t.me/moe_sticker_bot)
A Telegram bot doing sticker stuffs, it imports LINE sticker set to Telegram, downloads LINE and Telegram stickers, creates and manages Telegram sticker set. 

Telegram用萌萌貼圖BOT, 可以從LINE Store匯入貼圖包到Telegram, 可以下載LINE和telegram的貼圖包, 可以創建和管理Telegram貼圖包.

Telegram用萌え萌えのスタンプBOTです。LINEストアからスタンプをTelegramにインポートしたり、LINEとTelegramのスタンプをダウンロードしたり、新しいTelegramステッカーセットを作ったり管理したり、色んなスタンプ関連機能があります。

![](https://user-images.githubusercontent.com/75669297/147678436-10bb9169-efad-4da8-acb5-9996edc78364.png)

Now supports video sticker! 現已支援動態貼圖! アニメーションスタンプ対応！


## Deployment
### Deploy with pre-built containers
A pre-built OCI container is available at https://github.com/users/star-39/packages/container/package/moe-sticker-bot

Simply run:
```
docker run -dt -e BOT_TOKEN=your_bot_token ghcr.io/star-39/moe-sticker-bot:latest
```

### Python Dependencies
* [python-telegram-bot](https://github.com/python-telegram-bot/python-telegram-bot)
* requests
* bs4
* emoji

### System Dependencies
* python 3.9+
* ImageMagick
* libwebp
* bsdtar (libarchive-tools)
* ffmpeg
* curl

Specify `BOT_TOKEN` environment variable and run:
```
BOT_TOKEN=your_bot_token python3 main.py
```

This software supports all platforms python supports, including Linux, Windows and Mac.

However, it's tested on Linux only.

### Manual deployment on Linux
For better performance, it's recommended to use my custom build of FFMpeg:

https://github.com/star-39/ffmpeg-nano-static
```
# For Fedora / RHEL / CentOS etc. (Requires RPM Fusion)
dnf install git ImageMagick libwebp bsdtar curl ffmpeg
# For Debian / Ubuntu etc.
apt install git imagemagick libwebp6 libarchive-tools curl ffmpeg 

git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
pip3 install -r requirements.txt
BOT_TOKEN=your_bot_token python3 main.py
```

## CHANGELOG
* 4.0 ALPHA-5 (20220210)

  Bring back fake RetryAfter check since some people still have this issue.
* 4.0 ALPHA-4 (20220210)

  Support user uploaded animated(video) stickers. You can both create or add to set.
  Better support sticon(line_emoji)
  Bug fixes.
* 4.0 ALPHA-3 (20220209)

  Supports all special line stickers,
  including effect_animation and sticon(emoji)
* 4.0 ALPHA-1 (20220209)

  Supports animated line sticker import.


## Known issue
### No response?
The bot might have encountered some unhandled exception, try sending `/cancel` or report to GitHub issues.


## License
The GPL V3 License

![image](https://www.gnu.org/graphics/gplv3-with-text-136x68.png)
