from threading import Timer
import glob
import os
import time
import shutil

def userdata_gc(data_dir):
    start_timer_userdata_gc(data_dir)
    if not os.path.exists(data_dir):
        os.mkdir(data_dir)
        return
    for d in glob.glob(os.path.join(data_dir, "*")):
        mtime = os.path.getmtime(d)
        nowtime = time.time()
        if nowtime - mtime > 43200:
            shutil.rmtree(d, ignore_errors=True)


def start_timer_userdata_gc(data_dir):
    timer = Timer(43200, userdata_gc(data_dir))
    timer.start()
