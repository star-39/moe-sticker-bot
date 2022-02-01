from threading import Timer
import glob
import os
import time
import shutil
import main
import telegram.bot
from telegram.ext import messagequeue as mq

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
