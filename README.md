# [@moe_sticker_bot](https://t.me/moe_sticker_bot)

[![Go Reference](https://pkg.go.dev/badge/github.com/star-39/moe-sticker-bot.svg)](https://pkg.go.dev/github.com/star-39/moe-sticker-bot)  ![Go Report](https://goreportcard.com/badge/github.com/star-39/moe-sticker-bot)  ![CI](https://github.com/star-39/moe-sticker-bot/actions/workflows/msb_oci.yml/badge.svg)  ![CI](https://github.com/star-39/moe-sticker-bot/actions/workflows/build_binaries.yml/badge.svg) 


[<img width="500" src="https://user-images.githubusercontent.com/75669297/222379608-1359ac0f-18ed-4a25-a91e-32974994d27b.png">](https://t.me/moe_sticker_bot)

---

A Telegram bot doing sticker stuffs!

Easily import LINE/Kakaotalk stickers, use your own image or video to create Telegram sticker set or CustomEmoji and manage it.

Download stickers/GIF, also supports exporting to WhatsApp.

---

Telegram用萌萌貼圖BOT。

匯入或下載LINE和kakaotalk貼圖包到Telegram. 使用自己的圖片和影片創建Telegram貼圖包或表情貼並管理.

下載Telegram貼圖包/GIF，還可以匯出到WhatsApp。


## Features/功能
  * Import LINE or kakao stickers to Telegram without effort, you can batch or separately assign emojis.
  * Create your own sticker set or CustomEmoji with your own images or videos in any format.
  * Batch download and convert Telegram stickers or GIFs to original or common formats.
  * Export Telegram stickers to WhatsApp (requires [Msb App](https://github.com/star-39/msb_app), supports iPhone and Android).
  * Manage your sticker set interactively through WebApp: add/move/remove/edit sticker and emoji.
  * Provides a CLI app [msbimport](https://github.com/star-39/moe-sticker-bot/tree/master/pkg/msbimport) for downloading LINE/Kakaotalk stickers.

  * 輕鬆匯入LINE/kakao貼圖包到Telegram, 可以統一或分開指定emoji.
  * 輕鬆使用自己任意格式的圖片和影片來創建自己的貼圖包或表情貼.
  * 下載Telegram/LINE/kakao貼圖包和GIF, 自動變換為常用格式, 並且保留原檔.
  * 匯出Telegram的貼圖包至WhatsApp（需要安裝[Msb App](https://github.com/star-39/msb_app), 支援iPhone和Android）。
  * 互動式WebApp可以輕鬆管理自己的貼圖包: 可以新增/刪除貼圖, 移動位置或修改emoji.
  * 提供名為[msbimport](https://github.com/star-39/moe-sticker-bot/tree/master/pkg/msbimport)的終端機程式， 用於下載LINE/kakao貼圖。
  
  
## Screenshots
[![MSB](https://img.shields.io/badge/-%40moe__sticker__bot-blue?style=plastic&logo=telegram)](https://t.me/moe_sticker_bot)

<img width="487" alt="スクリーンショット 2023-02-27 午後7 29 35" src="https://user-images.githubusercontent.com/75669297/221539624-c0cc32a9-477c-425f-8e98-6566326385b4.png">


<img width="500" alt="スクリーンショット 2023-02-27 午後7 24 14" src="https://user-images.githubusercontent.com/75669297/221538927-526a878a-5d86-4b45-ab9a-d324743e3b91.png">

<img width="500" alt="スクリーンショット 2023-02-27 午後7 37 17" src="https://user-images.githubusercontent.com/75669297/221541547-4618c9ef-9be3-4d50-b7da-fdc5b25e64b8.png">


<!-- <img width="500" alt="スクリーンショット 2023-02-27 午後7 21 22" src="https://user-images.githubusercontent.com/75669297/221538953-6c69dc08-5cb1-4f07-a9ce-43bacb9f1566.png"> -->

<!-- ![スクリーンショット 2022-02-12 003746](https://user-images.githubusercontent.com/75669297/153621406-16a619a8-e897-4857-947b-7d41e88fddcb.png) 

<img width="511" alt="スクリーンショット 2022-03-24 19 58 09" src="https://user-images.githubusercontent.com/75669297/159902095-fefbcbbf-1c5e-4c3e-9e55-eb28b48567e0.png"> -->
<!--<img width="500" alt="スクリーンショット 2022-05-11 19 33 27" src="https://user-images.githubusercontent.com/75669297/167830628-1dfc9941-4b3c-408d-bf64-1ef814e3efe8.png"> <img width="500" alt="スクリーンショット 2022-05-11 19 51 46" src="https://user-images.githubusercontent.com/75669297/167833015-806b4f11-ecd9-4f10-9b9c-ecb7a20f8f97.png">-->

<!--<img width="500" alt="スクリーンショット 2022-05-11 19 58 52" src="https://user-images.githubusercontent.com/75669297/167834522-f1e988e8-bd24-44b1-a90c-f69791a9a623.png">

<img width="500" alt="スクリーンショット 2023-02-11 午前1 53 55" src="https://user-images.githubusercontent.com/75669297/218149914-65db79c0-c3f9-44ed-8043-1673eba41bc0.png">
-->
<!--
<img width="500" alt="スクリーンショット 2023-02-11 午前2 15 56" src="https://user-images.githubusercontent.com/75669297/218154599-c0af7aa5-e8ff-4f6d-9110-9b7175fbe585.png">

<img width="517" alt="スクリーンショット 2022-12-12 午後2 31 32" src="https://user-images.githubusercontent.com/75669297/206968834-d86c69d5-7e1d-4e36-9370-a66addc0c4fa.png">
-->
<!-- 
<img width="535" alt="スクリーンショット 2022-12-12 午後2 26 46" src="https://user-images.githubusercontent.com/75669297/206968863-1bb7e5cd-0c43-4573-8292-3e3e629f39bf.png"> 


<img width="562" alt="スクリーンショット 2023-02-11 午前2 21 40" src="https://user-images.githubusercontent.com/75669297/218155866-912739bc-b954-4ca2-97c1-d99e43f02a89.png">
<!--<img width="517" alt="スクリーンショット 2022-12-12 午後2 47 22" src="https://user-images.githubusercontent.com/75669297/206969650-cff19478-898a-4344-a73a-80469184053c.png">
-->

<img width="394" alt="スクリーンショット 2022-12-12 午後2 27 10" src="https://user-images.githubusercontent.com/75669297/206968889-1fe25c05-6071-422b-9e1b-549d56f5d351.png">


<img width="500" alt="スクリーンショット 2023-02-11 午前2 24 37" src="https://user-images.githubusercontent.com/75669297/218156358-0145264f-ab11-4010-bfcd-2e38621d7381.png">


<img width="300" src="https://user-images.githubusercontent.com/75669297/218153727-5fb1d3e0-3770-4dc8-a2b5-3e0ecd89a003.png"/> <img width="300" src="https://user-images.githubusercontent.com/75669297/221529085-2581bcca-fe49-46b0-8123-5614e90a838c.png"/>



## Deployment
### Deploy with pre-built containers
It is __highly recommended__ to deploy moe-sticker-bot using containers.
A pre-built OCI container is available at https://github.com/users/star-39/packages/container/package/moe-sticker-bot

Simply run:
```
docker run -dt ghcr.io/star-39/moe-sticker-bot /moe-sticker-bot --bot_token="..."
```
If you are on ARM64 machine, use `aarch64` tag.

To deploy all features - including database and webapp, you can use kubernetes or podman.

See a real world deployment example on [deployments/kubernetes_msb.yaml](https://github.com/star-39/moe-sticker-bot/blob/master/deployments/kubernetes_msb.yaml).


### System Dependencies
* ImageMagick
* bsdtar (libarchive-tools)
* ffmpeg
* curl
* gifsicle (for converting GIF)
* python3 (for following tools)
* [msb_emoji.py](https://github.com/star-39/moe-sticker-bot/tree/master/tools/msb_emoji.py) (for extracting emoji)
* [msb_kakao_decrypt.py](https://github.com/star-39/moe-sticker-bot/tree/master/tools/msb_kakao_decrypt.py) (for decrypting animated kakao)
* [msb_rlottie.py](https://github.com/star-39/moe-sticker-bot/tree/master/tools/msb_rlottie.py) (for converting TGS)
* mariadb-server (optional, for database)
* nginx (optional, for WebApp)


## Build
### Build Dependencies
 * golang v1.18+
 * nodejs v18+ (optional, for WebApp)
 * react-js v18+ (optional, for WebApp)

```bash
# For Fedora / RHEL / CentOS etc. (Requires RPM Fusion)
dnf install git ImageMagick libwebp bsdtar curl ffmpeg go gifsicle
# For Debian / Ubuntu etc.
apt install git imagemagick libarchive-tools curl ffmpeg go gifsicle
# For Arch
pacman -S install git ffmpeg imagemagick curl libarchive go gifsicle
# For macOS
brew install git imagemagick ffmpeg curl go bsdtar gifsicle
# For Windows, please install System Dependencies manually.

git clone https://github.com/star-39/moe-sticker-bot && cd moe-sticker-bot

go build -o moe-sticker-bot cmd/moe-sticker-bot/main.go 

# Install MSB dependencies(optional).
install tools/msb_emoji.py /usr/local/bin/
install tools/msb_kakao_decrypt.py /usr/local/bin/
install tools/msb_rlottie.py /usr/local/bin/
```

#### WebApp
Since 2.0 version of moe-sticker-bot, managing sticker set's order and emoji is now through Telegram's
new WebApp technology. 

See details on [web/webapp](https://github.com/star-39/moe-sticker-bot/tree/master/web/webapp3)


## CHANGELOG
v2.4.0-RC1-RC2(20240207)
  * Support Importing LINE Emoji into CustomEmoji.
  * Support creating CustomEmoji.
  * Support editing sticker emoji and title.
  
v2.3.14-2.3.15(20230228)
  * Fix missing libwebp in OCI.
  * Support TGS export.
  * Improve GIF efficiency.

v2.3.11-v2.3.13 (20230227)
  * Fix flood limit by ignoring it.
  * Fix managing video stickers.
  * Improved WhatsApp export.
  * Support region locked LINE Message sticker.
  * 修復flood limit匯入錯誤
  * 修復動態貼圖管理。
  * 改善WhatsApp匯出。
  * 支援有區域鎖的line訊息貼圖。

v2.3.8-2.3.10 (20230217)
  * Fix kakao import fatal, support more animated kakao.
  * 修復KAKAO匯入錯誤, 支援更多KAKAO動態貼圖.
  * Fix static webp might oversize(256KB)
  * Fix panic during assign: invalid sticker emojis
  * Fix some kakao animated treated as static
  * Improved kakao static sticker quality
  * Improved user experience

v2.3.6-2.3.7 (20230213)
  * Support webhook.
  * Support animated webp for user sticker.
  * Add change sticker set title "feature"
  * Fix "sticker order mismatch" when using WebApp to sort.
  * Fix error on emoji assign.
  * Fix too large animated kakao sticker.
  
v2.3.1-2.3.5 (20230209)
  * Fix i18n titles.
  * Fix flood limit by implementing channel to limit autocommit cocurrency.
  * Fix error on webapp
  * Fix import hang
  * Fix fatal error not being reported to user.
  * Fix typos
  
v2.3.0 (20230207)
  * Fix flood limit by using local api server.
  * Support webhook and local api server.
  * Huge performance gain.
  * Fix /search panic.

v2.2.1 (20230204)
  * Fix webm too big.
  * Fix id too long.
  * Tuned flood limit algorithm.

v2.2.0 (20230131)
  * Support animated kakao sticker.
  * 支援動態kakao貼圖。

v2.1.0 (20230129)
  * Support exporting sticker to WhatsApp.
  * 支援匯出貼圖到WhatsApp

2.0.1 (20230106)
  * Fix special LINE officialaccount sticker.
  * Fix `--log_level` cmdline parsing.
  * Thank you all! This project has reached 100 stars!

2.0.0 (20230105)
  * Use new WebApp from /manage command to edit sticker set with ease.
  * Send text or use /search command to search imported LINE/kakao sticker sets by all users.
  * Auto import now happens on backgroud.
  * Downloading sticker set is now lot faster.
  * Fix many LINE import issues.
  * 通過/manage指令使用新的WebApp輕鬆管理貼圖包.
  * 直接傳送文字或使用/search指令來搜尋所有用戶匯入的LINE/KAKAO貼圖包.
  * 自動匯入現在會在背景處理.
  * 下載整個貼圖包的速度現在會快許多.
  * 修復了許多LINE貼圖匯入的問題.
  
<details>
<summary>Detailed 2.0 Changelogs 詳細的2.0變更列表</summary>

2.0.0 (20230104)
  * Improve flood limit handling.
  * Auto LINE import now happens on backgroud.
  * Improve GIF download.

2.0.0 RC-7 (20221230)
  * Support /search in group chat.
  * Fix search result.
  * Fix empty sticker title.
  * Sticker download is now parallel.

2.0.0 RC-6 (20221220)
  * Fix line APNG with unexpected `tEXt` chunk.
  * Changed length of webm from 2.9s to 3s.
  * Minor improvements.

2.0.0 RC-5 (20221211)
  * Fix potential panic when onError
  * Warn user sticker set is full.
  * Fix LINE message sticker with region lock.

2.0.0 RC-4 (20221211)
  * Fix edit sticker on iOS
  * Fix error editing multiple emojis.

2.0.0 RC-3 (20221210)
  * Complies to LINE store's new UA requeirments.
  * Fix animated sticker in webapp.
  * Fixed sticker download
  * Fixed webapp image aspect ratio.

2.0.0 RC-2 (20221210)
  * Fix fatalpanic on webapp.
  * Add /search functionality.
  * Removed gin TLS support.
  * Auto database curation.

2.0.0 RC-1 (20221206)
  * WebApp support for edit stickers.
  * Code structure refactored.
  * Now accepts options from cmdline instead of env var.
  * Support parallel sticker download.
  * Fix LINE officialaccount/event/sticker
  * Fix kakao link with queries.

</details>

1.2.4 (20221111)
  * Minor improvements.
  * Fixed(almost) flood limit.
  * Fixed kakao link with queries.

1.2.2 (20220523)
  * Improved user experience.

1.2.1 (20220520)
  * Improved emoji edit.

1.2 (20220518)
  * Fix import error for LINE ID < 775 
  * Improved UX during /import.
  * Warn user if sticker set is already full.

1.1 (20220517)
  * Code refactors.
  * UX improvements.
  * Skip error on TGS to GIF due to lottie issues.

1.0 (20220513)
  * First stable release in go version.
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

<details>
<summary>Old changelogs</summary>

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
</details>

 <!--
## Technical Details
![MSB_INFO](https://user-images.githubusercontent.com/75669297/210700704-4c9b366a-c72c-42fe-919c-336b7b8024c4.svg)
-->


## Special Thanks:
[<img width=200 src="https://idcs-cb5322c0a68345bb83637843d27aa437.identity.oraclecloud.com/ui/v1/public/common/asset/defaultBranding/oracle-desktop-logo.gif">](https://www.oracle.com/cloud/) for free 4CPU AArch64 Cloud Instance.

<a href="http://t.me/StickerGroup">貼圖群 - Sticker Group Taiwan</a> for testing and reporting.

[LINE Corp](https://linecorp.com/) / [Kakao Corp](http://www.kakaocorp.com/) for cute stickers.
 
https://github.com/blluv/KakaoTalkEmoticonDownloader MIT License Copyright @blluv
 
https://github.com/laggykiller/rlottie-python GPL-2.0 license  Copyright @laggykiller

You and all the users! ☺


## License
The GPL V3 License

![image](https://www.gnu.org/graphics/gplv3-with-text-136x68.png)
