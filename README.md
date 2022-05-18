# [@moe_sticker_bot](https://t.me/moe_sticker_bot)
A Telegram bot doing sticker stuffs, it imports LINE and kakaotalk sticker set to Telegram, downloads stickers, creates and manages Telegram sticker set. 

Telegram用萌萌貼圖BOT, 可以匯入LINE和kakaotalk貼圖包到Telegram, 可以下載貼圖包, 可以創建和管理Telegram貼圖包.

Telegram用萌え萌えのスタンプBOTです。LINEストアからスタンプをTelegramにインポートしたり、スタンプをダウンロードしたり、新しいTelegramステッカーセットを作ったり管理したり、色んなスタンプ関連機能があります。

## Features
  * Import LINE or kakao stickers to Telegram without effort, you can batch or separately assign emojis.
  * Batch download and convert Telegram stickers to original or common formats.
  * Full support of video stickers.
  * Create your own sticker set with your own images easily.
  * Manage your sticker set: add or remove sticker, change order, edit emoji.
  * Top class performance with simultaneous execution to save your time.

  * 輕鬆匯入LINE/kakao貼圖包到Telegram, 可以統一或分開指定emoji.
  * 下載Telegram/LINE/kakao貼圖包, 自動變換為常用格式, 並且保留原檔.
  * 完整支援動態貼圖.
  * 輕鬆使用自己任意格式的圖片,短片來創建自己的貼圖包.
  * 可以管理自己的貼圖包: 可以新增/刪除貼圖, 移動位置或修改emoji.
  * 擁有超高處理速度, 節省您的時間. 

## Screenshots
[@moe_sticker_bot](https://t.me/moe_sticker_bot)
<!-- ![スクリーンショット 2022-02-12 003746](https://user-images.githubusercontent.com/75669297/153621406-16a619a8-e897-4857-947b-7d41e88fddcb.png) 
<img width="511" alt="スクリーンショット 2022-03-24 19 58 09" src="https://user-images.githubusercontent.com/75669297/159902095-fefbcbbf-1c5e-4c3e-9e55-eb28b48567e0.png"> -->

<!--<img width="500" alt="スクリーンショット 2022-05-11 19 33 27" src="https://user-images.githubusercontent.com/75669297/167830628-1dfc9941-4b3c-408d-bf64-1ef814e3efe8.png"> <img width="500" alt="スクリーンショット 2022-05-11 19 51 46" src="https://user-images.githubusercontent.com/75669297/167833015-806b4f11-ecd9-4f10-9b9c-ecb7a20f8f97.png">-->

<img width="500" alt="スクリーンショット 2022-05-11 19 58 52" src="https://user-images.githubusercontent.com/75669297/167834522-f1e988e8-bd24-44b1-a90c-f69791a9a623.png">

<img width="500" alt="スクリーンショット 2022-05-11 19 59 46" src="https://user-images.githubusercontent.com/75669297/167834725-425a6b32-ba60-4201-b27e-33588ceff1df.png">

<img width="500" alt="スクリーンショット 2022-05-11 20 00 16" src="https://user-images.githubusercontent.com/75669297/167834589-bbc80f8b-c6bc-43bb-aa33-44bbc800e5eb.png"> 


## Deployment
### Deploy with pre-built containers
A pre-built OCI container is available at https://github.com/users/star-39/packages/container/package/moe-sticker-bot

Simply run:
```
docker run -dt -e BOT_TOKEN=your_bot_token ghcr.io/star-39/moe-sticker-bot:latest
```
If you are on ARM64(AArch64) arch, use `aarch64` tag:
```
docker run -dt -e BOT_TOKEN=your_bot_token ghcr.io/star-39/moe-sticker-bot:aarch64
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
pacman -S install git ffmpeg imagemagick curl libarchive go
# For macOS
brew install git imagemagick ffmpeg curl go

git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot
go build
BOT_TOKEN=your_bot_token ./moe-sticker-bot
```
#### Windows
Please install scoop(https://scoop.sh) first, using Windows Powershell:
```
scoop install ffmpeg imagemagick go
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
1.2 (20220518)
  * Fix import error for LINE ID < 775 
  * Improved UX during /import.
  * Warn user if sticker set is already full.

1.1 (20220517)
  * Code refactors.
  * UX improvements.
  * Skip error on TGS to GIF due to lottie issues.

1.0 (20220513)
  * First stable release in go version.(Python one never got stablized btw)
  * Added support for downloading TGS and convert to GIF.
  * Backing database for @moe_sticker_bot has gone complete sanitization.

1.0 RC-9(20220512)
  * Add an administrative command to _sanitize_ database, which purges duplicated stickers.
  * Add an advanced command /register, to register your sticker to database.
  * Minor bug fixes.
  * This is the REAL final RC release, next one is stable!

1.0 RC-8 GO(20220512)
  * Fix rand number in ID.
  * Major code refactor.
  * Downlaod sticker now happens on background.
  * Better documentation.
  * This release should be the final RC... hopefully.

1.0 RC-7 GO(20220511)
  * You can specify custom ID when /create.
  * Changed import ID naming scheme for cleaner look.
  * Die immediately if FloodLimit exceeds 4.
  * If everything looks good, this should be the last RC for 1.0

1.0 RC-6 GO(20220511)
  * New feature: Change sticker order.
  * New feature: Edit sticker emoji.
  * New import support: kakaotalk emoticon.
  * Fix possible panic when editMessage.
  * We are closing to a stable release! RC should not exceed 8.

1.0 RC-5 GO(20220510)
  * New feature: download raw line stickers to zip.
  * FatalError now prints stack trace.
  * zh-Hant is now default in auto LINE title.
  * Quality of video sticker should improve by a bit.
  * Fix possible slice out of range panic.
  * If user experience FloodLimit over 3 times, terminate process.s

1.0 RC-4 GO(20220509)
  * Use my custom fork of telebot
  * User sent sticker now supports any file.

1.0 RC-3 GO(20220509)
  * Split large zip to under 50MB chunks.
  * Split long message to chunks.
  * GIF downloaded is now in original 512px resolution.
  * You can press "No" now when sent link or sticker.
  * If error is not HTTP400, bot will retry for 3 times.
  * Other minor improvements.

1.0 RC-2 GO(20220508)
  * Fix SEGV when user requested /quit
  * Ignore FloodLimit by default since API will do retry at TDLib level.
  * Fix emoji in database.
  * Fix video sticker when /manage.
  * Support line message sticker and emoticon(emoji).

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


## Special Thanks:
[<img width=200 src="https://idcs-cb5322c0a68345bb83637843d27aa437.identity.oraclecloud.com/ui/v1/public/common/asset/defaultBranding/oracle-desktop-logo.gif">](https://www.oracle.com/cloud/) for free 4CPU AArch64 Cloud Instance.

<a href="http://t.me/StickerGroup">貼圖群 - Sticker Group Taiwan</a> for testing and reporting.

[LINE Corp](https://linecorp.com/) / [Kakao Corp](http://www.kakaocorp.com/) for cute stickers.

You and all the users! ☺


## License
The GPL V3 License

![image](https://www.gnu.org/graphics/gplv3-with-text-136x68.png)
