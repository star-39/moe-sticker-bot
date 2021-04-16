# moe-sticker-bot
A bot doing sticker stuffs

萌萌貼圖BOT

萌え萌えのスタンプBOTです


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

## Known issue
__LINE animated stickers will never be supported beacause of Telegram's restrictions.__

LINE's animated stickers are in APNG bitmap format and convertable to GIFs, however,
Telegram's animated sticker only allow vector images, which is a completely different
format comparing to normal bitmap. Telegram even sets the size limit to 64KB, hence
converting stickers to Telegram animated format is inpossible.

## License
The GPL V3 License

![image](http://www.gnu.org/graphics/gplv3-127x51.png)
