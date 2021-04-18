# moe-sticker-bot
A Telegram bot doing sticker stuffs, it imports LINE sticker set to Telegram, downloads LINE and Telegram stickers, creates new sticker set. 

Telegram用萌萌貼圖BOT, 可以從LINE Store導入貼圖包到Telegram, 可以下載LINE和telegram的貼圖包, 可以創建新的貼圖包.

Telegram用萌え萌えのスタンプBOTです、LINEストアからスタンプをTelegramへインポートしたり、LINEとTelegramのスタンプをダウンロードしたり、新しいスタンプセットを作ったり


## Dependencies
### Python Dependencies
* [python-telegram-bot](https://github.com/python-telegram-bot/python-telegram-bot)
* requests
* bs4
* emoji


### System Dependencies
* ImageMagick
* libwebp
* bsdtar (libarchive-tools)
* curl

## Usage
Fill your telegram Bot API in `config.ini`

Then it's ready to fly!

`python3 main.py`

This software runs ONLY on Linux systems!!!

It probably runs on Mac, but never on Windows!! Please use WSL.

## Step by Step Usage
### Ubuntu / Debian
```
apt install imagemagick libwebp6 libarchive-tools curl 
pip3 install python-telegram-bot requests beautifulsoup4 emoji 
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
### open config.ini and fill your bot api.
python3 main.py
```

### Fedora / RHEL / CentOS
```
dnf install ImageMagick libwebp bsdtar curl 
pip3 install python-telegram-bot requests beautifulsoup4 emoji 
git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
### open config.ini and fill your bot api.
python3 main.py
```

## Known issue
__LINE animated stickers will never be supported beacause of Telegram's restrictions.__

LINE's animated stickers are in APNG bitmap format and convertable to GIFs, however,
Telegram's animated sticker only allow vector images, which is a completely different
format comparing to normal bitmap. Telegram even sets the size limit to 64KB, hence
converting stickers to Telegram animated format is inpossible.

## License
The GPL V3 License

![image](http://www.gnu.org/graphics/gplv3-127x51.png)
