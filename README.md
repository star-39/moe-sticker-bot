# [@moe_sticker_bot](https://t.me/moe_sticker_bot)
A Telegram bot doing sticker stuffs, it imports LINE and kakaotalk sticker set to Telegram, downloads LINE and Telegram stickers, creates and manages Telegram sticker set. 

Telegram用萌萌貼圖BOT, 可以匯入LINE和kakaotalk貼圖包到Telegram, 可以下載LINE和telegram的貼圖包, 可以創建和管理Telegram貼圖包.

Telegram用萌え萌えのスタンプBOTです。LINEストアからスタンプをTelegramにインポートしたり、LINEとTelegramのスタンプをダウンロードしたり、新しいTelegramステッカーセットを作ったり管理したり、色んなスタンプ関連機能があります。

<!-- ![スクリーンショット 2022-02-12 003746](https://user-images.githubusercontent.com/75669297/153621406-16a619a8-e897-4857-947b-7d41e88fddcb.png) -->
<img width="511" alt="スクリーンショット 2022-03-24 19 58 09" src="https://user-images.githubusercontent.com/75669297/159902095-fefbcbbf-1c5e-4c3e-9e55-eb28b48567e0.png">

> This project is undergoing a complete rewrite to golang due to python-telegram-bot's countless problems. No new python code will be commited. Completion date is unknown, probably before June. This python version will be re-branched after go version is published to master brach.

## Deployment
### Deploy with pre-built containers
A pre-built OCI container is available at https://github.com/users/star-39/packages/container/package/moe-sticker-bot

Simply run:
```
docker run -dt -e BOT_TOKEN=your_bot_token ghcr.io/star-39/moe-sticker-bot:latest
```

### System Dependencies
* ImageMagick
* bsdtar (libarchive-tools)
* ffmpeg
* curl
* mariadb-server (optional)

### Manual deployment
#### Linux/macOS
```
# For Fedora / RHEL / CentOS etc. (Requires RPM Fusion)
dnf install git ImageMagick libwebp bsdtar curl ffmpeg go
# For Debian / Ubuntu etc.
apt install git imagemagick libarchive-tools curl ffmpeg go
# For Arch
pacman -S install ffmpeg imagemagick curl libarchive go
# For macOS
brew install git imagemagick ffmpeg curl go

git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
go build
BOT_TOKEN=your_bot_token ./moe-sticker-bot
```
#### Windows
Please install scoop(https://scoop.sh) first, using Windows Powershell:
```
scoop install ffmpeg imagemagick python go
git clone https://github.com/star-39/moe-sticker-bot ; cd moe-sticker-bot
go build
$ENV:BOT_TOKEN=your_bot_token ; .\moe-sticker-bot
```

#### mariadb
Bot supports saving imported line stickers into a database and will notify user that a already imported set is available.

To deploy this feature. Set up mariadb-server and set the following env variables:

`DB_USER DB_PASS DB_NAME DB_ADDR`

`USE_DB=1`

## CHANGELOG
1.0 RC-1 GO(20220506)
  * Completely rewritten whole project to golang
  * Countless bug fixes.
  * You can send sticker or link without a command now.
  * Performance gained by a lot thanks to goroutine and worker pool.

5.1 RC-4 (20220423)
  * Fix duplicated sticker.
  * Fix alpha channel converting GIF.
  
5.1 RC-3 (20220416)
  * Do not use joblib due to their bugs.
  * /download_telegram_sticker now converts to standard GIFs.

5.1 RC-2 (20220326)
  * Significantly improved perf by using parallel loop.
  * Sanitize kakao id starting with -

5.1 RC-1 (20220309)
  * Support kakaotalk emoticon.
  * Add more check for telegram sticker id.

5.0 RC-12 (20220305)
  * Database now records line link and emoji settings.
  * Fix issue when line name has <> marks.
  * Fix issue adding video to static set.
  * Fix hang handling CallbackQuery.
  * 
5.0 RC-11 (20220303)
  * You can now delete a sticker set.
  * /manage_sticker_set will show sets you created.
  * Fix missing sticker during USER_STICKER.

5.0 RC-10 (20220226)
  * Performance is now significantly improved.
  * Fix issue converting Line message stickers.
  * Bypass some regional block by LINE.

5.0 RC-9 (20220223)
  * Splitted line popups to two categories, one keeping animated only.
  * Bot now has a database stroing "good" imported sticker sets.
  * Fix duplicated stickers in sticker set.

5.0 RC-8 (20220222)
  * Fix user sticker parsing.
  * Add support for MdIcoFlashAni_b

5.0 RC-7 (20220215)
  * Fix exception if user sent nothing during USER_STICKER
  * Fix a bug where /import_line_sticker may have no response.
  * Corrected file download limit.
  * Fix animated sticon
  * Fix import hang due to missing ffmpeg '-y' param.

5.0 RC-6 (20220215)
  * Fix python-telegram-bot WebHook problem.
  * Fix emoji assign.
  * Fix black background video sticker.
  * Fix "Sticker too big" when uploading video sticker.

5.0 RC-5 (20220214)
  * Allow using WebHook for better performance.
  * Code refactors.
  * Support Line name sticker.

5.0 RC-4 (20220212)
  * Improved user sticker exprerience.

5.0 RC-3 (20220212)
  * Fix a bug where creating sticker set with one sticker will cause fatal error.
  * Fix missing clean_userdata in /download_line_sticker
  * Tune VP9 params to avoid hitting 256K limit. This reduces video quality by a bit.
  
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
