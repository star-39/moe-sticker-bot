import json
import time
import logging
from urllib.parse import urlparse
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
import argparse
import shlex

from telegram.utils.types import ConversationDict


class GlobalConfigs:
    BOT_NAME = ""
    BOT_TOKEN = ""
    BOT_VERSION = "2.0 ALPHA-1"


# logging.basicConfig(level=logging.INFO,
#                     format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
# logger = logging.getLogger(ctx_name_ctx)


LINE_STICKER_INFO, EMOJI, TITLE, MANUAL_EMOJI = range(4)

GET_TG_STICKER = range(1)

reply_kb_for_auto_markup = ReplyKeyboardMarkup([['auto']], one_time_keyboard=True)
reply_kb_for_manual_markup = ReplyKeyboardMarkup([['manual']], one_time_keyboard=True)


def start(update: Update, ctx: CallbackContext) -> None:
    update.message.reply_text(
"""
Hello! I'm moe_sticker_bot doing sticker stuffs! Please select command below:
你好! 歡迎使用萌萌貼圖BOT, 請從下方選擇指令:
こんにちは！　萌え萌えのスタンプBOTです！下からコマンドを選択してくださいね

<b>/import_line_sticker</b><code>
    Import sticker set from LINE Store to Telegram
    從LINE STORE將貼圖包匯入至Telegram
    LINE STOREからスタンプをTelegramにインポート
</code>
<b>/download_line_sticker</b><code>
    Download sticker set from LINE Store
    從LINE STORE下載貼圖包
    LINE STOREからスタンプをダウンロード
</code>
<b>/get_animated_line_sticker</b><code>
    Get GIF sticker from LINE Store animated sticker set
    獲取GIF版LINE STORE動態貼圖
    LINE STOREから動くスタンプをGIF形式で入手
</code>
<b>/download_telegram_sticker</b><code>
    Download Telegram sticker set.(webp png)
    下載Telegram的貼圖包.(webp png)
    Telegramのステッカーセットをダウンロード(webp png)
</code>
<b>/help</b><code>
    Get help. 幫助訊息. ヘルプ
</code>
<b>/cancel</b><code>
    Cancel conversation. 中斷指令. キャンセル 
</code>
"""
        , parse_mode="HTML")


def help_command(update: Update, ctx: CallbackContext) -> None:
    update.message.reply_text(
f"""
@moe_sticker_bot by @plow283
https://github.com/star-39/moe-sticker-bot
Credit: Thank you @StickerGroup for feedbacks and advices!
<code>
This software comes with ABSOLUTELY NO WARRANTY!
Released under the GPL v3 License. All rights reserved.
PRIVACY NOTICE:
    This software(bot) does not collect or save any kind of your personal information.
    Complies with Hong Kong Legislation Cap. 486 Personal Data (Privacy) Ordinance.
</code>
<b>
Please send /start to start using!
請傳送 /start 來開始!
始めるには　/start　を入力してください！
</b>
Advanced mode commands: 進階模式指令: <code>alsi</code>

<b>FAQ:</b>
=>Q: The generated sticker set ID has the bot's name as suffix! 
　　　創建的貼圖包ID末尾有這個bot的名字!
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

=>Q: No response? 沒有反應? 応答なし？
=>A: The bot might encountered an error, please report to GitHub issue:
     BOT可能遇到了問題, 請報告至GitHub issue網頁:
     https://github.com/star-39/moe-sticker-bot/issues
<code>
========================
BOT NAME: {GlobalConfigs.BOT_NAME}
VERSION: {GlobalConfigs.BOT_VERSION}
</code>
"""
        , parse_mode="HTML")


def do_auto_import_line_sticker(update, ctx):
    notify_import_starting(update, ctx)

    img_files_path = prepare_sticker_files(ctx, want_animated=False)
    # Create a new sticker set using the first image.
    try:
        ctx.bot.create_new_sticker_set(user_id=update.message.from_user.id,
                                     name=ctx.user_data['telegram_sticker_id'],
                                     title=ctx.user_data['telegram_sticker_title'],
                                     emojis=ctx.user_data['telegram_sticker_emoji'],
                                     png_sticker=open(img_files_path[0], 'rb'))
    except Exception as e:
        update.message.reply_text("Failed to create new sticker set!\n" + str(e))
        return ConversationHandler.END

    message_progress = report_progress(None, 1, len(img_files_path), update=update)
    for index, img_file_path in enumerate(img_files_path):
        # Skip the first file.
        if index != 0:
            # Retry 3 times
            for _ in range(3):
                try:
                    time.sleep(1)
                    ctx.bot.add_sticker_to_set(user_id=update.message.from_user.id,
                                             name=ctx.user_data['telegram_sticker_id'],
                                             emojis=ctx.user_data['telegram_sticker_emoji'],
                                             png_sticker=open(img_file_path, 'rb'))
                except telegram.error.RetryAfter as ra:
                    subprocess.run("date", shell=True)
                    print("!!! Flood limit triggered@do_auto_import_line_sticker, logging this incident.")
                    time.sleep(int(ra.retry_after))
                    # API sometimes return retry_after EVEN addStickerToSet has success! Check sticker set's actual status.
                    if index + 1 == ctx.bot.get_sticker_set(name=ctx.user_data['telegram_sticker_id']).stickers:
                        break
                    else:
                        continue
                except Exception as e:
                    update.message.reply_text("Fatal error! Please try again.\n" + str(e))
                    return ConversationHandler.END
                else:
                    break
            report_progress(message_progress, index + 1, len(img_files_path))

    notify_sticker_done(update, ctx)
    # clean up
    directory_path = os.path.dirname(img_files_path[0]) + "/"
    subprocess.run(f"rm -r {directory_path}*", shell=True)


def notify_import_starting(update, ctx):
    try:
        update.message.reply_text("Now starting, please wait...\n"
                                  "正在開始, 請稍後...\n"
                                  "作業がまもなく開始します、少々お時間を...\n\n"
                                  "<code>"
                                  f"LINE TYPE: {ctx.user_data['line_sticker_type']}\n"
                                  f"LINE ID: {ctx.user_data['line_sticker_id']}\n"
                                  f"TG ID: {ctx.user_data['telegram_sticker_id']}\n"
                                  f"TG TITLE: {ctx.user_data['telegram_sticker_title']}\n"
                                  "</code>",
                                  parse_mode="HTML", reply_markup=ReplyKeyboardRemove())
        if ctx.user_data['line_sticker_type'] == "sticker_message":
            update.message.reply_text("You are importing LINE Message Stickers which needs more time to complete.\n"
                                      "您正在匯入LINE訊息貼圖, 這需要更長的等候時間.")
    except:
        pass


def prepare_sticker_files(ctx, want_animated):
    os.makedirs("line_sticker", exist_ok=True)
    directory_path = "line_sticker/" + ctx.user_data['line_sticker_id'] + "/"
    os.makedirs(directory_path, exist_ok=True)
    subprocess.run(f"rm -r {directory_path}*", shell=True)
    if ctx.user_data['line_sticker_type'] == "sticker_message":
        for element in BeautifulSoup(ctx.user_data['line_store_webpage'].text, "html.parser").find_all('li'):
            json_text = element.get('data-preview')
            if json_text is not None:
                json_data = json.loads(json_text)
                base_image = json_data['staticUrl'].split(';')[0]
                overlay_image = json_data['customOverlayUrl'].split(';')[0]
                base_image_link_split = base_image.split('/')
                image_id = base_image_link_split[base_image_link_split.index('sticker') + 1]
                subprocess.run(f"curl -Lo {directory_path}{image_id}.base.png {base_image}", shell=True)
                subprocess.run(f"curl -Lo {directory_path}{image_id}.overlay.png {overlay_image}", shell=True)
                subprocess.run(f"convert {directory_path}{image_id}.base.png {directory_path}{image_id}.overlay.png "
                               f"-background none -filter Lanczos -resize 512x512 -composite -define webp:lossless=true "
                               f"{directory_path}{image_id}.webp", shell=True)
        return sorted([directory_path + f for f in os.listdir(directory_path) if
                       os.path.isfile(os.path.join(directory_path, f)) and f.endswith(".webp")])

    zip_file_path = "line_sticker/" + ctx.user_data['line_sticker_id'] + ".zip"
    subprocess.run(f"curl -Lo {zip_file_path} {ctx.user_data['line_sticker_download_url']}", shell=True)
    subprocess.run(f"bsdtar -xf {zip_file_path} -C {directory_path}", shell=True)
    if not want_animated:
        # Remove garbage
        subprocess.run(f"rm {directory_path}*key* {directory_path}tab* {directory_path}productInfo.meta", shell=True)
        # Resize to fulfill telegram's requirement, AR is automatically retained
        # Lanczos resizing produces much sharper image.
        subprocess.run(f"mogrify -background none -filter Lanczos -resize 512x512 "
                       f"-format webp -define webp:lossless=true {directory_path}*.png", shell=True)
        return sorted([directory_path + f for f in os.listdir(directory_path) if
                       os.path.isfile(os.path.join(directory_path, f)) and f.endswith("webp")])
    else:
        directory_path += "animation@2x/"
        # Magic!
        # LINE's apng has fps of 9, however ffmpeg defaults to 25
        subprocess.run(f'find {directory_path}*.png -type f -print0 | '
                       'xargs -I{} -0 ffmpeg -hide_banner -loglevel warning -i {} '
                       '-lavfi "color=white[c];[c][0]scale2ref[cs][0s];[cs][0s]overlay=shortest=1,setsar=1:1" '
                       '-c:v libx264 -r 9 -crf 26 -y {}.mp4', shell=True)
        return sorted([directory_path + f for f in os.listdir(directory_path) if
                       os.path.isfile(os.path.join(directory_path, f)) and f.endswith(".mp4")])


def report_progress(message_progress, current, total, update=None):
    progress_1 =  '[=>                  ]'
    progress_25 = '[====>               ]'
    progress_50 = '[=========>          ]'
    progress_75 = '[==============>     ]'
    progress_100 ='[====================]'
    try:
        if update is not None:
            return update.message.reply_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_1 + "</code>\n"
                                       "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == int(0.25 * total):
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_25 + "</code>\n"
                                       "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == int(0.5 * total):
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_50 + "</code>\n"
                                       "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == int(0.75 * total):
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_75 + "</code>\n"
                                       "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == total:
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_100 + "</code>\n"
                                       "<code>       " + str(current) + " of " + str(total) + "     </code>",
                                       parse_mode="HTML")
    except:
        pass


def do_download_line_sticker(update, ctx):
    update.message.reply_text(ctx.user_data['line_sticker_download_url'])


def initialise_manual_import(update, ctx):
    notify_import_starting(update, ctx)
    ctx.user_data['img_files_path'] = prepare_sticker_files(ctx, want_animated=False)
    # This is the FIRST sticker.
    ctx.user_data['manual_emoji_index'] = 0
    notify_next(update, ctx)


# MANUAL_EMOJI
def manual_add_emoji(update: Update, ctx: CallbackContext) -> int:
    # Verify emoji.
    em = ''.join(e for e in re.findall(emoji.get_emoji_regexp(), update.message.text))
    if em == '':
        update.message.reply_text("Please send emoji! Try again")
        return MANUAL_EMOJI

    # First sticker to create new set.
    if ctx.user_data['manual_emoji_index'] == 0:
        try:
            ctx.bot.create_new_sticker_set(user_id=update.message.from_user.id,
                                         name=ctx.user_data['telegram_sticker_id'],
                                         title=ctx.user_data['telegram_sticker_title'],
                                         emojis=em,
                                         png_sticker=open(ctx.user_data['img_files_path'][0], 'rb'))
        except Exception as e:
            update.message.reply_text("Error creating sticker set! Please try again!\n" + str(e))
            return ConversationHandler.END
    else:
        try:
            ctx.bot.add_sticker_to_set(user_id=update.message.from_user.id,
                                     name=ctx.user_data['telegram_sticker_id'],
                                     emojis=em,
                                     png_sticker=open(
                                         ctx.user_data['img_files_path'][ ctx.user_data['manual_emoji_index']], 'rb'
                                         )
                                     )
        # Catch not only retry_after exception when manual_emoji.
        except Exception as e:
            time.sleep(10)
            if ctx.user_data['manual_emoji_index'] + 1 == ctx.bot.get_sticker_set(name=ctx.user_data['telegram_sticker_id']).stickers:
                pass
            else:
                update.message.reply_text("Error assigning this one! Please send the same emoji again.\n" + str(e))
                return MANUAL_EMOJI

        if ctx.user_data['manual_emoji_index'] == len(ctx.user_data['img_files_path']) - 1:
            notify_sticker_done(update, ctx)
            # clean up
            directory_path = os.path.dirname(ctx.user_data['img_files_path'][0]) + "/"
            subprocess.run(f"rm -r {directory_path}*", shell=True)
            return ConversationHandler.END

    ctx.user_data['manual_emoji_index'] += 1
    notify_next(update, ctx)
    return MANUAL_EMOJI


def notify_next(update, ctx):
    for _ in range(3):
        try:
            ctx.bot.send_photo(chat_id=update.effective_chat.id,
                             caption="Please send emoji(s) representing this sticker\n"
                                     "請輸入代表這個貼圖的emoji(可以多個)\n"
                                     "このスタンプにふさわしい絵文字を入力してください(複数可)\n" +
                                     f"{ctx.user_data['manual_emoji_index'] + 1} of {len(ctx.user_data['img_files_path'])}",
                             photo=open(ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']], 'rb'))
        except:
            time.sleep(10)
            continue
        else:
            break


def notify_sticker_done(update, ctx):
    for _ in range(3):
        try:
            update.message.reply_text("The sticker set has been successfully created!\n"
                                      "貼圖包已經成功創建!\n"
                                      "ステッカーセットの作成が成功しました！\n\n"
                                      "https://t.me/addstickers/" + ctx.user_data['telegram_sticker_id'])
            if ctx.user_data['line_sticker_type'] == "sticker_animated":
                update.message.reply_text("It seems the sticker set you imported also has a animated version\n"
                                          "You can use /get_animated_line_sticker to have their GIF version\n"
                                          "您匯入的貼圖包還有動態貼圖版本\n"
                                          "可以使用 /get_animated_line_sticker 獲取GIF版動態貼圖\n"
                                          "このスタンプの動くバージョンもございます。\n"
                                          "/get_animated_line_sticker を使ってGIF版のスタンプを入手できます")
            update.message.reply_text(ctx.user_data['in_command'] + " done! 指令成功完成!")
        except:
            time.sleep(10)
            continue
        else:
            break

# TITLE
# This is the final conversaion step, if user wants to assign each sticker a different emoji, return to MANUAL_EMOJI,
# otherwise, do auto import then END conversation.
def parse_title(update: Update, ctx: CallbackContext) -> int:
    # Manual ID assignation is now removed, it seems meaningless.
    # Auto ID generation.
    ctx.user_data['telegram_sticker_id'] = f"line_{ctx.user_data['line_sticker_type']}_" \
                                             f"{ctx.user_data['line_sticker_id']}_" \
                                             f"{secrets.token_hex(nbytes=3)}_by_{GlobalConfigs.BOT_NAME}"


    if update.message.text.strip().lower() != "auto":
        ctx.user_data['telegram_sticker_title'] = update.message.text.strip()

    if ctx.user_data['manual_emoji'] is True:
        initialise_manual_import(update, ctx)
        return MANUAL_EMOJI
    else:
        do_auto_import_line_sticker(update, ctx)
        return ConversationHandler.END


# EMOJI
def parse_emoji(update: Update, ctx: CallbackContext) -> int:
    if update.message.text.strip().lower() == "manual":
        ctx.user_data['manual_emoji'] = True
    else:
        em = ''.join(e for e in re.findall(emoji.get_emoji_regexp(), update.message.text))
        if em == '':
            update.message.reply_text("Please send emoji! Try again")
            return EMOJI
        ctx.user_data['telegram_sticker_emoji'] = em

    ctx.user_data['telegram_sticker_title'] = BeautifulSoup(ctx.user_data['line_store_webpage'].text, 'html.parser')\
                                                            .find("title").get_text().split('|')[0].strip().split('LINE')[0][:-3] + \
                                                            f" @{GlobalConfigs.BOT_NAME}"
    update.message.reply_text(
        "Please set a title for this sticker set. Send <code>auto</code> to automatically set original title from LINE Store as shown below:\n"
        "請設定貼圖包的標題, 也就是名字. 輸入<code>auto</code>可以自動設為LINE Store中原版的標題如下:\n"
        "スタンプのタイトルを入力してください。<code>auto</code>を入力すると、LINE STORE上に表記されているのタイトルが自動的以下の通りに設定されます。" + "\n\n" + 
        "<code>" + ctx.user_data['telegram_sticker_title'] + "</code>" ,
        reply_markup=reply_kb_for_auto_markup,
        parse_mode="HTML")
    return TITLE


# LINE_STICKER_INFO
def parse_line_url(update: Update, ctx: CallbackContext) -> int:
    try:
        message_url = re.findall(r'\b(?:https?):[\w/#~:.?+=&%@!\-.:?\\-]+?(?=[.:?\-]*(?:[^\w/#~:.?+=&%@!\-.:?\-]|$))',
                                 update.message.text)[0]
        ctx.user_data['line_store_webpage'] = requests.get(message_url)
        ctx.user_data['line_sticker_url'], \
        ctx.user_data['line_sticker_type'], \
        ctx.user_data['line_sticker_id'], \
        ctx.user_data['line_sticker_download_url'] = get_line_sticker_detail(ctx.user_data['line_store_webpage'])
    except:
        update.message.reply_text('URL parse error! Make sure you sent a LINE Store URL !! Try again please.\n'
                                  'URL解析錯誤! 請確認輸入的是正確的LINE商店URL連結. 請重試.\n'
                                  'URL解析エラー！もう一度、正しいLINEスタンプストアのリンクを入力してください')
        return LINE_STICKER_INFO
    if str(ctx.user_data['in_command']).startswith("/import_line_sticker"):
        print_ask_emoji(update)
        return EMOJI
    elif str(ctx.user_data['in_command']).startswith("/download_line_sticker"):
        do_download_line_sticker(update, ctx)
        return ConversationHandler.END
    elif str(ctx.user_data['in_command']).startswith("/get_animated_line_sticker"):
        do_get_animated_line_sticker(update, ctx)
        return ConversationHandler.END
    else:
        pass


def do_get_animated_line_sticker(update, ctx):
    if ctx.user_data['line_sticker_type'] != "sticker_animated":
        update.message.reply_text("Sorry! This LINE Sticker set is NOT animated! Please check again.\n"
                                  "抱歉! 這個LINE貼圖包沒有動態版本! 請檢查連結是否有誤.\n"
                                  "このスタンプの動くバージョンはございません。もう一度ご確認してください。")
        return ConversationHandler.END
    notify_import_starting(update, ctx)
    for gif_file in prepare_sticker_files(ctx, want_animated=True):
        time.sleep(1)
        for _ in range(3):
            try:
                ctx.bot.send_animation(chat_id=update.effective_chat.id, animation=open(gif_file, 'rb'))
            except telegram.error.RetryAfter as ra:
                time.sleep(int(ra.retry_after))
                continue
            else:
                break


def get_line_sticker_detail(webpage):
    if not webpage.url.startswith("https://store.line.me") or not webpage.status_code == 200:
        raise Exception("Invalid link!")
    json_details = json.loads(BeautifulSoup(webpage.text, "html.parser").find_all('script')[0].contents[0])
    i = json_details['sku']
    url = json_details['url']
    url_comps = urlparse(url).path[1:].split('/')
    if url_comps[0] == "stickershop":
        # First one matches AnimatedSticker with NO sound and second one with sound.
        if 'MdIcoPlay_b' in webpage.text or 'MdIcoAni_b' in webpage.text:
            t = "sticker_animated"
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/iphone/stickerpack@2x.zip"
        elif 'MdIcoMessageSticker_b' in webpage.text:
            t = "sticker_message"
            u = webpage.url
        else:
            t = "sticker"
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/iphone/stickers@2x.zip"
    elif url_comps[0] == "emojishop":
        t = "emoji"
        u = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + i + "/iphone/package.zip"
    else:
        raise Exception("Not a supported sticker type!")

    return url, t, i, u


def command_import_line_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_line_store_link(update)
    return LINE_STICKER_INFO


def command_download_line_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_line_store_link(update)
    return LINE_STICKER_INFO


def command_get_animated_line_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_line_store_link(update)
    return LINE_STICKER_INFO


def initialize_user_data(update, ctx):
    ctx.user_data['in_command'] = update.message.text
    ctx.user_data['manual_emoji'] = False
    ctx.user_data['line_sticker_url'] = ""
    ctx.user_data['line_store_webpage'] = None
    ctx.user_data['line_sticker_download_url'] = ""
    ctx.user_data['line_sticker_type'] = "sticker"
    ctx.user_data['line_sticker_id'] = ""
    ctx.user_data['telegram_sticker_emoji'] = ""
    ctx.user_data['telegram_sticker_id'] = ""
    ctx.user_data['telegram_sticker_title'] = ""


def command_alsi(update: Update, ctx: CallbackContext) -> int:
    update.message.reply_text("INFO: You are using Advanced Line Sticker Import (alsi), be sure syntax is correct.")
    if update.message.text.startswith("alsi"):
        alsi_parser = argparse.ArgumentParser(prog="alsi", exit_on_error=False, add_help=False, 
                                              formatter_class=argparse.RawDescriptionHelpFormatter,
                                              description="Advanced Line Sticker Import",
                                              epilog='Example usage:\n'
                                              '  alsi -id=exmaple_id_00 -title="Example Title" -link=https://store.line.me/stickershop/product/9124676/ja\n\n'
                                              'Note:\n  Argument containing white space must be closed by quotes.\n'
                                              '  ID must contain alphabet, number and underscore only.')
        alsi_parser.add_argument('-id', help="Telegram sticker name(ID), used for share link", required=True)
        alsi_parser.add_argument('-title', help="Telegram sticker set title", required=True)
        alsi_parser.add_argument('-link', help="LINE Store link of LINE sticker pack", required=True)
        try:
            alsi_args = alsi_parser.parse_args(shlex.split(update.message.text)[1:])
        except:
            update.message.reply_text("Wrong syntax!!\n" + "<code>" + alsi_parser.format_help() + "</code>", parse_mode="HTML")
            return ConversationHandler.END
        # initialise
        ctx.user_data['in_command'] = "alsi"
        ctx.user_data['manual_emoji'] = True
        ctx.user_data['line_sticker_url'] = alsi_args.link
        ctx.user_data['line_store_webpage'] = None
        ctx.user_data['line_sticker_download_url'] = ""
        ctx.user_data['line_sticker_type'] = "sticker"
        ctx.user_data['line_sticker_id'] = ""
        # parse link
        try:
            ctx.user_data['line_store_webpage'] = requests.get(alsi_args.link)
            ctx.user_data['line_sticker_url'], \
            ctx.user_data['line_sticker_type'], \
            ctx.user_data['line_sticker_id'], \
            ctx.user_data['line_sticker_download_url'] = get_line_sticker_detail(ctx.user_data['line_store_webpage'])
        except:
            update.message.reply_text("Wrong link!!")
            return ConversationHandler.END
        # add id and title
        if not re.match('^\w+$',alsi_args.id):
            update.message.reply_text("Wrong ID!!")
            return ConversationHandler.END
        ctx.user_data['telegram_sticker_id'] = alsi_args.id + "_by_" + GlobalConfigs.BOT_NAME
        ctx.user_data['telegram_sticker_title'] = alsi_args.title

        initialise_manual_import(update, ctx)
        return MANUAL_EMOJI


    else:
        update.message.reply_text("wrong command!")
        return ConversationHandler.END 


# GET_TG_STICKER
def parse_tg_sticker(update: Update, ctx: CallbackContext) -> int:
    sticker_set = ctx.bot.get_sticker_set(name=update.message.sticker.set_name)
    update.message.reply_text("This might take some time, please wait...\n"
                              "此項作業可能需時較長, 請稍等...\n"""
                              "少々お待ちください...\n"
                              "<code>\n"
                              f"Name: {sticker_set.name}\n"
                              f"Title: {sticker_set.title}\n"
                              f"Amount: {str(len(sticker_set.stickers))}\n"
                              "</code>",
                              parse_mode="HTML")
    save_path = "tg_sticker/" + sticker_set.name + "/"
    os.makedirs(save_path, exist_ok=True)
    subprocess.run("rm -r " + save_path + "*", shell=True)
    for index, sticker in enumerate(sticker_set.stickers):
        try:
            ctx.bot.get_file(sticker.file_id).download(save_path + sticker.set_name +
                                                     "_" + str(index).zfill(3) + "_" +
                                                      emoji.demojize(sticker.emoji)[1:-1] +
                                                     (".tgs" if sticker_set.is_animated else ".webp"))
        except Exception as e:
            print(str(e))
    if sticker_set.is_animated:
        subprocess.run(f'find {save_path}*.tgs -type f -print0 | '
                       "xargs -I{} -0 lottie_convert.py {} {}.webp", shell=True)
        subprocess.run("bsdtar -acvf " + save_path + sticker_set.name + "_tgs.zip " + save_path + "*.tgs", shell=True)
    else:
        subprocess.run("mogrify -format png " + save_path + "*.webp", shell=True)
        subprocess.run("bsdtar -acvf " + save_path + sticker_set.name + "_png.zip " + save_path + "*.png", shell=True)

    subprocess.run("bsdtar -acvf " + save_path + sticker_set.name + "_webp.zip " + save_path + "*.webp", shell=True)

    try:
        if sticker_set.is_animated:
            ctx.bot.send_document(chat_id=update.effective_chat.id,
                                document=open(save_path + sticker_set.name + "_tgs.zip", 'rb'))
            ctx.bot.send_document(chat_id=update.effective_chat.id,
                                document=open(save_path + sticker_set.name + "_webp.zip", 'rb'))
        else:
            ctx.bot.send_document(chat_id=update.effective_chat.id,
                                document=open(save_path + sticker_set.name + "_webp.zip", 'rb'))
            ctx.bot.send_document(chat_id=update.effective_chat.id,
                                document=open(save_path + sticker_set.name + "_png.zip", 'rb'))
    except Exception as e:
        print(str(e))

    # clean up
    subprocess.run("rm -r " + save_path + "*", shell=True)

    return ConversationHandler.END


def command_download_telegram_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    update.message.reply_text("Please send a sticker.\n"
                              "請傳送一張Telegram貼圖.\n"
                              "ステッカーを一つ送信してください。")
    return GET_TG_STICKER


def print_ask_emoji(update):
    update.message.reply_text("Please enter emoji representing this sticker set\n"
                              "請輸入用於表示整個貼圖包的emoji\n"
                              "このスタンプセットにふさわしい絵文字を入力してください\n"
                              "eg. ☕ \n\n"
                              "If you want to manually assign different emoji for each sticker, send <code>manual</code>\n"
                              "如果您想要為貼圖包內每個貼圖設定不同的emoji, 請傳送<code>manual</code>\n"
                              "一つずつ絵文字を付けたいなら、<code>manual</code>を送信してください。",
                              reply_markup=reply_kb_for_manual_markup,
                              parse_mode="HTML")


def print_ask_line_store_link(update):
    update.message.reply_text("Please enter LINE store URL of the sticker set\n"
                              "請輸入貼圖包的LINE STORE連結\n"
                              "スタンプのLINE STOREリンクを入力してください\n\n"
                              "<code>eg. https://store.line.me/stickershop/product/9961437/ja</code>",
                              parse_mode="HTML")


def command_cancel(update: Update, ctx: CallbackContext) -> int:
    update.message.reply_text("SESSION END.")
    start(update, ctx)
    return ConversationHandler.END


def handle_text_message(update: Update, ctx: CallbackContext):
    update.message.reply_text("Please use /start to see available commands!\n"
                            "請先傳送 /start 來看看可用的指令\n"
                            "/start を送信してコマンドで始めましょう")
    if update.message.text.startswith("https://store.line.me") or update.message.text.startswith("https://line.me"):
        update.message.reply_text("You have sent a LINE Store link, guess you want to import LINE sticker to Telegram? Please send /import_line_sticker\n"
                                    "您傳送了一個LINE商店連結, 是想要把LINE貼圖包匯入至Telegram嗎? 請使用 /import_line_sticker\n"
                                    "LINEスタンプをインポートしたいんですか？ /import_line_sticker で始めてください")

def handel_sticker_message(update: Update, ctx: CallbackContext):
    update.message.reply_text("Please use /start to see available commands!\n"
                            "請先傳送 /start 來看看可用的指令\n"
                            "/start を送信してコマンドで始めましょう")
    update.message.reply_text("You have sent a sticker, guess you want to download this sticker set? Please send /download_telegram_sticker\n"
                              "您傳送了一個貼圖, 是想要下載這個Telegram貼圖包嗎? 請使用 /download_telegram_sticker\n"
                              "このステッカーセットを丸ごとダウンロードしようとしていますか？ /download_telegram_sticker で始めてください")


def print_warning(update: Update, ctx: CallbackContext):
    update.message.reply_text("You are already in command : " + ctx.user_data['in_command'][1:] + "\n"
                              "If you encountered a problem, please send /cancel and start over.")


def main() -> None:
    # Load configs
    config = configparser.ConfigParser()
    config.read('config.ini')
    GlobalConfigs.BOT_TOKEN = os.getenv("BOT_TOKEN", config['TELEGRAM']['BOT_TOKEN'])
    GlobalConfigs.BOT_NAME = Bot(GlobalConfigs.BOT_TOKEN).get_me().username
    updater = Updater(GlobalConfigs.BOT_TOKEN)

    dispatcher = updater.dispatcher

    conv_advanced_import = ConversationHandler(
        entry_points=[MessageHandler(Filters.regex('^alsi*') & ~Filters.command, command_alsi)],
        states={
            MANUAL_EMOJI : [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(Filters.command, print_warning)],
        run_async=True
    )

    # Each conversation is time consuming, enable run_async
    conv_import_line_sticker = ConversationHandler(
        entry_points=[CommandHandler('import_line_sticker', command_import_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji)],
            # ID: [MessageHandler(Filters.text & ~Filters.command, parse_id)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title)],
            MANUAL_EMOJI : [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(Filters.command, print_warning)],
        run_async=True
    )
    conv_get_animated_line_sticker = ConversationHandler(
        entry_points=[CommandHandler('get_animated_line_sticker', command_get_animated_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(Filters.command, print_warning)],
        run_async=True
    )
    conv_download_line_sticker = ConversationHandler(
        entry_points=[CommandHandler('download_line_sticker', command_download_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(Filters.command, print_warning)],
        run_async=True
    )
    conv_download_telegram_sticker = ConversationHandler(
        entry_points=[CommandHandler('download_telegram_sticker', command_download_telegram_sticker)],
        states={
            GET_TG_STICKER: [MessageHandler(Filters.sticker, parse_tg_sticker)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(Filters.command, print_warning)],
        run_async=True
    )
    # 派遣します！
    dispatcher.add_handler(conv_import_line_sticker)
    dispatcher.add_handler(conv_get_animated_line_sticker)
    dispatcher.add_handler(conv_download_line_sticker)
    dispatcher.add_handler(conv_download_telegram_sticker)
    dispatcher.add_handler(conv_advanced_import)
    dispatcher.add_handler(CommandHandler('start', start))
    dispatcher.add_handler(CommandHandler('help', help_command))
    dispatcher.add_handler(MessageHandler(Filters.text & ~Filters.command, handle_text_message))
    dispatcher.add_handler(MessageHandler(Filters.sticker, handel_sticker_message))

    updater.start_polling()
    updater.idle()


if __name__ == '__main__':
    main()
