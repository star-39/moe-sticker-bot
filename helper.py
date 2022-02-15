import subprocess
from threading import Timer
import glob
import os
import time
import shutil
import traceback
import main
import telegram.bot
import platform
import telegram
from telegram.ext import messagequeue as mq
from telegram import Update, File
from telegram.ext import CallbackContext


# class MQBot(telegram.bot.Bot):
#     def __init__(self, *args, is_queued_def=True, mqueue=None, **kwargs):
#         super(MQBot, self).__init__(*args, **kwargs)
#         # below 2 attributes should be provided for decorator usage
#         self._is_messages_queued_default = is_queued_def
#         self._msg_queue = mqueue or mq.MessageQueue()

#     def __del__(self):
#         try:
#             self._msg_queue.stop()
#         except:
#             pass

#     @mq.queuedmessage
#     def add_sticker_to_set(self, *args, **kwargs):
#         return super(MQBot, self).add_sticker_to_set(*args, **kwargs)
#     @mq.queuedmessage
#     def create_sticker_set(self, *args, **kwargs):
#         return super(MQBot, self).create_sticker_set(*args, **kwargs)


# Names of binaries that we depend on vary across different OSes.
# To make the code truely cross-platform, this problem should be sloved in code,
# but not in environment.
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
def retry_do(func, is_fake_ra):
    for index in range(5):
        try:
            func()
        except telegram.error.RetryAfter as ra:
            if index == 4:
                return ra
            time.sleep(ra.retry_after)

            if is_fake_ra():
                break
            else:
                continue

        except Exception as e:
            if index == 4:
                print(traceback.format_exc())
                return e
            # It could probably be a network problem, sleep for a while and try again.
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
    # ctx.user_data['user_sticker_download_queue'].append(save_path)
    # print("start")
    f.download(save_path)
    # print("end")
    # ctx.user_data['user_sticker_download_queue'].remove(save_path)


def wait_download_queue(update, ctx):
    time.sleep(5)
    # if len(ctx.user_data['user_sticker_download_queue']) > 0:
    #     for _ in range(12):
    #         if len(ctx.user_data['user_sticker_download_queue']) == 0:
    #             return
    #         else:
    #             time.sleep(5)
    # else:
    #     return

    # update.effective_chat.send_message(
    #     "unknown error! try sending done again or /cancel")


def im_convert_to_webp(f: str, unsharp=False):
    # Resize to fulfill telegram's requirement, AR is automatically retained
    # Lanczos resizing produces much sharper image.
    params = []
    if unsharp:
        params = ['-unsharp', '0']
    return subprocess.run(main.CONVERT_BIN + ["-background", "none", "-filter", "Lanczos", "-resize", "512x512", "-define", "webp:lossless=true"] + params + [f + "[0]", f + ".webp"], capture_output=True)


def ff_convert_to_webm(f: str, unsharp=False):
    ret = subprocess.run(main.FFMPEG_BIN + ["-hide_banner", "-loglevel", "error", "-i", f,
                                             "-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
                                             "-c:v", "libvpx-vp9", "-cpu-used", "5", "-minrate", "50k", "-b:v", "350k", "-maxrate", "450k",
                                             "-to", "00:00:02.800", "-an",
                                             f + '.webm'], capture_output=True)

    if ret.returncode != 0:
        return ret
    else:
        if os.path.getsize(f + '.webm') > 255000:
            return subprocess.run(main.FFMPEG_BIN + ["-hide_banner", "-loglevel", "error", "-i", f,
                                             "-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
                                             "-c:v", "libvpx-vp9", "-cpu-used", "5", "-minrate", "50k", "-b:v", "250k", "-maxrate", "300k",
                                             "-to", "00:00:02.800", "-an",
                                             f + '.webm'], capture_output=True)
        else:
            return ret


def verify_user_sticker_message(update: Update):
    supported_types = ('document', 'photo', 'video')
    if update.message is None:
        return False
    elif not supported_types in update.message:
        return False
    else: 
        return True


def initialize_user_data(update: Update, ctx):
    if not os.path.exists(main.DATA_DIR):
        os.makedirs(main.DATA_DIR)
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
    ctx.user_data['telegram_sticker_edit_choice'] = ""
    ctx.user_data['telegram_sticker_edit_mov_prev'] = None
    ctx.user_data['telegram_sticker_files'] = []
    ctx.user_data['user_sticker_archive'] = ""
    ctx.user_data['user_sticker_files'] = []
    ctx.user_data['user_sticker_download_queue'] = []


# Clean temparary user data after each conversasion.
def clean_userdata(update: Update, ctx: CallbackContext):
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
    subprocess.run(['curl', '-F', 'url=' + main.WEBHOOK_URL + main.BOT_TOKEN, 'https://api.telegram.org/bot' + main.BOT_TOKEN + '/setWebhook'])
    