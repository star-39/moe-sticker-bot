from threading import Timer
import glob
import os
import time
import shutil
import traceback
import main
import telegram.bot
from telegram.ext import messagequeue as mq
from telegram import Update
from telegram.ext import CallbackContext


class MQBot(telegram.bot.Bot):
    def __init__(self, *args, is_queued_def=True, mqueue=None, **kwargs):
        super(MQBot, self).__init__(*args, **kwargs)
        # below 2 attributes should be provided for decorator usage
        self._is_messages_queued_default = is_queued_def
        self._msg_queue = mqueue or mq.MessageQueue()

    def __del__(self):
        try:
            self._msg_queue.stop()
        except:
            pass

    @mq.queuedmessage
    def add_sticker_to_set(self, *args, **kwargs):
        return super(MQBot, self).add_sticker_to_set(*args, **kwargs)
    @mq.queuedmessage
    def create_sticker_set(self, *args, **kwargs):
        return super(MQBot, self).create_sticker_set(*args, **kwargs)


# Uploading sticker could easily trigger Telegram's flood limit,
# however, documentation never specified this limit,
# hence, we should at least retry after triggering the limit.
def retry_do(func):
    for index in range(5):
        try:
            func()
        except telegram.error.RetryAfter as ra:
            if index == 4:
                return ra
            time.sleep(ra.retry_after)

        except Exception as e:
            if index == 4:
                print(traceback.format_exc())
                return traceback.format_exc().replace('<', '＜').replace('>', '＞')
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
