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
from urllib.parse import urlparse
# import telegram.error
from telegram import Update, Bot, error
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
import shutil
import glob
import threading

BOT_VERSION = "5.0 RC-10"
BOT_TOKEN = os.getenv("BOT_TOKEN")
WEBHOOK_URL = os.getenv("WEBHOOK_URL")
BOT_NAME = Bot(BOT_TOKEN).get_me().username
DATA_DIR = os.path.join(BOT_NAME + "_data", "data")
HAS_DB = False

# Stages of conversations
GET_TG_STICKER, TYPE_SELECT, LINE_URL, TITLE, ID, EDIT_CHOICE, SET_EDIT, EMOJI_SELECT, USER_STICKER, EMOJI_ASSIGN = range(10)

# Line sticker types
LINE_STICKER_STATIC = "line_s"
LINE_STICKER_ANIMATION = "line_s_ani"
LINE_STICKER_POPUP = "line_s_popup"  #全螢幕
LINE_STICKER_POPUP_EFFECT = "line_s_eff" #特效
LINE_EMOJI_STATIC = "line_e"
LINE_EMOJI_ANIMATION = "line_e_ani"
LINE_STICKER_MESSAGE = "line_s_msg"  #訊息
LINE_STICKER_NAME = "line_s_name"  #隨你填

# Load other files
from database import *
from notifications import *
from helper import *

# Binary names for cross-platform purpose
FFMPEG_BIN = get_ffmpeg_bin()
MOGRIFY_BIN = get_mogrify_bin()
CONVERT_BIN = get_convert_bin()
BSDTAR_BIN = get_bsdtar_bin()


def do_auto_create_sticker_set(update: Update, ctx: CallbackContext):
    message_progress = print_import_processing(update, ctx)

    prepare_sticker_files(update, ctx)

    if not ctx.user_data['in_command'].startswith("/manage_sticker_set"):
        # Create a new sticker set using the first image.
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(update, ctx, lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=ctx.user_data['telegram_sticker_emoji'],
                                                                  webm_sticker=get_webm_sticker(ctx.user_data['telegram_sticker_files'][0])), lambda: False)
        else:
            err = retry_do(update, ctx, lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=ctx.user_data['telegram_sticker_emoji'],
                                                                  png_sticker=get_png_sticker(ctx.user_data['telegram_sticker_files'][0])), lambda: False)
        if err is not None:
            raise(err)

    for index, img in enumerate(ctx.user_data['telegram_sticker_files']):
        edit_message_progress(message_progress, ctx, index + 1, len(ctx.user_data['telegram_sticker_files']))
        # If not add sticker to set, skip the first image.
        if not ctx.user_data['in_command'].startswith("/manage_sticker_set"):
            if index == 0:
                continue
        
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(update, ctx, lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=ctx.user_data['telegram_sticker_emoji'],
                                                              webm_sticker=get_webm_sticker(img)),
                           lambda: (
                           index + 1 == len(ctx.bot.get_sticker_set(name=ctx.user_data['telegram_sticker_id']).stickers)))
        else:
            err = retry_do(update, ctx, lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=ctx.user_data['telegram_sticker_emoji'],
                                                              png_sticker=get_png_sticker(img)),
                           lambda: (
                           index + 1 == len(ctx.bot.get_sticker_set(name=ctx.user_data['telegram_sticker_id']).stickers)))
        if err is not None:
            raise(err)

    edit_message_progress(message_progress, ctx, 1, 0)
    print_sticker_done(update, ctx)


def initialize_emoji_assign(update: Update, ctx: CallbackContext):
    print_import_processing(update, ctx)
    prepare_sticker_files(update, ctx)
    # This is the FIRST sticker.
    ctx.user_data['telegram_sticker_emoji_assign_index'] = 0
    print_ask_emoji_for_single_sticker(update, ctx)


# EMOJI_ASSIGN
def parse_emoji_assign(update: Update, ctx: CallbackContext) -> int:
    # Verify emoji.
    if not hasattr(update.message, 'text'):
        update.effective_chat.send_message("Please send emoji! Try again")
        return EMOJI_ASSIGN
    em = ''.join(e for e in re.findall(
        emoji.get_emoji_regexp(), update.message.text))
    if em == '':
        update.effective_chat.send_message("Please send emoji! Try again")
        return EMOJI_ASSIGN

    # First sticker to create new set.
    if ctx.user_data['telegram_sticker_emoji_assign_index'] == 0 and not ctx.user_data['in_command'].startswith("/manage_sticker_set"):
        # Create a new sticker set using the first image.:
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(update, ctx, lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=em,
                                                                  webm_sticker=get_webm_sticker(
                                                                      ctx.user_data['telegram_sticker_files'][0])
                                                                  ), lambda: False)
        else:
            err = retry_do(update, ctx, lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=em,
                                                                  png_sticker=get_png_sticker(
                                                                      ctx.user_data['telegram_sticker_files'][0])
                                                                  ), lambda: False)
    else:
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(update, ctx, lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=em,
                                                              webm_sticker=get_webm_sticker(
                                                                  ctx.user_data['telegram_sticker_files'][ctx.user_data['telegram_sticker_emoji_assign_index']])
                                                              ),
                           lambda: (ctx.user_data['telegram_sticker_emoji_assign_index'] + 1 == len(ctx.bot.get_sticker_set(
                               name=ctx.user_data['telegram_sticker_id']).stickers)))
        else:
            err = retry_do(update, ctx, lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=em,
                                                              png_sticker=get_png_sticker(
                                                                  ctx.user_data['telegram_sticker_files'][ctx.user_data['telegram_sticker_emoji_assign_index']])
                                                              ),
                           lambda: (ctx.user_data['telegram_sticker_emoji_assign_index'] + 1 == len(ctx.bot.get_sticker_set(
                               name=ctx.user_data['telegram_sticker_id']).stickers)))
    if err is not None:
        if err is telegram.error.BadRequest:
            if "Stickers_too_much" in err.message:
                print_sticker_full(update)
                print_sticker_done(update, ctx)
            else:
                print_fatal_error(update, traceback.format_exc())
        else:
            print_fatal_error(update, traceback.format_exc())
        clean_userdata(update, ctx)
        return ConversationHandler.END

    if ctx.user_data['telegram_sticker_emoji_assign_index'] == len(ctx.user_data['telegram_sticker_files']) - 1:
        if ctx.user_data['in_command'].startswith("/import_line_sticker"):
            insert_line_and_tg_id(ctx.user_data['line_sticker_id'], ctx.user_data['telegram_sticker_id'], ctx.user_data['telegram_sticker_title'])
        print_sticker_done(update, ctx)
        print_command_done(update, ctx)
        clean_userdata(update, ctx)
        return ConversationHandler.END

    ctx.user_data['telegram_sticker_emoji_assign_index'] += 1
    print_ask_emoji_for_single_sticker(update, ctx)
    return EMOJI_ASSIGN


# ID
# Only /create_sticker_set and /manage_sticker_set will hit this expression.
def parse_id(update: Update, ctx: CallbackContext) -> int:
    # /create_sticker_set
    if str(ctx.user_data['in_command']).startswith("/create_sticker_set"):
        if update.callback_query is not None and update.callback_query.data == "auto":
            ctx.user_data['telegram_sticker_id'] = f"sticker_" + \
                f"{secrets.token_hex(nbytes=4)}_by_{BOT_NAME}"
            update.callback_query.answer()
            edit_inline_kb_auto_selected(update.callback_query)
        elif update.message.text is not None:
            ctx.user_data['telegram_sticker_id'] = update.message.text.strip(
            ) + "_by_" + BOT_NAME
            if not re.match(r'^[a-zA-Z0-9_]+$', ctx.user_data['telegram_sticker_id']) or \
                    len(ctx.user_data['telegram_sticker_id']) > 64:
                print_wrong_id_syntax(update)
                return ID

            try:
                sticker_set = ctx.bot.get_sticker_set(
                    ctx.user_data['telegram_sticker_id'])
            except:
                pass
            else:
                update.effective_chat.send_message(
                    "set already exists! try again")
                return ID
        else:
            update.effective_chat.send_message("Please send ID.")
            return ID
        print_ask_user_sticker(update, ctx)
        return USER_STICKER
    # /manage_sticker_set
    else:
        if update.message.sticker is not None:
            ctx.user_data['telegram_sticker_id'] = update.message.sticker.set_name
        else:
            m = update.message.text.strip()
            if "/" in m:
                m = m.split('/')[-1]
            ctx.user_data['telegram_sticker_id'] = m
        if not ctx.user_data['telegram_sticker_id'].endswith(BOT_NAME):
            update.effective_chat.send_message("set not created by this bot! Try another one.")
            return ID
        try:
            sticker_set = ctx.bot.get_sticker_set(
                ctx.user_data['telegram_sticker_id'])
            ctx.user_data['telegram_sticker_is_animated'] = sticker_set.is_video
        except:
            update.effective_chat.send_message("No such set! try again.")
            return ID
        print_ask_what_to_edit(update)
        return EDIT_CHOICE


# TITLE
def parse_title(update: Update, ctx: CallbackContext) -> int:
    # /create_sticker_set
    if str(ctx.user_data['in_command']).startswith("/create_sticker_set"):
        ctx.user_data['telegram_sticker_title'] = update.message.text.strip()
        print_ask_id(update)
        return ID

    # importing LINE stickers.
    # Auto ID generation.
    ctx.user_data['telegram_sticker_id'] = ctx.user_data['line_sticker_type'] + \
        f"_{ctx.user_data['line_sticker_id']}_" \
        f"{secrets.token_hex(nbytes=3)}_by_{BOT_NAME}"

    if update.callback_query is None:
        ctx.user_data['telegram_sticker_title'] = update.message.text.strip()
    elif update.callback_query.data == "auto":
        update.callback_query.answer()
        edit_inline_kb_auto_selected(update.callback_query)
        pass
    else:
        return TITLE

    print_ask_emoji(update)
    return EMOJI_SELECT


# EMOJI_SELECT
def parse_emoji(update: Update, ctx: CallbackContext) -> int:
    if update.callback_query is not None:
        if update.callback_query.data == "manual":
            update.callback_query.answer()
            edit_inline_kb_manual_selected(update.callback_query)
            try:
                initialize_emoji_assign(update, ctx)
            except:
                print_fatal_error(update, traceback.format_exc())
                clean_userdata(update, ctx)
                return ConversationHandler.END
            return EMOJI_ASSIGN
        elif update.callback_query.data == "random":
            ctx.user_data['telegram_sticker_emoji'] = "⭐️"
            update.callback_query.answer()
            edit_inline_kb_random_selected(update.callback_query)
        else:
            return EMOJI_SELECT
    else:
        emojis = ''.join(e for e in re.findall(
            emoji.get_emoji_regexp(), update.message.text))
        if emojis == '':
            update.message.reply_text("Please send emoji! Try again")
            return EMOJI_SELECT
        ctx.user_data['telegram_sticker_emoji'] = emojis

    try:
        do_auto_create_sticker_set(update, ctx)
    except telegram.error.BadRequest as br:
        if "Stickers_too_much" in br.message:
            print_sticker_full(update)
        else:
            print_fatal_error(update, traceback.format_exc())
    except:
        print_fatal_error(update, traceback.format_exc())

    clean_userdata(update, ctx)
    return ConversationHandler.END
    

# LINE_URL
def parse_line_url(update: Update, ctx: CallbackContext) -> int:
    try:
        get_line_sticker_detail(update.message.text, ctx)
    except Exception as e:
        print_wrong_LINE_STORE_URL(update, str(e))
        return LINE_URL
    if ctx.user_data['in_command'].startswith("/download_line_sticker"):
        update.message.reply_text(ctx.user_data['line_sticker_download_url'])
        clean_userdata(update, ctx)
        return ConversationHandler.END

    # Generate auto title with bot name as suffix.
    ctx.user_data['telegram_sticker_title'] = BeautifulSoup(ctx.user_data['line_store_webpage'].text, 'html.parser')\
        .find("title").get_text().split('|')[0].strip().split('LINE')[0][:-3] + \
        f" @{BOT_NAME}"
    ctx.user_data['telegram_sticker_is_animated'] = ctx.user_data['line_sticker_is_animated']

    queried_tg_id = query_tg_id_by_line_id(ctx.user_data['line_sticker_id'])
    if queried_tg_id is not None:
        print_notify_line_sticker_set_exists(update, ctx, queried_tg_id)
        
    print_ask_title(update, ctx.user_data['telegram_sticker_title'])
    return TITLE


def get_line_sticker_detail(message, ctx: CallbackContext):
    message_url = re.findall(r'\b(?:https?):[\w/#~:.?+=&%@!\-.:?\\-]+?(?=[.:?\-]*(?:[^\w/#~:.?+=&%@!\-.:?\-]|$))',
                             message)[0]
    request_header = {'User-Agent': "curl/7.61.1"}
    webpage = requests.get(message_url, headers=request_header)
    ctx.user_data['line_store_webpage'] = webpage
    if not webpage.url.startswith("https://store.line.me"):
        raise Exception("Not a LINE Store link! 不是LINE商店連結!")
    json_details = json.loads(BeautifulSoup(
        webpage.text, "html.parser").find_all('script')[0].contents[0])
    i = json_details['sku']
    url = json_details['url']
    url_comps = urlparse(url).path[1:].split('/')
    is_animated = False
    if url_comps[0] == "stickershop":
        # First one matches AnimatedSticker with NO sound and second one with sound.
        if 'MdIcoPlay_b' in webpage.text or 'MdIcoAni_b' in webpage.text:
            t = LINE_STICKER_ANIMATION
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + \
                i + "/iphone/stickerpack@2x.zip"
            is_animated = True
        elif 'MdIcoMessageSticker_b' in webpage.text:
            t = LINE_STICKER_MESSAGE
            u = webpage.url
        elif 'MdIcoNameSticker_b' in webpage.text:
            t = LINE_STICKER_NAME
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + \
                i + "/iphone/sticker_name_base@2x.zip"
        elif 'MdIcoFlash_b' in webpage.text or 'MdIcoFlashAni_b' in webpage.text:
            t = LINE_STICKER_POPUP
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + \
                i + "/iphone/stickerpack@2x.zip"
            is_animated = True
        elif 'MdIcoEffectSoundSticker_b' in webpage.text or 'MdIcoEffectSticker_b' in webpage.text:
            t = LINE_STICKER_POPUP_EFFECT
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + \
                i + "/iphone/stickerpack@2x.zip"
            is_animated = True
        else:
            t = LINE_STICKER_STATIC
            u = "https://stickershop.line-scdn.net/stickershop/v1/product/" + \
                i + "/iphone/stickers@2x.zip"
    elif url_comps[0] == "emojishop":
        if 'MdIcoPlay_b' in webpage.text:
            t = LINE_EMOJI_ANIMATION
            u = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + \
                i + "/iphone/package_animation.zip"
            is_animated = True
        else:
            t = LINE_EMOJI_STATIC
            u = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + \
                i + "/iphone/package.zip"
    else:
        raise Exception("Not a supported sticker type!\nLink is: " + url)

    ctx.user_data['line_sticker_url'] = url
    ctx.user_data['line_sticker_type'] = t
    ctx.user_data['line_sticker_id'] = i
    ctx.user_data['line_sticker_download_url'] = u
    ctx.user_data['line_sticker_is_animated'] = is_animated


def command_import_line_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_line_store_link(update)
    return LINE_URL


def command_download_line_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_line_store_link(update)
    return LINE_URL


def command_get_animated_line_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_line_store_link(update)
    return LINE_URL


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
        initialize_emoji_assign(update, ctx)
        return EMOJI_ASSIGN
    else:
        ctx.user_data['telegram_sticker_emoji'] = alsi_args.emoji
        try:
            do_auto_create_sticker_set(update, ctx)
        except:
            clean_userdata(update, ctx)
            print_fatal_error(update, traceback.format_exc())
        return ConversationHandler.END


# GET_TG_STICKER
def get_tg_sticker(update: Update, ctx: CallbackContext) -> int:
    sticker_set = ctx.bot.get_sticker_set(name=update.message.sticker.set_name)
    ctx.user_data['telegram_sticker_id'] = sticker_set.name
    ctx.user_data['telegram_sticker_title'] = sticker_set.title
    sticker_dir = os.path.join(
        DATA_DIR, str(update.effective_user.id), sticker_set.name)
    os.makedirs(sticker_dir, exist_ok=True)

    message_progress = print_import_processing(update, ctx)

    if sticker_set.is_animated:
        sticker_suffix = ".tgs"
    elif sticker_set.is_video:
        sticker_suffix = ".webm"
    else:
        sticker_suffix = ".webp"

    for index, sticker in enumerate(sticker_set.stickers):
        try:
            edit_message_progress(message_progress, ctx, index+1, len(sticker_set.stickers))
            ctx.bot.get_file(sticker.file_id).download(os.path.join(sticker_dir,
                                                                    sticker.set_name +
                                                                    "_" + str(index).zfill(3) + "_" +
                                                                    emoji.demojize(sticker.emoji)[1:-1] +
                                                                    sticker_suffix))
        except Exception as e:
            print_fatal_error(update, traceback.format_exc())
            clean_userdata(update, ctx)
            return ConversationHandler.END
    webp_zip = os.path.join(sticker_dir, sticker_set.name + "_webp.zip")
    webm_zip = os.path.join(sticker_dir, sticker_set.name + "_webm.zip")
    tgs_zip = os.path.join(sticker_dir, sticker_set.name + "_tgs.zip")
    png_zip = os.path.join(sticker_dir, sticker_set.name + "_png.zip")
    try:
        if sticker_set.is_animated:
            subprocess.run(BSDTAR_BIN + ["--strip-components", "3", "-acvf", tgs_zip] +
                           glob.glob(os.path.join(sticker_dir, "*.tgs")))
            update.effective_chat.send_document(open(tgs_zip, 'rb'))
        elif sticker_set.is_video:
            subprocess.run(BSDTAR_BIN + ["--strip-components", "3", "-acvf", webm_zip] +
                           glob.glob(os.path.join(sticker_dir, "*.webm")))
            update.effective_chat.send_document(open(webm_zip, 'rb'))
        else:
            subprocess.run(MOGRIFY_BIN + ["-format", "png"] +
                           glob.glob(os.path.join(sticker_dir, "*.webp")))
            subprocess.run(BSDTAR_BIN + ["--strip-components", "3", "-acvf", png_zip] +
                           glob.glob(os.path.join(sticker_dir, "*.png")))
            subprocess.run(BSDTAR_BIN + ["--strip-components", "3", "-acvf", webp_zip] +
                           glob.glob(os.path.join(sticker_dir, "*.webp")))
            update.effective_chat.send_document(open(webp_zip, 'rb'))
            update.effective_chat.send_document(open(png_zip, 'rb'))

    except Exception as e:
        print_fatal_error(update, traceback.format_exc())
        clean_userdata(update, ctx)
        return ConversationHandler.END

    edit_message_progress(message_progress, ctx, 1, 0)
    clean_userdata(update, ctx)
    return ConversationHandler.END


# USER_STICKER
def parse_user_sticker(update: Update, ctx: CallbackContext) -> int:
    work_dir = os.path.join(DATA_DIR, str(update.effective_user.id))
    os.makedirs(work_dir, exist_ok=True)
    media_file_path = os.path.join(
        work_dir, secrets.token_hex(nbytes=4) + ".media")
    if not verify_user_sticker_message(update):
        print_ask_user_sticker(update, ctx)
        return USER_STICKER
    if update.message.document is not None:
        if update.message.media_group_id is not None:
            print_do_not_send_media_group(update, ctx)
            return USER_STICKER
        if update.message.document.file_size > 20 * 1024 * 1000:
            print_file_too_big(update)
        if guess_file_is_archive(update.message.document.file_name):
            # libarchive is smart enough to recognize actual archive format.
            if len(ctx.user_data['user_sticker_files']) > 0:
                update.message.reply_text(
                    "Do not send archive after sending images! skipping...")
                return USER_STICKER
            archive_file_path = media_file_path.replace(".media", ".archive")
            queued_download(update.message.document.get_file(), archive_file_path, ctx)
            ctx.user_data['user_sticker_archive'] = archive_file_path
            print_sticker_archive_received(update, ctx)
            print_ask_emoji(update)
            return EMOJI_SELECT
        else:
            # ImageMagick and ffmpeg are smart enough to recognize actual image format.
            queued_download(update.message.document.get_file(), media_file_path, ctx)
            ctx.user_data['user_sticker_files'].append(media_file_path)
            return USER_STICKER
    # Compressed image.
    elif len(update.message.photo) > 0:
        if update.message.media_group_id is not None:
            print_do_not_send_media_group(update, ctx)
            return USER_STICKER
        queued_download(update.message.photo[-1].get_file(), media_file_path, ctx)
        ctx.user_data['user_sticker_files'].append(media_file_path)
        return USER_STICKER

    elif update.message.video is not None:
        if update.message.video.file_size > 20 * 1024 * 1000:
            print_file_too_big(update)
            return USER_STICKER
        queued_download(update.message.video.get_file(), media_file_path, ctx)
        ctx.user_data['user_sticker_files'].append(media_file_path)
        return USER_STICKER
    # Telegram sticker.
    elif update.message.sticker is not None:
        if ctx.user_data['telegram_sticker_is_animated'] is False and update.message.sticker.is_video is False and update.message.sticker.is_animated is False:
            # If you send .webp image on PC, API will recognize it as a sticker
            # However, that "sticker" may not fufill requirements.
            if (update.message.sticker.width != 512 and update.message.sticker.height != 512):
                queued_download(update.message.sticker.get_file(), media_file_path, ctx)
                ctx.user_data['user_sticker_files'].append(media_file_path)
            else:
                sticker_file_id = update.message.sticker.file_id
                ctx.user_data['user_sticker_files'].append(sticker_file_id)
        else:
            update.effective_chat.send_message(
                reply_to_message_id=update.message.message_id, text="wrong sticker type! skipping this one")
        return USER_STICKER

    elif "done" in update.message.text.lower():
        if len(ctx.user_data['user_sticker_files']) == 0:
            print_no_user_sticker_received(update)
            return USER_STICKER

        wait_download_queue(update, ctx)
        print_user_sticker_done(update, ctx)
        print_ask_emoji(update)
        return EMOJI_SELECT
    else:
        update.effective_chat.send_message("Please send done.")


def parse_type(update: Update, ctx: CallbackContext):
    if update.callback_query is not None:
        if update.callback_query.data == "animated":
            ctx.user_data['telegram_sticker_is_animated'] = True
        elif update.callback_query.data == "static":
            ctx.user_data['telegram_sticker_is_animated'] = False
        else:
            return TYPE_SELECT

        update.callback_query.answer()
        print_ask_title(update, "")
        return TITLE
    else:
        return TYPE_SELECT


def parse_edit_choice(update: Update, ctx: CallbackContext):
    if update.callback_query is None:
        update.effective_chat.send_message("Please select command above.")
        return EDIT_CHOICE
    elif update.callback_query.data == "add":
        ctx.user_data['telegram_sticker_edit_choice'] = "add"
        update.callback_query.answer()
        print_ask_user_sticker(update, ctx)
        return USER_STICKER
    elif update.callback_query.data == "del":
        ctx.user_data['telegram_sticker_edit_choice'] = "del"
        update.callback_query.answer()
        print_ask_which_to_delete(update)
        return SET_EDIT
    elif update.callback_query.data == "mov":
        ctx.user_data['telegram_sticker_edit_choice'] = "mov"
        update.callback_query.answer()
        print_ask_which_to_move(update)
        return SET_EDIT
    else:
        return EDIT_CHOICE


def parse_set_edit(update: Update, ctx: CallbackContext):
    if update.message.sticker is None:
        update.effective_chat.send_message("please send sticker!")
        return SET_EDIT
    s = update.message.sticker
    if s.set_name != ctx.user_data['telegram_sticker_id']:
        update.effective_chat.send_message("sticker from wrong set! try again.")
        return SET_EDIT
    if ctx.user_data['telegram_sticker_edit_choice'] == "del":
        try:
            ctx.bot.delete_sticker_from_set(s.file_id)
        except Exception as e:
            update.effective_chat.send_message("Failed to delete! Try again.\n" + str(e))
            return SET_EDIT
    elif ctx.user_data['telegram_sticker_edit_choice'] == "mov":
        # Receiving FIRST sticker.
        if ctx.user_data['telegram_sticker_edit_mov_prev'] is None:
            ctx.user_data['telegram_sticker_edit_mov_prev'] = s
            print_ask_where_to_move(update)
            return SET_EDIT
        # Receiving SECOND sticker.
        else:
            try:
                sticker_set = ctx.bot.get_sticker_set(s.set_name)
                dest = sticker_set.stickers.index(s)
                ctx.bot.set_sticker_position_in_set(ctx.user_data['telegram_sticker_edit_mov_prev'].file_id, dest)
            except Exception as e:
                update.effective_chat.send_message("Failed to move! Try again.\n" + str(e))
                ctx.user_data['telegram_sticker_edit_mov_prev'] = None
                print_ask_which_to_move(update)
                return SET_EDIT
    else:
        pass
    print_command_done(update, ctx)
    clean_userdata(update, ctx)
    return ConversationHandler.END


def command_download_telegram_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_telegram_sticker(update)
    return GET_TG_STICKER


def command_create_sticker_set(update: Update, ctx: CallbackContext) -> int:
    initialize_user_data(update, ctx)
    print_ask_type_to_create(update)
    return TYPE_SELECT


def command_manage_sticker_set(update: Update, ctx: CallbackContext) -> int:
    initialize_user_data(update, ctx)
    print_ask_sticker_set(update)
    return ID


def command_cancel(update: Update, ctx: CallbackContext) -> int:
    clean_userdata(update, ctx)
    print_command_canceled(update)
    command_start(update, ctx)
    return ConversationHandler.END


def handle_text_message(update: Update, ctx: CallbackContext):
    # PTB has some weired bugs that triggers here
    # even during a conversation.
    # At least some workarounds.
    if update.message.text == "done":
        update.effective_chat.send_message(
            'Please send "done" again. 請再傳送一次 done', reply_markup=reply_kb_DONE)
        return
    if 'in_command' in ctx.user_data:
        print_in_conv_warning(update, ctx)
    else:
        print_use_start_command(update)


def handle_sticker_message(update: Update, ctx: CallbackContext):
    if 'in_command' in ctx.user_data:
        print_in_conv_warning(update, ctx)
    else:
        print_use_start_command(update)
        print_suggest_download(update)


def command_about(update: Update, ctx: CallbackContext):
    print_about_message(update, BOT_NAME, BOT_VERSION)


def command_faq(update: Update, ctx: CallbackContext):
    print_faq_message(update)


def command_start(update: Update, ctx: CallbackContext):
    print_start_message(update)


def handle_timeout(update: Update, ctx: CallbackContext):
    clean_userdata(update, ctx)
    print_timeout_message(update)


def main() -> None:
    if not os.path.exists(DATA_DIR):
        os.makedirs(DATA_DIR)

    updater = Updater(BOT_TOKEN)
    dispatcher = updater.dispatcher

    conv_advanced_import = ConversationHandler(
        entry_points=[MessageHandler(Filters.regex(
            '^alsi*') & ~Filters.command, command_alsi), CommandHandler("alsi", command_alsi)],
        states={
            EMOJI_ASSIGN: [MessageHandler(Filters.text & ~Filters.command, parse_emoji_assign)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=43200,
        run_async=True
    )
    conv_import_line_sticker = ConversationHandler(
        entry_points=[CommandHandler(
            'import_line_sticker', command_import_line_sticker)],
        states={
            LINE_URL: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title), CallbackQueryHandler(parse_title)],
            EMOJI_SELECT: [MessageHandler(Filters.text & ~Filters.command, parse_emoji), CallbackQueryHandler(parse_emoji)],
            EMOJI_ASSIGN: [MessageHandler(Filters.text & ~Filters.command, parse_emoji_assign)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=43200,
        run_async=True
    )
    conv_download_line_sticker = ConversationHandler(
        entry_points=[CommandHandler(
            'download_line_sticker', command_download_line_sticker)],
        states={
            LINE_URL: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)],
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=43200,
        run_async=True
    )
    conv_download_telegram_sticker = ConversationHandler(
        entry_points=[CommandHandler(
            'download_telegram_sticker', command_download_telegram_sticker)],
        states={
            GET_TG_STICKER: [MessageHandler(Filters.sticker, get_tg_sticker)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=43200,
        run_async=True
    )
    conv_create_sticker_set = ConversationHandler(
        entry_points=[CommandHandler(
            'create_sticker_set', command_create_sticker_set)],
        states={
            TYPE_SELECT: [CallbackQueryHandler(parse_type)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title)],
            ID: [MessageHandler(Filters.text & ~Filters.command, parse_id), CallbackQueryHandler(parse_id)],
            USER_STICKER: [MessageHandler(Filters.all & ~Filters.command, parse_user_sticker)],
            EMOJI_SELECT: [MessageHandler(Filters.text & ~Filters.command, parse_emoji), CallbackQueryHandler(parse_emoji)],
            EMOJI_ASSIGN: [MessageHandler(Filters.text & ~Filters.command, parse_emoji_assign)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=43200,
        run_async=True
    )
    conv_manage_sticker_set = ConversationHandler(
        entry_points=[CommandHandler(
            'manage_sticker_set', command_manage_sticker_set)],
        states={
            ID:  [MessageHandler(((Filters.text & ~Filters.command) | Filters.sticker), parse_id)],
            EDIT_CHOICE: [CallbackQueryHandler(parse_edit_choice)],
            SET_EDIT: [MessageHandler((Filters.sticker), parse_set_edit)],
            USER_STICKER: [MessageHandler(Filters.all & ~Filters.command, parse_user_sticker)],
            EMOJI_SELECT: [MessageHandler(Filters.text & ~Filters.command, parse_emoji), CallbackQueryHandler(parse_emoji)],
            EMOJI_ASSIGN: [MessageHandler(Filters.text & ~Filters.command, parse_emoji_assign)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=43200,
        run_async=True
    )
    # 派遣します！
    dispatcher.add_handler(conv_import_line_sticker)
    dispatcher.add_handler(conv_download_line_sticker)
    dispatcher.add_handler(conv_download_telegram_sticker)
    dispatcher.add_handler(conv_advanced_import)
    dispatcher.add_handler(conv_create_sticker_set)
    dispatcher.add_handler(conv_manage_sticker_set)
    dispatcher.add_handler(CommandHandler('start', command_start))
    dispatcher.add_handler(CommandHandler('help', command_start))
    dispatcher.add_handler(CommandHandler('about', command_about))
    dispatcher.add_handler(CommandHandler('faq', command_faq))
    dispatcher.add_handler(MessageHandler(
        Filters.text & ~Filters.command, handle_text_message))
    dispatcher.add_handler(MessageHandler(
        Filters.sticker, handle_sticker_message))

    start_timer_userdata_gc()
    if attempt_connect_to_mariadb():
        global HAS_DB
        HAS_DB = True

    if WEBHOOK_URL is not None:
        threading.Timer(10, delayed_set_webhook).start()
        updater.start_webhook(listen='0.0.0.0', port=443, url_path=BOT_TOKEN, key='/privkey.pem', cert='/fullchain.pem', webhook_url=WEBHOOK_URL + BOT_TOKEN)
        # Fix PTB's weired SSL: TLSV1_ALERT_UNKNOWN_CA problem.
    else:
        updater.start_polling()

    updater.idle()


if __name__ == '__main__':
    main()
