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
from telegram import Update, Bot
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

from notifications import *
from helper import *


BOT_VERSION = "4.0 ALPHA-6"
BOT_TOKEN = os.getenv("BOT_TOKEN")
BOT_NAME = Bot(BOT_TOKEN).get_me().username
DATA_DIR = os.path.join(BOT_NAME + "_data", "data")

# Stages of conversations
SELECT_TYPE, LINE_STICKER_INFO, TITLE, ID, EMOJI, USER_STICKER, MANUAL_EMOJI = range(
    7)
# USER_STICKER, EMOJI, TITLE, ID, MANUAL_EMOJI = range(5)
GET_TG_STICKER = range(1)

# Line sticker types
LINE_STICKER_STATIC = "line_s"
LINE_STICKER_ANIMATION = "line_s_ani"
LINE_STICKER_EFFECT_ANIMATION = "line_s_sfxani"
LINE_EMOJI_STATIC = "line_e"
LINE_EMOJI_ANIMATION = "line_e_ani"
LINE_STICKER_MESSAGE = "line_s_msg"

FFMPEG_BIN = get_ffmpeg_bin()
MOGRIFY_BIN = get_mogrify_bin()
CONVERT_BIN = get_convert_bin()
BSDTAR_BIN = get_bsdtar_bin()



def do_auto_create_sticker_set(update: Update, ctx: CallbackContext):
    message_progress = print_import_processing(update, ctx)

    if ctx.user_data['in_command'].startswith("/create_sticker_set") or ctx.user_data['in_command'].startswith("/add_sticker_to_set"):
        img_files_path = prepare_user_sticker_files(
            update, ctx, ctx.user_data['telegram_sticker_is_animated'])
    else:
        img_files_path = prepare_line_sticker_files(update, ctx)

    if not ctx.user_data['in_command'].startswith("/add_sticker_to_set"):
        # Create a new sticker set using the first image.
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=ctx.user_data['telegram_sticker_emoji'],
                                                                  webm_sticker=get_webm_sticker(img_files_path[0])), lambda: False)
        else:
            err = retry_do(lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=ctx.user_data['telegram_sticker_emoji'],
                                                                  png_sticker=get_png_sticker(img_files_path[0])), lambda: False)
        if err is not None:
            print_fatal_error(update, str(err))
            return

    edit_message_progress(message_progress, 1, len(img_files_path))
    for index, img_file_path in enumerate(img_files_path):
        # If not add sticker to set, skip the first image.
        if not ctx.user_data['in_command'].startswith("/add_sticker_to_set"):
            if index == 0:
                continue
        edit_message_progress(message_progress, index + 1, len(img_files_path))
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=ctx.user_data['telegram_sticker_emoji'],
                                                              webm_sticker=get_webm_sticker(img_file_path)),
                           lambda: (
                           index + 1 == ctx.bot.get_sticker_set(name=ctx.user_data['telegram_sticker_id']).stickers))
        else:
            err = retry_do(lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=ctx.user_data['telegram_sticker_emoji'],
                                                              png_sticker=get_png_sticker(img_file_path)),
                           lambda: (
                           index + 1 == ctx.bot.get_sticker_set(name=ctx.user_data['telegram_sticker_id']).stickers))
        if err is not None:
            print_fatal_error(update, str(err))
            return

    print_sticker_done(update, ctx)
    print_command_done(update, ctx)


def prepare_user_sticker_files(update: Update, ctx, want_animated):
    images_path = []
    # animated, convert to WEBM VP9
    if want_animated:
        for f in ctx.user_data['user_sticker_files']:
            if f.endswith('.media'):
                ret = subprocess.run(FFMPEG_BIN + ["-hide_banner", "-loglevel", "error", "-i", f,
                                      "-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
                                      "-c:v", "libvpx-vp9", "-cpu-used", "3", "-minrate", "50k", "-b:v", "400k", "-maxrate", "500k",
                                      "-to", "00:00:02.800", "-an",
                                      f + '.webm'], capture_output=True)
                # Skip errored conversion. Don't panic.
                # That may ruin user experience.
                if ret.returncode == 0:
                    images_path.append(f + '.webm')
                else:
                    update.effective_chat.send_message(
                        "WARN: failed processing one media.\n\n" + str(ret.stderr))
            else:
                images_path.append(f)
    # static, convert to webp
    else:
        for f in ctx.user_data['user_sticker_files']:
            if f.endswith('.media'):
                ret = subprocess.run(MOGRIFY_BIN + ["-background", "none", "-filter", "Lanczos", "-resize", "512x512",
                                      "-format", "webp", "-define", "webp:lossless=true", f + "[0]"], capture_output=True)
                if ret.returncode == 0:
                    images_path.append(f.replace('.media', '.webp'))
                else:
                    update.effective_chat.send_message(
                        "WARN: failed processing one media.\n\n" + str(ret.stderr))
            else:
                images_path.append(f)
    return images_path


def prepare_line_sticker_files(update: Update, ctx: CallbackContext):
    work_dir = os.path.join(
        DATA_DIR, str(update.effective_user.id), ctx.user_data['line_sticker_id'])
    os.makedirs(work_dir, exist_ok=True)
    # Special line "message" stickers
    if ctx.user_data['line_sticker_type'] == LINE_STICKER_MESSAGE:
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
                    ["curl", "-Lo", f"{work_dir}{image_id}.base.png", base_image])
                subprocess.run(
                    ["curl", "-Lo", f"{work_dir}{image_id}.overlay.png", overlay_image])
                subprocess.run(CONVERT_BIN + [f"{work_dir}{image_id}.base.png", f"{work_dir}{image_id}.overlay.png",
                                "-background", "none", "-filter", "Lanczos", "-resize", "512x512", "-composite",
                                "-define", "webp:lossless=true",
                                f"{work_dir}{image_id}.webp"])

    else:
        zip_file_path = os.path.join(
            work_dir, ctx.user_data['line_sticker_id'] + ".zip")
        subprocess.run(["curl", "-Lo", zip_file_path,
                        ctx.user_data['line_sticker_download_url']])
        subprocess.run(BSDTAR_BIN + ["-xf", zip_file_path, "-C", work_dir])
        for f in glob.glob(os.path.join(work_dir, "*key*")) + glob.glob(os.path.join(work_dir, "tab*")) + glob.glob(os.path.join(work_dir, "productInfo.meta")):
            os.remove(f)
        # standard static line stickers.
        if ctx.user_data['line_sticker_type'] == LINE_STICKER_STATIC or ctx.user_data['line_sticker_type'] == LINE_EMOJI_STATIC:
            # Resize to fulfill telegram's requirement, AR is automatically retained
            # Lanczos resizing produces much sharper image.
            for f in glob.glob(os.path.join(work_dir, "*")):
                subprocess.run(MOGRIFY_BIN + ["-background", "none", "-filter", "Lanczos", "-resize", "512x512",
                                "-format", "webp", "-define", "webp:lossless=true", f])
        # is animated line stickers/emojis.
        else:
            if ctx.user_data['line_sticker_type'] == LINE_STICKER_ANIMATION:
                work_dir = os.path.join(work_dir, "animation@2x")
            elif ctx.user_data['line_sticker_type'] == LINE_STICKER_EFFECT_ANIMATION:
                for f in glob.glob(os.path.join(work_dir, "popup", "*.png")):
                    # workaround for sticker orders.
                    shutil.move(f, os.path.join(work_dir, os.path.basename(
                        f)[:os.path.basename(f).index('.png')] + '@99x.png'))
            # LINE_EMOJI_ANIMATION
            else:
                pass
            for f in glob.glob(os.path.join(work_dir, "**", "*.png"), recursive=True):
                subprocess.run(FFMPEG_BIN + ["-hide_banner", "-loglevel", "error", "-i", f,
                                "-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
                                "-c:v", "libvpx-vp9", "-cpu-used", "3", "-minrate", "50k", "-b:v", "400k", "-maxrate", "500k",
                                "-to", "00:00:02.800", "-an",
                                f + '.webm'])

            return sorted([f for f in glob.glob(os.path.join(work_dir, "*.webm"))])

    return sorted([f for f in glob.glob(os.path.join(work_dir, "**", "*.webp"), recursive=True)])


def initialize_manual_emoji(update: Update, ctx: CallbackContext):
    print_import_processing(update, ctx)
    if ctx.user_data['in_command'].startswith("/create_sticker_set") or ctx.user_data['in_command'].startswith("/add_sticker_to_set"):
        ctx.user_data['img_files_path'] = prepare_user_sticker_files(
            update, ctx, ctx.user_data['telegram_sticker_is_animated'])
    else:
        ctx.user_data['img_files_path'] = prepare_line_sticker_files(
            update, ctx)
    # This is the FIRST sticker.
    ctx.user_data['manual_emoji_index'] = 0
    print_ask_emoji_for_single_sticker(update, ctx)


# MANUAL_EMOJI
def manual_add_emoji(update: Update, ctx: CallbackContext) -> int:
    # Verify emoji.
    em = ''.join(e for e in re.findall(
        emoji.get_emoji_regexp(), update.message.text))
    if em == '':
        update.message.reply_text("Please send emoji! Try again")
        return MANUAL_EMOJI

    # First sticker to create new set.
    if ctx.user_data['manual_emoji_index'] == 0 and not ctx.user_data['in_command'].startswith("/add_sticker_to_set"):
        # Create a new sticker set using the first image.:
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=em,
                                                                  webm_sticker=get_webm_sticker(
                                                                      ctx.user_data['img_files_path'][0])
                                                                  ), lambda: False)
        else:
            err = retry_do(lambda: ctx.bot.create_new_sticker_set(user_id=update.effective_user.id,
                                                                  name=ctx.user_data['telegram_sticker_id'],
                                                                  title=ctx.user_data['telegram_sticker_title'],
                                                                  emojis=em,
                                                                  png_sticker=get_png_sticker(
                                                                      ctx.user_data['img_files_path'][0])
                                                                  ), lambda: False)
        if err is not None:
            clean_userdata(update, ctx)
            print_fatal_error(update, str(err))
            return ConversationHandler.END
    else:
        if ctx.user_data['line_sticker_is_animated'] is True or ctx.user_data['telegram_sticker_is_animated'] is True:
            err = retry_do(lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=em,
                                                              webm_sticker=get_webm_sticker(
                                                                  ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']])
                                                              ),
                           lambda: (ctx.user_data['manual_emoji_index'] + 1 == ctx.bot.get_sticker_set(
                               name=ctx.user_data['telegram_sticker_id']).stickers))
        else:
            err = retry_do(lambda: ctx.bot.add_sticker_to_set(user_id=update.effective_user.id,
                                                              name=ctx.user_data['telegram_sticker_id'],
                                                              emojis=em,
                                                              png_sticker=get_png_sticker(
                                                                  ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']])
                                                              ),
                           lambda: (ctx.user_data['manual_emoji_index'] + 1 == ctx.bot.get_sticker_set(
                               name=ctx.user_data['telegram_sticker_id']).stickers))
        if err is not None:
            clean_userdata(update, ctx)
            print_fatal_error(update, str(err))
            return ConversationHandler.END

        if ctx.user_data['manual_emoji_index'] == len(ctx.user_data['img_files_path']) - 1:
            print_sticker_done(update, ctx)
            print_command_done(update, ctx)
            clean_userdata(update, ctx)
            return ConversationHandler.END

    ctx.user_data['manual_emoji_index'] += 1
    print_ask_emoji_for_single_sticker(update, ctx)
    return MANUAL_EMOJI


# ID
# Only /create_sticker_set and /add_sticker_to_set will hit this expression.
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

        elif update.message.sticker is not None:
            update.effective_chat.send_message(
                "do not send sticker, try again")
            return ID
        else:
            print_fatal_error(update, "unknown error")
            clean_userdata(update, ctx)
            return ConversationHandler.END
    # /add_sticker_to_set
    else:
        if update.message.sticker is not None:
            ctx.user_data['telegram_sticker_id'] = update.message.sticker.set_name
        else:
            m = update.message.text.strip()
            if "/" in m:
                m = m.split('/')[-1]
            ctx.user_data['telegram_sticker_id'] = m
        if not ctx.user_data['telegram_sticker_id'].endswith(BOT_NAME):
            print_wrong_id_syntax(update)
            return ID
        try:
            sticker_set = ctx.bot.get_sticker_set(
                ctx.user_data['telegram_sticker_id'])
            ctx.user_data['telegram_sticker_is_animated'] = sticker_set.is_video
        except:
            update.effective_chat.send_message("No such set! try again.")
            return ID
        if len(sticker_set.stickers) > 120:
            update.effective_chat.send_message(
                "Set already full! try another one.")
            return ID

    print_ask_user_sticker(update, ctx)
    return USER_STICKER


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
    return EMOJI


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

    if ctx.user_data['manual_emoji'] is True:
        initialize_manual_emoji(update, ctx)
        return MANUAL_EMOJI
    else:
        do_auto_create_sticker_set(update, ctx)
        clean_userdata(update, ctx)
        return ConversationHandler.END


# LINE_STICKER_INFO
def parse_line_url(update: Update, ctx: CallbackContext) -> int:
    try:
        get_line_sticker_detail(update.message.text, ctx)
    except Exception as e:
        print_wrong_LINE_STORE_URL(update, str(e))
        return LINE_STICKER_INFO
    if str(ctx.user_data['in_command']).startswith("/import_line_sticker"):
        # Generate auto title with bot name as suffix.
        ctx.user_data['telegram_sticker_title'] = BeautifulSoup(ctx.user_data['line_store_webpage'].text, 'html.parser')\
            .find("title").get_text().split('|')[0].strip().split('LINE')[0][:-3] + \
            f" @{BOT_NAME}"
        print_ask_title(update, ctx.user_data['telegram_sticker_title'])
        return TITLE
    elif str(ctx.user_data['in_command']).startswith("/download_line_sticker"):
        update.message.reply_text(ctx.user_data['line_sticker_download_url'])
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
        elif 'MdIcoEffectSoundSticker_b' in webpage.text:
            t = LINE_STICKER_EFFECT_ANIMATION
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
    if not os.path.exists(DATA_DIR):
        os.makedirs(DATA_DIR)
    clean_userdata(update, ctx)
    ctx.user_data['in_command'] = update.message.text.strip().split(' ')[0]
    ctx.user_data['manual_emoji'] = False
    ctx.user_data['line_sticker_url'] = ""
    ctx.user_data['line_store_webpage'] = None
    ctx.user_data['line_sticker_download_url'] = ""
    ctx.user_data['line_sticker_type'] = None
    ctx.user_data['line_sticker_is_animated'] = False
    ctx.user_data['line_sticker_id'] = ""
    ctx.user_data['telegram_sticker_emoji'] = ""
    ctx.user_data['telegram_sticker_id'] = ""
    ctx.user_data['telegram_sticker_title'] = ""
    ctx.user_data['telegram_sticker_is_animated'] = False
    ctx.user_data['user_sticker_archive'] = ""
    ctx.user_data['user_sticker_count'] = 0
    ctx.user_data['user_sticker_files'] = []


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
    if sticker_set.is_animated:
        sticker_suffix = ".tgs"
    elif sticker_set.is_video:
        sticker_suffix = ".webm"
    else:
        sticker_suffix = ".webp"

    for index, sticker in enumerate(sticker_set.stickers):
        try:
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

    print_command_done(update, ctx)
    clean_userdata(update, ctx)
    return ConversationHandler.END


# USER_STICKER
def parse_user_sticker(update: Update, ctx: CallbackContext) -> int:
    work_dir = os.path.join(DATA_DIR, str(update.effective_user.id))
    os.makedirs(work_dir, exist_ok=True)
    media_file_path = os.path.join(
        work_dir, secrets.token_hex(nbytes=4) + ".media")
    if update.message.document is not None:
        # Uncompressed image.
        if update.message.document.file_size > 10 * 1024 * 1024:
            update.effective_chat.send_message(
                reply_to_message_id=update.message.message_id, text="file too big! skipping this one.")
        if update.message.document.mime_type.startswith("image"):
            # ImageMagick and ffmpeg are smart enough to recognize actual image format.
            update.message.document.get_file().download(media_file_path)
            ctx.user_data['user_sticker_files'].append(media_file_path)
            return USER_STICKER
        else:
            update.effective_chat.send_message(
                reply_to_message_id=update.message.message_id, text="Unable to process this file.")
            return USER_STICKER
    # Compressed image.
    elif len(update.message.photo) > 0:
        update.message.photo[-1].get_file().download(media_file_path)
        ctx.user_data['user_sticker_files'].append(media_file_path)
        return USER_STICKER

    elif update.message.video is not None:
        if update.message.video.file_size > 10 * 1024 * 1024:
            update.effective_chat.send_message(
                reply_to_message_id=update.message.message_id, text="video too big! skipping this one. try trimming it.")
        update.message.video.get_file().download(media_file_path)
        ctx.user_data['user_sticker_files'].append(media_file_path)
        return USER_STICKER
    # Telegram sticker.
    elif update.message.sticker is not None:
        if ctx.user_data['telegram_sticker_is_animated'] is False and update.message.sticker.is_video is False and update.message.sticker.is_animated is False:
            # If you send .webp image on PC, API will recognize it as a sticker
            # However, that "sticker" may not fufill requirements.
            if update.message.sticker.width != 512 and update.message.sticker.height != 512:
                update.message.sticker.get_file().download(media_file_path)
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
        else:
            # Given 5 seconds.
            # The other threads will hopefully finish downloading images.
            time.sleep(5)
            print_user_sticker_done(update, ctx)
            print_ask_emoji(update)
            return EMOJI
    else:
        update.effective_chat.send_message("Please send done.")


def parse_type(update: Update, ctx: CallbackContext):
    if update.callback_query is not None:
        if update.callback_query.data == "animated":
            ctx.user_data['telegram_sticker_is_animated'] = True
        elif update.callback_query.data == "static":
            ctx.user_data['telegram_sticker_is_animated'] = False
        else:
            return SELECT_TYPE

        update.callback_query.answer()
        print_ask_title(update, "")
        return TITLE
    else:
        return SELECT_TYPE


def command_download_telegram_sticker(update: Update, ctx: CallbackContext):
    initialize_user_data(update, ctx)
    print_ask_telegram_sticker(update)
    return GET_TG_STICKER


def command_create_sticker_set(update: Update, ctx: CallbackContext) -> int:
    initialize_user_data(update, ctx)
    print_ask_type_to_create(update)
    return SELECT_TYPE


def command_add_sticker_to_set(update: Update, ctx: CallbackContext) -> int:
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
            'Please send "done" again. 請再傳送一次 done')
        return
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
    clean_userdata(update, ctx)
    print_timeout_message(update)


def main() -> None:
    # from telegram.utils.request import Request
    if not os.path.exists(DATA_DIR):
        os.makedirs(DATA_DIR)

    # q = mq.MessageQueue(all_burst_limit=20, all_time_limit_ms=1000)
    # req = Request(con_pool_size=8)
    # bot = MQBot(BOT_TOKEN, mqueue=q, request=req)

    updater = Updater(BOT_TOKEN)
    dispatcher = updater.dispatcher

    conv_advanced_import = ConversationHandler(
        entry_points=[MessageHandler(Filters.regex(
            '^alsi*') & ~Filters.command, command_alsi), CommandHandler("alsi", command_alsi)],
        states={
            MANUAL_EMOJI: [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
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
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title), CallbackQueryHandler(parse_title)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji), CallbackQueryHandler(parse_emoji)],
            MANUAL_EMOJI: [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
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
            LINE_STICKER_INFO: [MessageHandler(Filters.text & ~Filters.command, parse_line_url)],
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
            GET_TG_STICKER: [MessageHandler(Filters.sticker, prepare_tg_sticker)],
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
            SELECT_TYPE: [CallbackQueryHandler(parse_type)],
            TITLE: [MessageHandler(Filters.text & ~Filters.command, parse_title)],
            ID: [MessageHandler(Filters.text & ~Filters.command, parse_id), CallbackQueryHandler(parse_id)],
            USER_STICKER: [MessageHandler(Filters.all & ~Filters.command, parse_user_sticker)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji), CallbackQueryHandler(parse_emoji)],
            MANUAL_EMOJI: [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
            ConversationHandler.TIMEOUT: [MessageHandler(None, handle_timeout)]
        },
        fallbacks=[CommandHandler('cancel', command_cancel), MessageHandler(
            Filters.command, print_in_conv_warning)],
        conversation_timeout=43200,
        run_async=True
    )
    conv_add_sticker_to_set = ConversationHandler(
        entry_points=[CommandHandler(
            'add_sticker_to_set', command_add_sticker_to_set)],
        states={
            ID:  [MessageHandler(((Filters.text & ~Filters.command) | Filters.sticker), parse_id)],
            USER_STICKER: [MessageHandler(Filters.all & ~Filters.command, parse_user_sticker)],
            EMOJI: [MessageHandler(Filters.text & ~Filters.command, parse_emoji), CallbackQueryHandler(parse_emoji)],
            MANUAL_EMOJI: [MessageHandler(Filters.text & ~Filters.command, manual_add_emoji)],
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
    dispatcher.add_handler(conv_add_sticker_to_set)
    dispatcher.add_handler(CommandHandler('start', command_start))
    dispatcher.add_handler(CommandHandler('help', command_start))
    dispatcher.add_handler(CommandHandler('about', command_about))
    dispatcher.add_handler(CommandHandler('faq', command_faq))
    dispatcher.add_handler(MessageHandler(
        Filters.text & ~Filters.command, handle_text_message))
    dispatcher.add_handler(MessageHandler(
        Filters.sticker, handle_sticker_message))

    start_timer_userdata_gc()

    updater.start_polling()
    updater.idle()


if __name__ == '__main__':
    main()
