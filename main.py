import time

import telegram.error
from telegram import  Update, Bot, ReplyKeyboardMarkup, Update, ReplyKeyboardRemove
from telegram.ext import Updater, CommandHandler, CallbackQueryHandler, CallbackContext, ConversationHandler, \
    MessageHandler, Filters
from bs4 import BeautifulSoup
import emoji
import requests
import configparser
import re
import os
import subprocess
import secrets
import traceback


class GlobalConfigs:
    BOT_NAME = ""
    BOT_TOKEN = ""
    BOT_VERSION = "0.2 BETA"

# DO NOT log user's activity if you are aiming for "privacy"
# TODO: Add functionality for manipulating Telegram Stickers.
# TODO: parse_mode=MarkdownV2 sucks, too many escapes, change all to HTML
# TODO: Separate text messages to a standalone "helper" python file.
# TODO: #1

LINE_STICKER_INFO, EMOJI, ID, TITLE, MANUAL_EMOJI = range(5)

GET_TG_STICKER = range(1)

reply_kb_for_auto_markup = ReplyKeyboardMarkup([['auto']], one_time_keyboard=True)
reply_kb_for_manual_markup = ReplyKeyboardMarkup([['manual']], one_time_keyboard=True)


def start(update: Update, _: CallbackContext) -> None:
    update.message.reply_text(''
"""
Hello\! I\'m moe\_sticker\_bot doing sticker stuffs\! Please select command below:
你好\! 歡迎使用萌萌貼圖BOT, 請從下方選擇指令:
こんにちは！　萌え萌えのスタンプBOTです！下からコマンドを選択してくださいね

*/import\_line\_sticker*
`    Import sticker set from LINE Store to Telegram`
`    從LINE STORE將貼圖包導入成Telegram的貼圖包`
`    LINE STOREからスタンプをTelegramへインポート`

*/download\_line\_sticker*
`    Download sticker set from LINE Store`
`    從LINE STORE下載貼圖包`
`    LINE STOREからスタンプをダウンロード`

*/get\_animated\_line\_sticker*
`    Get GIF sticker from LINE Store animated sticker set`
`    獲取GIF版LINE STORE動態貼圖`
`    LINE STOREから動くスタンプをGIF形式で入手`

*/download\_telegram\_sticker*
`    Download Telegram sticker set.(webp png)`
`    下載Telegram的貼圖包.(webp png)`
`    Telegramのステッカーセットをダウンロード(webp png)`

*/help*
`    Get help. 幫助訊息. ヘルプ`

*/cancel*
`    Cancel interacting process.`
`    終止互動式過程.`
`    プロセスを中止する`
"""
                              , parse_mode="MarkdownV2")


def help_command(update: Update, _: CallbackContext) -> None:
    update.message.reply_text(
f"""
<code>
moe-sticker-bot by @Plow
This software comes with ABSOLUTELY NO WARRANTY!
Released under the GPL v3 License. All rights reserved.
Source code is available at: </code>
https://github.com/star-39/moe-sticker-bot
<b>
Please send /start to start using!
請傳送 /start 來開始!
始めるには　/start　を入力してください！
</b>
<b>FAQ:</b>
=>Q: The generated sticker set ID has the bot's name as suffix! 
　　　創建的貼圖包ID末尾有這個bot的名字ㄟ!
　　　出来立てのステッカーセットのIDの最後にBOTの名前がはいてる！
=>A: This is compulsory by Telegram, ID of sticker set created by a bot must has it's name as suffix.
   　這是Telegram的強制要求, BOT創建的貼圖包ID末尾必須要有BOT的名字.
   　これはTelegramからのおきてです。BOTで生成されたステッカーセットのIDの最後に必ずBOTの名前が入っています。
   
=>Q: The sticker set title is in English when <code>auto</code> is used during setting title.
   　當設定標題時使用了<code>auto</code>, 結果貼圖包的標題是英文的
   　タイトルを入力している時に<code>auto</code>を入力すると、タイトルは英語になっちゃうん
=>A: Line sometimes has sticker set in multiple languages, you should paste LINE store link with language suffix.
   　有的LINE貼圖包有多種語言, 請確認輸入LINE商店連結的時候末尾有指定語言.
   　LINEのスタンプは時々多言語対応です、リンクを入力するとき、最後に言語コードがあるかどうかを確認してください。

=>Q: Donation? 捐贈? 寄付？
=>A: I created this BOT solely for fun and public interest, and hosting it in a small server.
   　If you want to donate, please support original Sticker creators by purchasing in LINE store.
   　If you are a power user, why not host your own one!? It's super easy! Please check Github repo.
  　 我創建這個BOT單純因為興趣和出於公眾利益, 目前host在一個小小伺服器上.
   　如果您想課金感謝, 請在LINE商店購買原版貼圖包支持原本的各位創作者.
   　如果你深諳資訊科技, 那麼完全可以自己host這個BOT! 一點也不難! 請見Github repo.
   　自分はただ面白いと公衆利益のためこのBOTを作った。今現在、小っちゃいサーバーに生きています。
   　もし寄付したいなら、ぜひスタンプのクリエーターを応援して、LINEストアからスタンプを購入してください、
   　もしあなたがITの達人だったら、自分でこのＢＯＴをホストするのはどうでしょう？すごく簡単ですよ！詳細はGithub repoに。
  
=>Q: Can I get an animated sticker set instead of GIF.
     可以轉換成動態貼圖包而不是GIF嗎?
     GIFじゃなく動くステッカーセットが欲しいんですけど？
=>A: It is technically impossible due to Telegram's restrictions.
　　  因為Telegram的限制, 這個方法技術上不可行
     Telegram側の制限により技術的に無理です。
     
=>Q: No response? 沒有反應? 応答なし？
=>A: The bot might triggered Telegram's flood control or encountered an error, please check GitHub issue:
     BOT可能觸發了Telegram的消息數量限制或遇到了問題, 請檢視GitHub issue網頁:
     https://github.com/star-39/moe-sticker-bot/issues
<code>
========================
PRIVACY NOTICE:
    This software(bot) does not collect or save any kind of your personal information.
    This software complies with Hong Kong Legislation Cap. 486 Personal Data (Privacy) Ordinance.
========================
BOT NAME: {GlobalConfigs.BOT_NAME} 
VERSION: {GlobalConfigs.BOT_VERSION} 
</code>
"""
        , parse_mode="HTML")


def do_auto_import_line_sticker(update, _):
    notify_import_starting(update, _)

    img_files_path = prepare_sticker_files(_, want_animated=False)
    # Create a new sticker set using the first image.
    try:
        _.bot.create_new_sticker_set(user_id=update.message.from_user.id,
                                     name=_.user_data['telegram_sticker_id'],
                                     title=_.user_data['telegram_sticker_title'],
                                     emojis=_.user_data['telegram_sticker_emoji'],
                                     png_sticker=open(img_files_path[0], 'rb'))
    except Exception as e:
        update.message.reply_text("Failed to create new sticker set!\n" + str(e))
        return ConversationHandler.END

    message_progress = report_progress(None, 1, len(img_files_path), update=update)
    for index, img_file_path in enumerate(img_files_path):
        # Skip the first file.
        if index != 0:
            try:
                time.sleep(1)
                _.bot.add_sticker_to_set(user_id=update.message.from_user.id,
                                         name=_.user_data['telegram_sticker_id'],
                                         emojis=_.user_data['telegram_sticker_emoji'],
                                         png_sticker=open(img_file_path, 'rb'))
                report_progress(message_progress, index + 1, len(img_files_path))
            except telegram.error.RetryAfter as e:

                update.message.reply_text("Error when adding sticker to set!\n" + str(e))

    notify_sticker_done(update, _)


def notify_import_starting(update, _):
    update.message.reply_text("Now starting, please wait...\n"
                              "正在開始, 請稍後...\n"
                              "作業がまもなく開始します、少々お時間を...\n\n"
                              "<code>"
                              f"LINE TYPE: {_.user_data['line_sticker_type']}\n"
                              f"LINE ID: {_.user_data['line_sticker_id']}\n"
                              f"TG ID: {_.user_data['telegram_sticker_id']}\n"
                              f"TG TITLE: {_.user_data['telegram_sticker_title']}\n"
                              "</code>",
                              parse_mode="HTML", reply_markup=ReplyKeyboardRemove())


def prepare_sticker_files(_, want_animated):
    os.makedirs("line_sticker", exist_ok=True)
    directory_path = "line_sticker/" + _.user_data['line_sticker_id'] + "/"
    zip_file_path = "line_sticker/" + _.user_data['line_sticker_id'] + ".zip"
    subprocess.run("curl -Lo " + zip_file_path + " " + _.user_data['line_sticker_download_url'], shell=True)
    os.makedirs("line_sticker/" + _.user_data['line_sticker_id'], exist_ok=True)
    subprocess.run(f"rm -r {directory_path}*", shell=True)
    subprocess.run("bsdtar -xf " + zip_file_path + " -C " + directory_path, shell=True)
    if not want_animated:
        # Remove garbage
        subprocess.run(f"rm {directory_path}*key* {directory_path}tab* {directory_path}productInfo.meta", shell=True)
        # Make a webp version
        subprocess.run(f"mogrify -format webp {directory_path}*.png", shell=True)
        # Resize to fulfill telegram's requirement, AR is automatically retained
        # Lanczos resizing produces much sharper image.
        subprocess.run(f"mogrify -background none -filter Lanczos -resize 512x512 {directory_path}*.png", shell=True)
        return sorted([directory_path + f for f in os.listdir(directory_path) if
                       os.path.isfile(os.path.join(directory_path, f)) and f.endswith("png")])
    else:
        directory_path += "animation@2x/"
        # Magic!!
        # LINE's apng has fps of 9, however ffmpeg defaults to 25
        subprocess.run(f'find {directory_path}*.png -type f -print0 | '
                       'xargs -I{} -0 ffmpeg -hide_banner -i {} '
                       '-lavfi "color=white[c];[c][0]scale2ref[cs][0s];[cs][0s]overlay=shortest=1,setsar=1:1" '
                       '-c:v libx264 -r 9 -crf 26 -y {}.mp4', shell=True)
        return sorted([directory_path + f for f in os.listdir(directory_path) if
                       os.path.isfile(os.path.join(directory_path, f)) and f.endswith(".mp4")])


def get_sticker_thumbnails_path(_):
    directory_path = "line_sticker/" + _.user_data['line_sticker_id'] + "/"
    thumb_files_path = sorted([directory_path + f for f in os.listdir(directory_path) if
                               os.path.isfile(os.path.join(directory_path, f)) and f.endswith(".webp")])
    return thumb_files_path


def report_progress(message_progress, current, total, update=None):
    progress_1 =  '[=>                  ]'
    progress_25 = '[====>               ]'
    progress_50 = '[=========>          ]'
    progress_75 = '[==============>     ]'
    progress_100 ='[====================]'

    if update is not None:
        return update.message.reply_text("<b>Current Status</b>\n"
                                   "<code>" + progress_1 + "</code>\n"
                                   "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                   parse_mode="HTML")
    if current == int(0.25 * total):
        message_progress.edit_text("<b>Current Status</b>\n"
                                   "<code>" + progress_25 + "</code>\n"
                                   "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                   parse_mode="HTML")
    if current == int(0.5 * total):
        message_progress.edit_text("<b>Current Status</b>\n"
                                   "<code>" + progress_50 + "</code>\n"
                                   "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                   parse_mode="HTML")
    if current == int(0.75 * total):
        message_progress.edit_text("<b>Current Status</b>\n"
                                   "<code>" + progress_75 + "</code>\n"
                                   "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                   parse_mode="HTML")
    if current == total:
        message_progress.edit_text("<b>Current Status</b>\n"
                                   "<code>" + progress_100 + "</code>\n"
                                   "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                   parse_mode="HTML")


def do_download_line_sticker(update, _):
    update.message.reply_text(_.user_data['line_sticker_download_url'])


# MANUAL_EMOJI
def manual_add_emoji(update: Update, _: CallbackContext) -> int:
    # Verify emoji.
    em = ''.join(e for e in re.findall(emoji.get_emoji_regexp(), update.message.text))
    if _.user_data['manual_emoji_index'] != -1 and em == '':
        update.message.reply_text("Please send emoji! Try again")
        return MANUAL_EMOJI

    # Initialise
    if _.user_data['manual_emoji_index'] == -1:
        notify_import_starting(update, _)
        _.user_data['img_files_path'] = prepare_sticker_files(_, want_animated=False)
        _.user_data['img_thumbnails_path'] = get_sticker_thumbnails_path(_)
        # This is the FIRST sticker.
        notify_next(update, _)

    # First sticker to create new set.
    elif _.user_data['manual_emoji_index'] == 0:
        try:
            _.bot.create_new_sticker_set(user_id=update.message.from_user.id,
                                         name=_.user_data['telegram_sticker_id'],
                                         title=_.user_data['telegram_sticker_title'],
                                         emojis=em,
                                         png_sticker=open(_.user_data['img_files_path'][0], 'rb'))
        except Exception as e:
            update.message.reply_text("Error creating! Please send the same emoji again.\n" + str(e))
            return MANUAL_EMOJI
        # This is the next sticker (in this case, the SECOND one).
        notify_next(update, _)

    else:
        try:
            _.bot.add_sticker_to_set(user_id=update.message.from_user.id,
                                     name=_.user_data['telegram_sticker_id'],
                                     emojis=em,
                                     png_sticker=open(
                                         _.user_data['img_files_path'][ _.user_data['manual_emoji_index'] ], 'rb'
                                         )
                                     )
        except Exception as e:
            update.message.reply_text("Error assigning this one! Please send the same emoji again.\n" + str(e))
            return MANUAL_EMOJI

        if _.user_data['manual_emoji_index'] == len(_.user_data['img_files_path']) - 1:
            notify_sticker_done(update, _)
            return ConversationHandler.END

        notify_next(update, _)

    _.user_data['manual_emoji_index'] += 1
    return MANUAL_EMOJI


def notify_next(update, _):
    time.sleep(1)
    _.bot.send_photo(chat_id=update.effective_chat.id,
                     caption="Please send an emoji representing this sticker\n"
                             "請輸入一個代表這個貼圖的emoji\n"
                             "このスタンプにふさわしい絵文字を一つ入力してください\n" +
                             f"{_.user_data['manual_emoji_index'] + 2} of {len(_.user_data['img_files_path'])}",
                     photo=open(_.user_data['img_thumbnails_path'][_.user_data['manual_emoji_index'] + 1], 'rb'))


def notify_sticker_done(update, _):
    time.sleep(1)
    update.message.reply_text("The sticker set has been successfully created!\n"
                              "貼圖包已經成功創建!\n"
                              "ステッカーセットの作成が成功しました！\n\n"
                              "https://t.me/addstickers/" + _.user_data['telegram_sticker_id'])
    if _.user_data['line_sticker_type'] == "sticker_animated":
        update.message.reply_text("It seems the sticker set you imported also has a animated version\n"
                                  "Please use /get_animated_line_sticker to have their GIF version\n"
                                  "您導入的貼圖包還有動態貼圖版本\n"
                                  "請使用 /get_animated_line_sticker 獲取GIF版動態貼圖\n"
                                  "このスタンプの動くバージョンもございます。\n"
                                  "/get_animated_line_sticker を使ってGIF版のスタンプを入手できます")

# TITLE
# This is the final step, if user wants to assign each sticker a different emoji, return to MANUAL_EMOJI,
# otherwise, END conversation.
def parse_title(update: Update, _: CallbackContext) -> int:
    if update.message.text.strip().lower() == "auto":
        _.user_data['telegram_sticker_title'] = BeautifulSoup(_.user_data['line_store_webpage_text'], 'html.parser'
                                                              ).find("title").get_text().split('|')[0].strip()
        update.message.reply_text("The title will be automatically set to:\n"
                                  "標題將會自動設定為: \n"
                                  "タイトルは自動的にこのように設定します: \n\n"
                                  "<code>" + _.user_data['telegram_sticker_title'] + "</code>", parse_mode="HTML")
    else:
        _.user_data['telegram_sticker_title'] = update.message.text.strip()

    if _.user_data['manual_emoji'] is True:
        _.user_data['manual_emoji_index'] = -1
        # Note that MANUAL_EMOJI will be called AFTER user send a emoji!!
        # Hence we need to call it in advance to "initialise" the process.
        manual_add_emoji(update, _)
        return MANUAL_EMOJI

    do_auto_import_line_sticker(update, _)
    return ConversationHandler.END


# ID
def parse_id(update: Update, _: CallbackContext) -> int:
    if update.message.text.strip().lower() == "auto":
        _.user_data['telegram_sticker_id'] = f"line_{_.user_data['line_sticker_type']}_" \
                                             f"{_.user_data['line_sticker_id']}_" \
                                             f"{secrets.token_hex(nbytes=3)}_by_{GlobalConfigs.BOT_NAME}"
        update.message.reply_text("The ID will be automatically set to:\n"
                                  "ID將會自動設定為: \n"
                                  "IDは自動的にこのように設定します: \n\n"
                                  "<code>" + _.user_data['telegram_sticker_id'] + "</code>", parse_mode="HTML")
    else:
        _.user_data['telegram_sticker_id'] = update.message.text.strip() + "_" + secrets.token_hex(nbytes=3) + \
                                             "_by_" + GlobalConfigs.BOT_NAME
        if not re.match(r'^[a-zA-Z0-9_]+$', _.user_data['telegram_sticker_id']):
            update.message.reply_text(
                "Error: Wrong format!\n"
                "Can contain only english letters, digits and underscores.\n"
                "Must begin with a letter, can't contain consecutive underscores.")
            return ID

    update.message.reply_text(
        "Please set a title for this sticker set\. Send `auto` to automatically set original title from LINE Store\n"
        "請設定貼圖包的標題, 也就是名字\. 輸入`auto`可以自動設為LINE Store中原版的標題\n"
        "スタンプのタイトルを入力してください。`auto`を入力すると、LINE STORE上に表記されているのタイトルが自動的に設定されます。",
        reply_markup=reply_kb_for_auto_markup,
        parse_mode="MarkdownV2")
    return TITLE


# EMOJI
def parse_emoji(update: Update, _: CallbackContext) -> int:
    if update.message.text.strip().lower() == "manual":
        _.user_data['manual_emoji'] = True
    else:
        em = ''.join(e for e in re.findall(emoji.get_emoji_regexp(), update.message.text))
        if em == '':
            update.message.reply_text("Please send emoji! Try again")
            return EMOJI
        _.user_data['telegram_sticker_emoji'] = em

    update.message.reply_text("Please enter an unique ID for this sticker set. Must contain alphanum and _ mark only.\n"
                              "請輸入一個用於識別此貼圖包的ID, 只可以由英文數字和 _ 記號組成\.\n"
                              "スタンプにIDを付けてください。内容は英字と数字と _ 記号のみです。\n\n"
                              "<code>eg. gochiusa_chino_stamp_1</code>\n"
                              "-----------------------------------------------------------\n"
                              "Send <code>auto</code> to automatically generate ID\n"
                              "傳送<code>auto</code>來自動生成ID\n"
                              "<code>auto</code>を入力すると、IDが自動的生成されます",
                              reply_markup=reply_kb_for_auto_markup,
                              parse_mode="HTML")
    return ID


# TODO: This function does too much, refactor needed
# LINE_STICKER_INFO
def parse_line_url(update: Update, _: CallbackContext) -> int:
    message = update.message.text.strip()
    if not message.isdigit():
        try:
            _.user_data['line_store_webpage_text'] = requests.get(message).text
            _.user_data['line_sticker_type'], _.user_data['line_sticker_id'] = \
                get_line_sticker_detail(message, _.user_data['line_store_webpage_text'])
            _.user_data['line_sticker_url'] = message
        except:
            update.message.reply_text('URL parse error! Make sure you sent a LINE Store URL !! Try again please.\n'
                                      'URL解析錯誤!! 請確認輸入的是正確的LINE商店URL連結. 請重試.\n'
                                      'URL解析エラー！もう一度、正しいLINEスタンプストアのリンクを入力してください')
            return LINE_STICKER_INFO
    else:
        _.user_data['line_sticker_id'] = message
        _.user_data['line_sticker_url'] = compose_line_store_url(_.user_data['line_sticker_type'],
                                                                 _.user_data['line_sticker_id'])

    _.user_data['line_sticker_download_url'] = compose_line_download_url(_.user_data['line_sticker_type'],
                                                                         _.user_data['line_sticker_id'])

    if str(_.user_data['in_command']).startswith("/import_line_sticker"):
        ask_emoji(update)
        return EMOJI
    elif str(_.user_data['in_command']).startswith("/download_line_sticker"):
        do_download_line_sticker(update, _)
        return ConversationHandler.END
    elif str(_.user_data['in_command']).startswith("/get_animated_line_sticker"):
        do_get_animated_line_sticker(update, _)
        return ConversationHandler.END
    else:
        pass


def do_get_animated_line_sticker(update, _):
    if _.user_data['line_sticker_type'] != "sticker_animated":
        update.message.reply_text("Sorry! This LINE Sticker set is NOT animated! Please check again.\n"
                                  "抱歉! 這個LINE貼圖包沒有動態版本! 請再檢查一次.\n"
                                  "このスタンプの動くバージョンはございません。もう一度ご確認してください。")
        return ConversationHandler.END
    notify_import_starting(update, _)
    for gif_file in prepare_sticker_files(_, want_animated=True):
        time.sleep(1)
        _.bot.send_animation(chat_id=update.effective_chat.id,
                             animation=open(gif_file, 'rb'))


def ask_emoji(update):
    update.message.reply_text("Please enter a emoji representing this sticker set\n"
                              "請輸入一個用於表示這個貼圖包的emoji\n"
                              "このスタンプセットにふさわしい絵文字を入力してください\n"
                              "eg\. ☕ \n"
                              "`---------------------------------------------------`\n"
                              "This operation assigns the same emoji for every stickers\n"
                              "If you want to manually assign different emoji for each sticker, send `manual`\n"
                              "這個操作將會為貼圖包內每一個貼圖都設定相同的emoji,\n"
                              "如果您想要手動為每個貼圖設定不同的emoji, 請傳送 ` manual`\n"
                              "このステップでは、すべてのステッカーに同じ絵文字を付けます。\n"
                              "一つずつ絵文字を付けたいなら、`manual`を送信してください。",
                              reply_markup=reply_kb_for_manual_markup,
                              parse_mode="MarkdownV2")


def get_line_sticker_detail(link, webpage_text):
    split_line_url = link.split('/')
    if split_line_url[split_line_url.index("store.line.me") + 1] == "stickershop":
        # First one matches AnimatedSticker with NO sound and second one with sound.
        if '<span class="MdIcoPlay_b" data-test="animation-sticker-icon">Animation only icon</span>' in webpage_text or \
                '<span class="MdIcoAni_b" data-test="animation-sound-sticker-icon">Animation & Sound icon</span>' in webpage_text:
            t = "sticker_animated"
        else:
            t = "sticker"
    elif split_line_url[split_line_url.index("store.line.me") + 1] == "emojishop":
        t = "emoji"
    else:
        t = ""

    i = split_line_url[split_line_url.index("product") + 1]
    return t, i


def compose_line_store_url(type, id):
    if type == "sticker" or type == "sticker_animated":
        return "https://store.line.me/stickershop/product/" + id
    elif type == "emoji":
        return "https://store.line.me/emojishop/product/" + id


def compose_line_download_url(type, id):
    if type == "sticker":
        return "https://stickershop.line-scdn.net/stickershop/v1/product/" + id + "/iphone/stickers@2x.zip"
    elif type == "sticker_animated":
        return "https://stickershop.line-scdn.net/stickershop/v1/product/" + id + "/iphone/stickerpack@2x.zip.zip"
    elif type == "emoji":
        return "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + id + "/iphone/package.zip"


def command_import_line_sticker(update: Update, _: CallbackContext):
    initialize_user_data(update, _)
    ask_line_store_link(update)
    return LINE_STICKER_INFO


def command_download_line_sticker(update: Update, _: CallbackContext):
    initialize_user_data(update, _)
    ask_line_store_link(update)
    return LINE_STICKER_INFO


def command_get_animated_line_sticker(update: Update, _: CallbackContext):
    initialize_user_data(update, _)
    _.user_data['line_sticker_type'] = "sticker_animated"
    ask_line_store_link(update)
    return LINE_STICKER_INFO


def initialize_user_data(update, _):
    _.user_data['in_command'] = update.message.text
    _.user_data['manual_emoji'] = False
    _.user_data['line_sticker_url'] = ""
    _.user_data['line_store_webpage_text'] = ""
    _.user_data['line_sticker_download_url'] = ""
    _.user_data['line_sticker_type'] = "sticker"
    _.user_data['line_sticker_id'] = ""
    _.user_data['telegram_sticker_emoji'] = ""
    _.user_data['telegram_sticker_id'] = ""
    _.user_data['telegram_sticker_title'] = ""


def parse_tg_sticker(update: Update, _: CallbackContext) -> int:
    sticker_set = _.bot.get_sticker_set(name=update.message.sticker.set_name)
    update.message.reply_text("This might take some time, please wait...\n"
                              "此項作業可能需時較長, 請稍後...\n"""
                              "少々お待ちください...\n"
                              "<code>\n"
                              f"Name: {sticker_set.name}\n"
                              f"Title: {sticker_set.title}\n"
                              f"Amount: {str(len(sticker_set.stickers))}\n"
                              "</code>",
                              parse_mode="HTML")
    save_path = "tg_sticker/" + sticker_set.name + "/"
    os.makedirs(save_path, exist_ok=True)
    subprocess.run("rm " + save_path + "*", shell=True)
    for index, sticker in enumerate(sticker_set.stickers):
        try:
            _.bot.get_file(sticker.file_id).download(save_path + sticker.set_name + "_" + str(index) + "_" +
                                                      emoji.demojize(sticker.emoji)[1:-1] + ".webp")
        except Exception as e:
            print(str(e))
    subprocess.run("mogrify -format png " + save_path + "*.webp", shell=True)
    subprocess.run("bsdtar -acvf " + save_path + sticker_set.name + "_webp.zip " + save_path + "*.webp", shell=True)
    subprocess.run("bsdtar -acvf " + save_path + sticker_set.name + "_png.zip " + save_path + "*.png", shell=True)
    try:
        _.bot.send_document(chat_id=update.effective_chat.id,
                            document=open(save_path + sticker_set.name + "_webp.zip", 'rb'))
        time.sleep(2)
        _.bot.send_document(chat_id=update.effective_chat.id,
                            document=open(save_path + sticker_set.name + "_png.zip", 'rb'))
    except Exception as e:
        print(str(e))

    return ConversationHandler.END


def command_download_telegram_sticker(update: Update, _: CallbackContext):
    update.message.reply_text("Please send a sticker.\n"
                              "請傳送一張Telegram貼圖.\n"
                              "ステッカーを一つ送信してください。")
    return GET_TG_STICKER


def ask_line_store_link(update):
    update.message.reply_text("Please enter LINE store URL or sticker ID\n"
                              "請輸入貼圖包的LINE STORE連結或貼圖包的ID\n"
                              "スタンプのLINE STOREリンクを入力してください\n\n"
                              "`eg. https://store.line.me/stickershop/product/9961437/ja`", parse_mode="MarkdownV2")


def command_cancel(update: Update, _: CallbackContext) -> int:
    update.message.reply_text("SESSION END.")
    start(update, _)
    return ConversationHandler.END


def command_test(update: Update, _: CallbackContext):
    update.message.reply_text("A test message.")


def reject_text(update: Update, _: CallbackContext):
    update.message.reply_text("Please do not just send text! Use /start to see available commands!\n"
                              "請不要直接傳送文字! 請傳送 /start 來看看可用的指令\n"
                              "テキストを直接入力しないでください。/start を送信してコマンドで始めましょう")

# def error_handler(update: Update, _: CallbackContext):


def main() -> None:
    # Load configs
    config = configparser.ConfigParser()
    config.read('config.ini')
    GlobalConfigs.BOT_TOKEN = config['TELEGRAM']['BOT_TOKEN']
    GlobalConfigs.BOT_NAME = Bot(GlobalConfigs.BOT_TOKEN).get_me().username

    updater = Updater(GlobalConfigs.BOT_TOKEN)
    dispatcher = updater.dispatcher

    # Each conversation is time consuming, enable run_async
    conv_import_line_sticker = ConversationHandler(
        entry_points=[CommandHandler('import_line_sticker', command_import_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji)],
            ID: [MessageHandler(Filters.text & ~Filters.command, parse_id)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title)],
            MANUAL_EMOJI : [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel)],
        run_async=True
    )
    conv_get_animated_line_sticker = ConversationHandler(
        entry_points=[CommandHandler('get_animated_line_sticker', command_get_animated_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel)],
        run_async=True
    )
    conv_download_line_sticker = ConversationHandler(
        entry_points=[CommandHandler('download_line_sticker', command_download_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel)],
        run_async=True
    )
    conv_download_telegram_sticker = ConversationHandler(
        entry_points=[CommandHandler('download_telegram_sticker', command_download_telegram_sticker)],
        states={
            GET_TG_STICKER: [MessageHandler(Filters.sticker, parse_tg_sticker)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel)],
        run_async=True
    )
    # 派遣します！
    dispatcher.add_handler(conv_import_line_sticker)
    dispatcher.add_handler(conv_get_animated_line_sticker)
    dispatcher.add_handler(conv_download_line_sticker)
    dispatcher.add_handler(conv_download_telegram_sticker)
    dispatcher.add_handler(CommandHandler('start', start))
    dispatcher.add_handler(CommandHandler('help', help_command))
    dispatcher.add_handler(CommandHandler('test', command_test))
    dispatcher.add_handler(MessageHandler(Filters.text & ~Filters.command, reject_text))
    # dispatcher.add_error_handler(error_handler)


    updater.start_polling()
    updater.idle()


if __name__ == '__main__':
    main()
