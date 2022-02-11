# [@moe_sticker_bot](https://t.me/moe_sticker_bot)
A Telegram bot doing sticker stuffs, it imports LINE sticker set to Telegram, downloads LINE and Telegram stickers, creates and manages Telegram sticker set. 

Telegram用萌萌貼圖BOT, 可以從LINE Store匯入貼圖包到Telegram, 可以下載LINE和telegram的貼圖包, 可以創建和管理Telegram貼圖包.

Telegram用萌え萌えのスタンプBOTです。LINEストアからスタンプをTelegramにインポートしたり、LINEとTelegramのスタンプをダウンロードしたり、新しいTelegramステッカーセットを作ったり管理したり、色んなスタンプ関連機能があります。

![スクリーンショット 2022-02-12 003746](https://user-images.githubusercontent.com/75669297/153621406-16a619a8-e897-4857-947b-7d41e88fddcb.png)

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

### Manual deployment
#### Linux/macOS
For better performance on Linux, it's recommended to use my custom build of FFMpeg:

https://github.com/star-39/ffmpeg-nano-static
```
# For Fedora / RHEL / CentOS etc. (Requires RPM Fusion)
dnf install git ImageMagick libwebp bsdtar curl ffmpeg python3
# For Debian / Ubuntu etc.
apt install git imagemagick libwebp6 libarchive-tools curl ffmpeg python3
# For macOS
brew install git imagemagick ffmpeg curl python3

git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
pip3 install -r requirements.txt
BOT_TOKEN=your_bot_token python3 main.py
```
#### Windows
Please install scoop(https://scoop.sh) first, using Windows Powershell:
```
scoop install python3 ffmpeg imagemagick python
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
pip install -r requirements.txt
$ENV:BOT_TOKEN=your_bot_token ; python main.py
```


## CHANGELOG
5.0 RC-2 (20220211)
  * Fix media_group
  * Minor bug fixes.
  * Version 5.0 now enters freature freeze. No new feature will be added. Will have bug fixes only.

5.0 RC-1 (20220211)
  * Support Line popup sticker without sound
  * Support AVIF.
  * Many bug fixes.

5.0 ALPHA-1 (20220211)
  * Full support of animated(video) sticker. 完整支援動態貼圖. アニメーションスタンプフル対応。
  * New feature: /manage_sticker_set, now you can add, delete, move sticker in a sticker set.
  * Add support for Line full screen sticker(animated).

4.0 ALPHA-5 (20220210)
  * Bring back fake RetryAfter check since some people still having this issue.

4.0 ALPHA-4 (20220210)
  * Support user uploaded animated(video) stickers. You can both create or add to set.
  * Better support sticon(line_emoji)
  * Bug fixes.

4.0 ALPHA-3 (20220209)
  * Supports all special line stickers,
  * including effect_animation and sticon(emoji)

4.0 ALPHA-1 (20220209)
  * Supports animated line sticker import.


## Known issue
### No response?
The bot might have encountered some unhandled exception, try sending `/cancel` or report to GitHub issues.


## License
The GPL V3 License

![image](https://www.gnu.org/graphics/gplv3-with-text-136x68.png)
