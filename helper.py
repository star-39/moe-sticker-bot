import pathlib
from urllib.parse import urlparse
import main
import subprocess
from threading import Timer
import glob
import os
import time
import shutil
import traceback
import telegram.bot
import platform
import telegram
import json
from bs4 import BeautifulSoup
from telegram import Update, File
from telegram.ext import CallbackContext
import requests
import re

from macros import *


def get_ffmpeg_bin():
    b = ['ffmpeg']
    check_bin(b[0])
    return b


def get_mogrify_bin():
    b = []
    if platform.system() == "Linux":
        b = ['mogrify']
    else:
        b = ['magick', 'mogrify']
    check_bin(b[0])
    return b


def get_convert_bin():
    b = []
    if platform.system() == "Linux":
        b = ['convert']
    else:
        b = ['magick', 'convert']
    check_bin(b[0])
    return b


def get_bsdtar_bin():
    b = []
    if platform.system() == "Linux":
        b = ['bsdtar']
    else:
        b = ['tar']
    check_bin(b[0])
    return b


def check_bin(bin: str):
    if shutil.which(bin) is None:
        # Die if deps not found.
        raise Exception(bin + " not found! Exiting...")


# Uploading sticker could easily trigger Telegram's flood limit,
# however, documentation never specified this limit,
# hence, we should at least retry after triggering the limit.
def retry_do(update: Update, ctx: CallbackContext, func, lambda_check_fake_ra):
    for index in range(5):
        try:
            func()
        except telegram.error.BadRequest as br:
            if br.message == 'Sticker_video_long':
                update.effective_chat.send_message("Failed uploading one sticker, skipped.")
                break
            else:
                return br
        except telegram.error.RetryAfter as ra:
            if index == 4:
                return ra
            time.sleep(ra.retry_after)

            if lambda_check_fake_ra():
                break
            else:
                continue

        except Exception as e:
            if index == 4:
                print(traceback.format_exc())
                return e
            time.sleep(5)
        else:
            break


# To save processing time, if a sticker already exist in Telegram's server,
# use its file id instead downloading than uploading it.
# Distinguish by judging the suffix.
def get_png_sticker(f: str):
    if f.endswith(".webp"):
        return open(f, 'rb')
    else:
        return f


def get_webm_sticker(f: str):
    if f.endswith(".webm"):
        return open(f, 'rb')
    else:
        return f


def guess_file_is_archive(f: str):
    archive_exts = ('.rar', '.7z', '.zip', '.tar',
                    '.gz', '.bz2', '.zst', '.rar5')
    if f is None:
        return False
    if f.lower().endswith(archive_exts):
        return True
    else:
        return False


def queued_download(f: File, save_path: str, ctx: CallbackContext):
    item = [f, save_path]
    ctx.user_data['user_sticker_download_queue'].append(item)


def wait_download_queue(update, ctx):
    if len(ctx.user_data['user_sticker_download_queue']) == 0:
        return
    items_count = str(len(ctx.user_data['user_sticker_download_queue']))
    update.effective_chat.send_message(f"Gathering {items_count} images/videos, please wait...\n正在取得 {items_count} 份圖片/短片, 請稍後...\n")
    time.sleep(1)

    for f, save_path in ctx.user_data['user_sticker_download_queue']:
        f.download(save_path)
        ctx.user_data['user_sticker_files'].append(save_path)


def get_kakao_emoticon_detail(url, ctx: CallbackContext):
    kakao_id = urlparse(url).path.split('/')[-1]
    api_url = 'https://e.kakao.com/api/v1/items/t/' + kakao_id

    json_data = requests.get(api_url).text
    json_details = json.loads(json_data)
    thumbnailUrls = json_details['result']['thumbnailUrls']
    title = json_details['result']['title']

    ctx.user_data['line_sticker_title'] = title
    ctx.user_data['line_sticker_image_sources'] = thumbnailUrls
    ctx.user_data['line_sticker_url'] = url
    ctx.user_data['line_sticker_type'] = KAKAO_EMOTICON
    ctx.user_data['line_sticker_id'] = kakao_id.replace('-', '_')
    ctx.user_data['line_sticker_is_animated'] = False


def get_line_sticker_detail(message, ctx: CallbackContext):
    message_url: str = re.findall(r'\b(?:https?):[\w/#~:.?+=&%@!\-.:?\\-]+?(?=[.:?\-]*(?:[^\w/#~:.?+=&%@!\-.:?\-]|$))',
                             message)[0]
    # redirect to kakao ones (workaround)
    if message_url.startswith("https://e.kakao.com/"):
        get_kakao_emoticon_detail(message_url, ctx)
        return

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

    title = BeautifulSoup(webpage.text, 'html.parser').find("title").get_text().split('|')[0].strip().split('LINE')[0][:-3]

    ctx.user_data['line_sticker_url'] = url
    ctx.user_data['line_sticker_type'] = t
    ctx.user_data['line_sticker_id'] = i
    ctx.user_data['line_sticker_download_url'] = u
    ctx.user_data['line_sticker_title'] = title
    ctx.user_data['line_sticker_is_animated'] = is_animated



def prepare_sticker_files(update: Update, ctx: CallbackContext, q):
    time_start = time.time()
    images = []
    # User stickers
    if ctx.user_data['in_command'].startswith("/create_sticker_set") or ctx.user_data['in_command'].startswith("/manage_sticker_set"):
        # User sent sticker archive
        if ctx.user_data['user_sticker_archive']:
            archive_path = ctx.user_data['user_sticker_archive']
            work_dir = os.path.dirname(archive_path)
            ret = subprocess.run(
                ['bsdtar', '-xf', archive_path, '-C', work_dir], capture_output=True)
            if ret.returncode != 0:
                raise Exception("Unable to extract image from archive!")
            os.remove(archive_path)
            for f in glob.glob(os.path.join(work_dir, "**"), recursive=True):
                if os.path.isfile(f):
                    #workaround preserving suffix
                    ext = pathlib.Path(f).suffix
                    shutil.move(f, f + ".media" + ext)
                    ctx.user_data['user_sticker_files'].append(f + ".media" + ext)

        for f in ctx.user_data['user_sticker_files']:
            if '.media' in f:
                if ctx.user_data['telegram_sticker_is_animated']:
                    ret = ff_convert_to_webm(f)
                else:
                    ret = im_convert_to_webp(f)
                    # Skip errored conversion. Don't panic.
                if ret.returncode == 0:
                    images.append(
                        f + '.webm' if ctx.user_data['telegram_sticker_is_animated'] else f + '.webp')
                else:
                    update.effective_chat.send_message(
                        "WARN: failed processing one media.\n\n" + str(ret.stderr))
            else:
                images.append(f)
    # Line stickers.
    else:
        work_dir = os.path.join(
            main.DATA_DIR, str(update.effective_user.id), ctx.user_data['line_sticker_id'])
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
                    image_name = os.path.join(work_dir, image_id)
                    subprocess.run(
                        ["curl", "-Lo", f"{image_name}.base.png", base_image])
                    subprocess.run(
                        ["curl", "-Lo", f"{image_name}.overlay.png", overlay_image])
                    subprocess.run(main.CONVERT_BIN + [f"{image_name}.base.png", f"{image_name}.overlay.png",
                                                       "-background", "none", "-filter", "Lanczos", "-resize", "512x512", "-composite",
                                                       "-define", "webp:lossless=true",
                                                       f"{image_name}.webp"])
            images = sorted(glob.glob(os.path.join(work_dir, "*.webp")))
        # kakao emoticons
        elif ctx.user_data['line_sticker_type'] == KAKAO_EMOTICON:
            for index, src in enumerate(ctx.user_data['line_sticker_image_sources']):
                subprocess.run(['curl', '-o', f'{os.path.join(work_dir, str(index))}.png', src])
            im_mogrify_to_webp(glob.glob(os.path.join(work_dir, "*.png")))
            images = sorted(glob.glob(os.path.join(work_dir, "*.webp")))

        else:
            zip_file_path = os.path.join(
                work_dir, ctx.user_data['line_sticker_id'] + ".zip")
            subprocess.run(["curl", "-Lo", zip_file_path,
                            ctx.user_data['line_sticker_download_url']])
            subprocess.run(main.BSDTAR_BIN +
                           ["-xf", zip_file_path, "-C", work_dir])
            for f in glob.glob(os.path.join(work_dir, "*key*")) + glob.glob(os.path.join(work_dir, "tab*")) + glob.glob(os.path.join(work_dir, "*meta*")):
                os.remove(f)
            # standard static line stickers.
            if not ctx.user_data['line_sticker_is_animated']:
                im_mogrify_to_webp(glob.glob(os.path.join(work_dir, "*.png")))
                images = sorted(glob.glob(os.path.join(work_dir, "*.webp")))
            # is animated line stickers/emojis.
            else:
                # For LINE Effect stickers, keep static and animated popups.
                if ctx.user_data['line_sticker_type'] == main.LINE_STICKER_POPUP_EFFECT:
                    for f in glob.glob(os.path.join(work_dir, "popup", "*.png")):
                        # workaround for sticker orders.
                        shutil.move(f, os.path.join(work_dir, os.path.basename(
                            f)[:os.path.basename(f).index('.png')] + '@99x.png'))
                elif ctx.user_data['line_sticker_type'] == main.LINE_STICKER_POPUP:
                    work_dir = os.path.join(work_dir, "popup")
                elif ctx.user_data['line_sticker_type'] == main.LINE_STICKER_ANIMATION:
                    work_dir = os.path.join(work_dir, "animation@2x")
                else:
                    pass
                for f in glob.glob(os.path.join(work_dir, "**", "*.png"), recursive=True):
                    ff_convert_to_webm(f)

                images = sorted([f for f in glob.glob(
                    os.path.join(work_dir, "**", "*.webm"), recursive=True)])
    #debug
    q.put(images)

    time_end = time.time()
    print(
        f"Prepared {str(len(images))} stickers in {str(time_end - time_start)} seconds.")


def im_mogrify_to_webp(flist: list):
    # Resize to fulfill telegram's requirement, AR is automatically retained
    # Lanczos resizing produces much sharper image.
    flist = [f + '[0]' for f in flist]
    params = []
    return subprocess.run(main.MOGRIFY_BIN + ["-background", "none", "-filter", "Lanczos", "-resize", "512x512", "-format", "webp",
                                              "-define", "webp:lossless=true", "-define", "webp:method=0"] + params + flist, capture_output=True)


def im_convert_to_webp(f: str):
    # Resize to fulfill telegram's requirement, AR is automatically retained
    # Lanczos resizing produces much sharper image.
    params = []
    return subprocess.run(main.CONVERT_BIN + ["-background", "none", "-filter", "Lanczos", "-resize", "512x512",
                                              "-define", "webp:lossless=true", "-define", "webp:method=0"] + params + [f + "[0]", f + ".webp"], capture_output=True)


def ff_convert_to_webm(f: str, unsharp=False):
    ret = subprocess.run(main.FFMPEG_BIN + ["-hide_banner", "-loglevel", "error", "-i", f,
                                            "-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
                                            "-c:v", "libvpx-vp9", "-cpu-used", "8", "-minrate", "50k", "-b:v", "350k", "-maxrate", "450k",
                                            "-to", "00:00:02.800", "-an", "-y",
                                            f + '.webm'], capture_output=True)

    if ret.returncode != 0:
        return ret
    else:
        if os.path.getsize(f + '.webm') > 255000:
            return subprocess.run(main.FFMPEG_BIN + ["-hide_banner", "-loglevel", "error", "-i", f,
                                                     "-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
                                                     "-c:v", "libvpx-vp9", "-cpu-used", "8", "-minrate", "50k", "-b:v", "200k", "-maxrate", "300k",
                                                     "-to", "00:00:02.800", "-an", "-y",
                                                     f + '.webm'], capture_output=True)
        else:
            return ret

def verify_sticker_id_availability(sticker_id, update, ctx):
    pass


def verify_user_sticker_message(update: Update):
    supported_types = ['document', 'photo', 'video', 'sticker']
    if update.message is None:
        return False
    elif not any(k in dir(update.message) for k in supported_types):
        return False
    else:
        return True


def initialize_user_data(in_command, update: Update, ctx):
    if not os.path.exists(main.DATA_DIR):
        os.makedirs(main.DATA_DIR)
    clean_userdata(update, ctx)
    ctx.user_data['in_command'] = in_command
    ctx.user_data['manual_emoji'] = False
    ctx.user_data['line_process'] = None
    ctx.user_data['line_queue'] = None
    ctx.user_data['line_sticker_url'] = ""
    ctx.user_data['line_store_webpage'] = None
    ctx.user_data['line_sticker_download_url'] = ""
    ctx.user_data['line_sticker_image_sources'] = []
    ctx.user_data['line_sticker_type'] = None
    ctx.user_data['line_sticker_is_animated'] = False
    ctx.user_data['line_sticker_id'] = ""
    ctx.user_data['telegram_sticker_emoji'] = ""
    ctx.user_data['telegram_sticker_id'] = ""
    ctx.user_data['telegram_sticker_title'] = ""
    ctx.user_data['telegram_sticker_is_animated'] = False
    ctx.user_data['telegram_sticker_edit_choice'] = ""
    ctx.user_data['telegram_sticker_edit_mov_prev'] = None
    ctx.user_data['telegram_sticker_files'] = []
    ctx.user_data['user_sticker_archive'] = ""
    ctx.user_data['user_sticker_files'] = []
    ctx.user_data['user_sticker_download_queue'] = []


# Clean temparary user data after each conversasion.
def clean_userdata(update: Update, ctx: CallbackContext):
    if 'line_process' in ctx.user_data:
        if ctx.user_data['line_process'] is not None:
            ctx.user_data['line_process'].kill()
    ctx.user_data.clear()
    userdata_dir = os.path.join(main.DATA_DIR, str(update.effective_user.id))
    if os.path.exists(userdata_dir):
        shutil.rmtree(userdata_dir, ignore_errors=True)


# Due to some unknown bugs, userdata may not be cleaned after conversation.
# If a RAMDISK is used to boost performance, this could be a problem.
# Run a timer to periodically clean outdated userdata.
def start_timer_userdata_gc():
    timer = Timer(43200, start_timer_userdata_gc)
    timer.start()

    data_dir = main.DATA_DIR
    if not os.path.exists(data_dir):
        os.mkdir(data_dir)
        return
    for d in glob.glob(os.path.join(data_dir, "*")):
        mtime = os.path.getmtime(d)
        nowtime = time.time()
        if nowtime - mtime > 43200:
            shutil.rmtree(d, ignore_errors=True)


def delayed_set_webhook():
    subprocess.run(['curl', '-F', 'url=' + main.WEBHOOK_URL + main.BOT_TOKEN,
                   'https://api.telegram.org/bot' + main.BOT_TOKEN + '/setWebhook'])
