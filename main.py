# moe-sticker-bot @moe_sticker_bot
# Copyright (c) 2020-2021, @plow283 @star-39. All rights reserved
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

import json
import time
# import logging
from typing import Any
from urllib.parse import urlparse
# from telegram.callbackquery import CallbackQuery
import telegram.error
from telegram import Update, Bot, Update, message
from telegram.ext import Updater, CommandHandler, CallbackContext, ConversationHandler, MessageHandler, Filters, CallbackQueryHandler
from bs4 import BeautifulSoup
import emoji
import requests
import re
import os
import subprocess
import secrets
import traceback
import argparse
import shlex

from telegram.ext.utils.types import ConversationDict
from notifications import *
import shutil
import glob

BOT_VERSION = "2.2 RC-1"
BOT_TOKEN = os.getenv("BOT_TOKEN")
BOT_NAME = Bot(BOT_TOKEN).get_me().username
DATA_DIR = BOT_NAME + "_data"

# Stages of conversations
LINE_STICKER_INFO, EMOJI, TITLE, MANUAL_EMOJI = range(4)
STICKER_ARCHIVE, EMOJI, TITLE, ID, MANUAL_EMOJI = range(5)
GET_TG_STICKER = range(1)


# Uploading sticker could easily trigger TG's flood limit,
# however, documentation never specified this limit,
# hence, we should at least retry after triggering the limit.
def retry_do(func) -> Any:
    for index in range(3):
        try:
            func()
        except telegram.error.RetryAfter as ra:
            if index == 2:
                raise e
            time.sleep(int(ra.retry_after))

        except Exception as e:
            if index == 2:
                raise e
            # It could probably be a network problem, sleep for a while and try again.
            time.sleep(5)
        else:
            break

# Clean temparary user data after each conversasion.
def clean_userdata(update: Update):
    userdata_dir = os.path.join(DATA_DIR, str(update.effective_user.id))
    if os.path.exists(userdata_dir):
        shutil.rmtree(userdata_dir, ignore_errors=True)


def do_auto_create_sticker_set(update: Update, ctx):
    print_import_starting(update, ctx)

    img_files_path = prepare_sticker_files(update, ctx, False)
    # Create a new sticker set using the first image.
    try:
        retry_do(lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                        name=ctx.user_data['telegram_sticker_id'],
                                                        title=ctx.user_data['telegram_sticker_title'],
                                                        emojis=ctx.user_data['telegram_sticker_emoji'],
                                                        png_sticker=open(img_files_path[0], 'rb')))
    except Exception as e:
        print_fatal_error(update, traceback.format_exc())
        return

    message_progress = print_progress(
        None, 1, len(img_files_path), update=update)
    for index, img_file_path in enumerate(img_files_path):
        # Skip the first file.
        if index == 0:
            continue
        print_progress(message_progress, index + 1, len(img_files_path))
        try:
            retry_do(lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                        name=ctx.user_data['telegram_sticker_id'],
                                                        emojis=ctx.user_data['telegram_sticker_emoji'],
                                                        png_sticker=open(img_file_path, 'rb')))
        except:
            print_fatal_error(update, traceback.format_exc())
            return

    print_sticker_done(update, ctx)


def do_get_animated_line_sticker(update, ctx):
    if ctx.user_data['line_sticker_type'] != "sticker_animated":
        print_not_animated_warning(update)
        return ConversationHandler.END
    print_import_starting(update, ctx)
    for gif_file in prepare_sticker_files(update, ctx, want_animated=True):
        try:
            retry_do(lambda: ctx.bot.send_animation(
                chat_id=update.effective_chat.id, animation=open(gif_file, 'rb')))
        except:
            print_fatal_error(update, traceback.format_exc())
    print_command_done(update, ctx)


def prepare_sticker_files(update: Update, ctx, want_animated):
    # User uploaded stickers
    if str(ctx.user_data['in_command']).startswith("/create_sticker_set"):
        archive_path = ctx.user_data['tg_sticker_archive']
        sticker_dir = os.path.dirname(ctx.user_data['tg_sticker_archive'])
        subprocess.run(['bsdtar', '-xf', archive_path, '-C', sticker_dir])
        for f in [f for f in glob.glob(os.path.join(sticker_dir, "**"), recursive=True) if os.path.isfile(f)]:
            if os.path.isfile(f):
                subprocess.run(["mogrify", "-background", "none", "-filter", "Lanczos", "-resize", "512x512",
                                "-format", "webp", "-define", "webp:lossless=true", f + "[0]"])
    # line stickers
    else:
        sticker_dir = os.path.join(
            DATA_DIR, str(update.effective_user.id), ctx.user_data['line_sticker_id'])
        os.makedirs(sticker_dir, exist_ok=True)
        # Special line "message" stickers
        if ctx.user_data['line_sticker_type'] == "sticker_message":
            for element in BeautifulSoup(ctx.user_data['line_store_webpage'].text, "html.parser").find_all('li'):
                json_text = element.get('data-preview')
                if json_text is not None:
                    json_data = json.loads(json_text)
                    base_image = json_data['staticUrl'].split(';')[0]
                    overlay_image = json_data['customOverlayUrl'].split(';')[0]
                    base_image_link_split = base_image.split('/')
                    image_id = base_image_link_split[base_image_link_split.index(
                        'sticker') + 1]
                    subprocess.run(
                        ["curl", "-Lo", f"{sticker_dir}{image_id}.base.png", base_image])
                    subprocess.run(
                        ["curl", "-Lo", f"{sticker_dir}{image_id}.overlay.png", overlay_image])
                    subprocess.run(["convert", f"{sticker_dir}{image_id}.base.png", f"{sticker_dir}{image_id}.overlay.png",
                                    "-background", "none", "-filter", "Lanczos", "-resize", "512x512", "-composite",
                                    "-define", "webp:lossless=true",
                                    f"{sticker_dir}{image_id}.webp"])
        # normal line stickers
        else:
            zip_file_path = os.path.join(
                sticker_dir, ctx.user_data['line_sticker_id'] + ".zip")
            subprocess.run(["curl", "-Lo", zip_file_path,
                            ctx.user_data['line_sticker_download_url']])
            subprocess.run(["bsdtar", "-xf", zip_file_path, "-C", sticker_dir])
            if not want_animated:
                # Remove garbage
                for f in glob.glob(os.path.join(sticker_dir, "*key*")) + glob.glob(os.path.join(sticker_dir, "tab*")) + glob.glob(os.path.join(sticker_dir, "productInfo.meta")):
                    os.remove(f)
                # Resize to fulfill telegram's requirement, AR is automatically retained
                # Lanczos resizing produces much sharper image.
                for f in glob.glob(os.path.join(sticker_dir, "*")):
                    subprocess.run(["mogrify", "-background", "none", "-filter", "Lanczos", "-resize", "512x512",
                                    "-format", "webp", "-define", "webp:lossless=true", f])
            else:
                sticker_dir = os.path.join(sticker_dir, "animation@2x")
                # LINE's apng has fps of 9, however ffmpeg defaults to 25
                for f in glob.glob(os.path.join(sticker_dir, "*.png")):
                    subprocess.run(["ffmpeg", "-hide_banner", "-loglevel", "warning", "-i", f,
                                    "-lavfi", 'color=white[c];[c][0]scale2ref[cs][0s];[cs][0s]overlay=shortest=1,setsar=1:1',
                                    "-c:v", "libx264", "-r", "9", "-crf", "26", "-y", f + ".mp4"])
                return sorted([f for f in glob.glob(os.path.join(sticker_dir, "*.mp4"))])

    return sorted([f for f in glob.glob(os.path.join(sticker_dir, "**", "*.webp"), recursive=True)])


def initialize_manual_emoji(update: Update, ctx: CallbackContext):
    print_import_starting(update, ctx)
    ctx.user_data['img_files_path'] = prepare_sticker_files(update, ctx, False)
    # This is the FIRST sticker.
    ctx.user_data['manual_emoji_index'] = 0
    notify_next(update, ctx)


# MANUAL_EMOJI
def manual_add_emoji(update: Update, ctx: CallbackContext) -> int:
    # Verify emoji.
    em = ''.join(e for e in re.findall(
        emoji.get_emoji_regexp(), update.message.text))
    if em == '':
        update.message.reply_text("Please send emoji! Try again")
        return MANUAL_EMOJI

    # First sticker to create new set.
    if ctx.user_data['manual_emoji_index'] == 0:
        try:
            retry_do(lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                            name=ctx.user_data['telegram_sticker_id'],
                                                            title=ctx.user_data['telegram_sticker_title'],
                                                            emojis=em,
                                                            png_sticker=open(ctx.user_data['img_files_path'][0], 'rb')))
        except Exception as e:
            clean_userdata(update)
            print_sticker_done(update, ctx)
            return ConversationHandler.END
    else:
        try:
            retry_do(lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                        name=ctx.user_data['telegram_sticker_id'],
                                                        emojis=em,
                                                        png_sticker=open(
                                                            ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']], 'rb'
                                                        )))
        except:
            clean_userdata(update)
            print_sticker_done(update, ctx)
            return ConversationHandler.END

        if ctx.user_data['manual_emoji_index'] == len(ctx.user_data['img_files_path']) - 1:
            clean_userdata(update)
            print_sticker_done(update, ctx)
            return ConversationHandler.END

    ctx.user_data['manual_emoji_index'] += 1
    notify_next(update, ctx)
    return MANUAL_EMOJI


def notify_next(update, ctx):
    ctx.bot.send_photo(chat_id=update.effective_chat.id,
                       caption="Please send emoji(s) representing this sticker\n"
                       "請傳送代表這個貼圖的emoji(可以多個)\n"
                       "このスタンプにふさわしい絵文字を送信してください(複数可)\n" +
                       f"{ctx.user_data['manual_emoji_index'] + 1} of {len(ctx.user_data['img_files_path'])}",
                       photo=open(ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']], 'rb'))


# ID
# Currently only /create_sticker_set will hit this expression.
def parse_id(update: Update, ctx: CallbackContext) -> int:
    ctx.user_data['telegram_sticker_id'] = update.message.text.strip(
    ) + "_by_" + BOT_NAME
    if not re.match(r'^[a-zA-Z0-9_]+$', ctx.user_data['telegram_sticker_id']):
        print_wrong_id_syntax(update)
        return ID
    if ctx.user_data['manual_emoji'] is True:
        initialize_manual_emoji(update, ctx)
        return MANUAL_EMOJI
    else:
        do_auto_create_sticker_set(update, ctx)
        clean_userdata(update)
        return ConversationHandler.END


# TITLE
# This is the final conversaion step, if user wants to assign each sticker a different emoji, return to MANUAL_EMOJI,
# otherwise, do auto import then END conversation.
def parse_title(update: Update, ctx: CallbackContext) -> int:
    if str(ctx.user_data['in_command']).startswith("/create_sticker_set"):
        ctx.user_data['telegram_sticker_title'] = update.message.text.strip()
        print_ask_id(update)
        return ID

    if update.callback_query is None:
        ctx.user_data['telegram_sticker_title'] = update.message.text.strip()
    elif update.callback_query.data == "auto":
        # Auto ID generation if NOT creating sticker set.
        ctx.user_data['telegram_sticker_id'] = f"line_{ctx.user_data['line_sticker_type']}_" \
            f"{ctx.user_data['line_sticker_id']}_" \
            f"{secrets.token_hex(nbytes=3)}_by_{BOT_NAME}"
        update.callback_query.answer()
        edit_inline_kb_auto_selected(update.callback_query)
    else:
        return TITLE

    if ctx.user_data['manual_emoji'] is True:
        initialize_manual_emoji(update, ctx)
        return MANUAL_EMOJI
    else:
        do_auto_create_sticker_set(update, ctx)
        clean_userdata(update)
        return ConversationHandler.END


# EMOJI
def parse_emoji(update: Update, ctx: CallbackContext) -> int:
    if update.callback_query is None:
        emojis = ''.join(e for e in re.findall(
            emoji.get_emoji_regexp(), update.message.text))
        if emojis == '':
            update.message.reply_text("Please send emoji! Try again")
            return EMOJI
        ctx.user_data['telegram_sticker_emoji'] = emojis
    elif update.callback_query.data == "manual":
        ctx.user_data['manual_emoji'] = True
        update.callback_query.answer()
        edit_inline_kb_manual_selected(update.callback_query)
    elif update.callback_query.data == "random":
        ctx.user_data['telegram_sticker_emoji'] = "⭐️"
        update.callback_query.answer()
        edit_inline_kb_random_selected(update.callback_query)
    else: 
        return EMOJI

    # Generate auto title if NOT creating sticker set.
    if not str(ctx.user_data['in_command']).startswith("/create_sticker_set"):
        ctx.user_data['telegram_sticker_title'] = BeautifulSoup(ctx.user_data['line_store_webpage'].text, 'html.parser')\
            .find("title").get_text().split('|')[0].strip().split('LINE')[0][:-3] + \
            f" @{BOT_NAME}"
    print_ask_title(update, ctx.user_data['telegram_sticker_title'])
    return TITLE


# LINE_STICKER_INFO
def parse_line_url(update: Update, ctx: CallbackContext) -> int:
    try:
        get_line_sticker_detail(update.message.text, ctx)
    except Exception as e:
        print_wrong_LINE_STORE_URL(update, str(e))  
        return LINE_STICKER_INFO
    if str(ctx.user_data['in_command']).startswith("/import_line_sticker"):
        print_ask_emoji(update)
        return EMOJI
    elif str(ctx.user_data['in_command']).startswith("/download_line_sticker"):
        update.message.reply_text(ctx.user_data['line_sticker_download_url'])
        return ConversationHandler.END
    elif str(ctx.user_data['in_command']).startswith("/get_animated_line_sticker"):
        do_get_animated_line_sticker(update, ctx)
        clean_userdata(update)
        return ConversationHandler.END
    else:
        pass


def get_line_sticker_detail(message, ctx: CallbackContext):
    message_url = re.findall(r'\b(?:https?):[\w/#~:.?+=&%@!\-.:?\\-]+?(?=[.:?\-]*(?:[^\w/#~:.?+=&%@!\-.:?\-]|$))',
                                 message)[0]
    webpage = requests.get(message_url)
    ctx.user_data['line_store_webpage'] = webpage
    if not webpage.url.startswith("https://store.line.me"):
        raise Exception("Not a LINE Store link! 不是LINE商店連結!")
    json_details = json.loads(BeautifulSoup(
        webpage.text, "html.parser").find_all('script')[0].contents[0])
    i = json_details['sku']
    url = json_details['url']
    url_comps = urlparse(url).path[1:].split('/')
    if url_comps[0] == "stickershop":
        # First one matches AnimatedSticker with NO sound and second one with sound.
        if 'MdIcoPlay_b' in webpage.text or 'MdIcoAni_b' in webpage.text:
            t = "sticker_animated"
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + \
                i + "/iphone/stickerpack@2x.zip"
        elif 'MdIcoMessageSticker_b' in webpage.text:
            t = "sticker_message"
            u = webpage.url
        else:
            t = "sticker"
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + \
                i + "/iphone/stickers@2x.zip"
    elif url_comps[0] == "emojishop":
        t = "emoji"
        u = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + \
            i + "/iphone/package.zip"
    else:
        raise Exception("Not a supported sticker type!\nLink is: " + url)

    ctx.user_data['line_sticker_url'] = url
    ctx.user_data['line_sticker_type'] = t
    ctx.user_data['line_sticker_id'] = i
    ctx.user_data['line_sticker_download_url'] = u


def command_import_line_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    if 'last_user_message_timestamp' in ctx.user_data and int(time.time()) - ctx.user_data['last_user_message_timestamp'] < 60:
        if ctx.user_data['user_sent_line_link'] is True:
            last_user_message = ctx.user_data['last_user_message']
            ctx.user_data['user_sent_line_link'] = False
            ctx.user_data['last_user_message'] = ''
            try:
                get_line_sticker_detail(last_user_message, ctx)
            except Exception as e:
                print_wrong_LINE_STORE_URL(update, str(e))  
                return LINE_STICKER_INFO
            print_ask_emoji(update)
            return EMOJI
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


def initialize_user_data(update: Update, ctx):
    clean_userdata(update)
    ctx.user_data['in_command'] = update.message.text.strip().split(' ')[0]
    ctx.user_data['manual_emoji'] = False
    ctx.user_data['line_sticker_url'] = ""
    ctx.user_data['line_store_webpage'] = None
    ctx.user_data['line_sticker_download_url'] = ""
    ctx.user_data['line_sticker_type'] = ""
    ctx.user_data['line_sticker_id'] = ""
    ctx.user_data['telegram_sticker_emoji'] = ""
    ctx.user_data['telegram_sticker_id'] = ""
    ctx.user_data['telegram_sticker_title'] = ""


def command_alsi(update: Update, ctx: CallbackContext) -> int:
    # if update.message.text.startswith("alsi"):
    alsi_parser = argparse.ArgumentParser(
        prog="alsi", exit_on_error=False, add_help=False,
        formatter_class=argparse.RawTextHelpFormatter,
        description="Advanced Line Sticker Import - command line tool to import LINE sticker.\n進階LINE貼圖匯入程式",
        epilog='Example usage:\n'
        '  alsi -id=example_id_1 -title="Example Title" -link=https://store.line.me/stickershop/product/8898\n\n'
        'Note:\n  Argument containing white space must be closed by quotes.\n'
        '  ID must contain alphabet, number and underscore only.')
    alsi_parser.add_argument(
        '-emoji', help="Emoji to assign to whole sticker set, ignore this option to assign manually\n"
        "指定給整個貼圖包的Emoji, 忽略這個選項來手動指定貼圖的emoji", required=False)
    alsi_parser.add_argument(
        '-id', help="Telegram sticker name(ID), used for share link\nTelegram貼圖包ID, 用於分享連結", required=True)
    alsi_parser.add_argument(
        '-title', help="Telegram sticker set title\nTelegram貼圖包的標題", required=True)
    alsi_parser.add_argument(
        '-link', help="LINE Store link of LINE sticker pack\nLINE商店貼圖包連結", required=True)
    try:
        alsi_args = alsi_parser.parse_args(
            shlex.split(update.message.text)[1:])
    except:
        update.message.reply_text(
            "Wrong syntax!!\n" + "<code>" + alsi_parser.format_help() + "</code>", parse_mode="HTML")
        return ConversationHandler.END
    # initialise
    initialize_user_data(update, ctx)
    # parse link
    try:
        get_line_sticker_detail(update.message.text, ctx)
    except:
        update.message.reply_text("Wrong link!!")
        return ConversationHandler.END
    # add id and title
    if not re.match('^\w+$', alsi_args.id):
        update.message.reply_text("Wrong ID!!")
        return ConversationHandler.END
    ctx.user_data['telegram_sticker_id'] = alsi_args.id + \
        "_by_" + BOT_NAME
    ctx.user_data['telegram_sticker_title'] = alsi_args.title

    if alsi_args.emoji is None:
        initialize_manual_emoji(update, ctx)
        return MANUAL_EMOJI
    else:
        ctx.user_data['telegram_sticker_emoji'] = alsi_args.emoji
        do_auto_create_sticker_set(update, ctx)
        return ConversationHandler.END


# GET_TG_STICKER
def prepare_tg_sticker(update: Update, ctx: CallbackContext) -> int:
    sticker_set = ctx.bot.get_sticker_set(name=update.message.sticker.set_name)
    print_preparing_tg_sticker(
        update, sticker_set.title, sticker_set.name, str(len(sticker_set.stickers)))
    sticker_dir = os.path.join(
        DATA_DIR, str(update.effective_user.id), sticker_set.name)
    os.makedirs(sticker_dir, exist_ok=True)
    for index, sticker in enumerate(sticker_set.stickers):
        try:
            ctx.bot.get_file(sticker.file_id).download(os.path.join(sticker_dir,
                                                                    sticker.set_name +
                                                                    "_" + str(index).zfill(3) + "_" +
                                                                    emoji.demojize(sticker.emoji)[1:-1] +
                                                                    (".tgs" if sticker_set.is_animated else ".webp")))
        except Exception as e:
            print_fatal_error(update, traceback.format_exc())
            clean_userdata(update)
            return ConversationHandler.END
    webp_zip = os.path.join(sticker_dir, sticker_set.name + "_webp.zip")
    tgs_zip = os.path.join(sticker_dir, sticker_set.name + "_tgs.zip")
    png_zip = os.path.join(sticker_dir, sticker_set.name + "_png.zip")

    if sticker_set.is_animated:
        fs = glob.glob(os.path.join(sticker_dir, "*.tgs"))
        for f in fs:
            subprocess.run(["lottie_convert.py", f, f + ".webp"])
        subprocess.run(["bsdtar", "--strip-components", "2", "-acvf", tgs_zip] + fs)
    else:
        subprocess.run(["mogrify", "-format", "png"] +
                       glob.glob(os.path.join(sticker_dir, "*.webp")))
        subprocess.run(["bsdtar", "--strip-components", "2", "-acvf", png_zip] +
                       glob.glob(os.path.join(sticker_dir, "*.png")))

    subprocess.run(["bsdtar", "--strip-components", "2", "-acvf", webp_zip] +
                   glob.glob(os.path.join(sticker_dir, "*.webp")))

    try:
        ctx.bot.send_document(chat_id=update.effective_chat.id,
                              document=open(webp_zip, 'rb'))
        if sticker_set.is_animated:
            ctx.bot.send_document(chat_id=update.effective_chat.id,
                                  document=open(tgs_zip, 'rb'))
        else:
            ctx.bot.send_document(chat_id=update.effective_chat.id,
                                  document=open(png_zip, 'rb'))
    except Exception as e:
        print_fatal_error(update, traceback.format_exc())

    clean_userdata(update)
    print_command_done(update, ctx)

    return ConversationHandler.END


# STICKER_ARCHIVE
def parse_sticker_archive(update: Update, ctx: CallbackContext) -> int:
    archive_hash = secrets.token_hex(nbytes=4)
    archive_dir = os.path.join(
        DATA_DIR, str(update.effective_user.id), archive_hash)
    os.makedirs(archive_dir, exist_ok=True)
    # libarchive is smart enough to recognize actual archive format.
    archive_file_path = os.path.join(archive_dir, archive_hash + ".archive")
    try:
        update.message.document.get_file().download(archive_file_path)
    except Exception as e:
        print_fatal_error(update, traceback.format_exc())
        clean_userdata(update)
        return ConversationHandler.END

    ctx.user_data['tg_sticker_archive'] = archive_file_path
    print_ask_emoji(update)
    return EMOJI


def command_download_telegram_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_telegram_sticker(update)
    return GET_TG_STICKER


def command_create_sticker_set(update: Update, ctx: CallbackContext) -> int:
    initialize_user_data(update, ctx)
    print_ask_sticker_archive(update)
    return STICKER_ARCHIVE


def command_cancel(update: Update, ctx: CallbackContext) -> int:
    clean_userdata(update)
    print_command_canceled(update)
    command_start(update, ctx)
    return ConversationHandler.END


def handle_text_message(update: Update, ctx: CallbackContext):
    print_use_start_command(update)
    ctx.user_data['last_user_message'] = update.message.text
    ctx.user_data['last_user_message_timestamp'] = int(time.time())
    if update.message.text.startswith("https://store.line.me") or update.message.text.startswith("https://line.me"):
        ctx.user_data['user_sent_line_link'] = True
        print_suggest_import(update)


def handle_sticker_message(update: Update, ctx: CallbackContext):
    print_use_start_command(update)
    print_suggest_download(update)


def command_about(update: Update, ctx: CallbackContext):
    print_about_message(update, BOT_NAME, BOT_VERSION)


def command_faq(update: Update, ctx: CallbackContext):
    print_faq_message(update)


def command_start(update: Update, ctx: CallbackContext):
    print_start_message(update)


def handle_timeout(update: Update, ctx: CallbackContext):
    clean_userdata(update)
    print_timeout_message(update)


def main() -> None:
    if not os.path.exists(DATA_DIR):
        os.makedirs(DATA_DIR)

    updater = Updater(BOT_TOKEN)
    dispatcher = updater.dispatcher

    conv_advanced_import = ConversationHandler(
        entry_points=[MessageHandler(Filters.regex(
            '^alsi*') & ~Filters.command, command_alsi), CommandHandler("alsi",command_alsi)],
        states={
            MANUAL_EMOJI: [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=86400,
        run_async=True
    )
    conv_import_line_sticker = ConversationHandler(
        entry_points=[CommandHandler(
            'import_line_sticker', command_import_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji),CallbackQueryHandler(parse_emoji)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title),CallbackQueryHandler(parse_title)],
            MANUAL_EMOJI: [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        run_async=True,
        conversation_timeout=86400
    )
    conv_get_animated_line_sticker = ConversationHandler(
        entry_points=[CommandHandler(
            'get_animated_line_sticker', command_get_animated_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        run_async=True,
        conversation_timeout=86400
    )
    conv_download_line_sticker = ConversationHandler(
        entry_points=[CommandHandler(
            'download_line_sticker', command_download_line_sticker)],
        states={
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        run_async=True,
        conversation_timeout=86400
    )
    conv_download_telegram_sticker = ConversationHandler(
        entry_points=[CommandHandler(
            'download_telegram_sticker', command_download_telegram_sticker)],
        states={
            GET_TG_STICKER: [MessageHandler(Filters.sticker, prepare_tg_sticker)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        run_async=True,
        conversation_timeout=86400
    )
    conv_create_sticker_set = ConversationHandler(
        entry_points=[CommandHandler(
            'create_sticker_set', command_create_sticker_set)],
        states={
            STICKER_ARCHIVE: [MessageHandler(Filters.document, parse_sticker_archive)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji),CallbackQueryHandler(parse_emoji)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title)],
            ID: [MessageHandler(Filters.text & ~Filters.command, parse_id)],
            MANUAL_EMOJI: [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        run_async=True,
        conversation_timeout=86400
    )
    # 派遣します！
    dispatcher.add_handler(conv_import_line_sticker)
    dispatcher.add_handler(conv_get_animated_line_sticker)
    dispatcher.add_handler(conv_download_line_sticker)
    dispatcher.add_handler(conv_download_telegram_sticker)
    dispatcher.add_handler(conv_advanced_import)
    dispatcher.add_handler(conv_create_sticker_set)
    dispatcher.add_handler(CommandHandler('start', command_start))
    dispatcher.add_handler(CommandHandler('help', command_start))
    dispatcher.add_handler(CommandHandler('about', command_about))
    dispatcher.add_handler(CommandHandler('faq', command_faq))
    dispatcher.add_handler(MessageHandler(
        Filters.text & ~Filters.command, handle_text_message))
    dispatcher.add_handler(MessageHandler(
        Filters.sticker, handle_sticker_message))

    updater.start_polling()
    updater.idle()


if __name__ == '__main__':
    main()
