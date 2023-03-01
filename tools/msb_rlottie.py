#!/usr/bin/python3

# Utilize rlottie-python to convert TGS images.
# Credit https://github.com/laggykiller/rlottie-python GPL-2.0 license  Copyright @laggykiller


# Example:
# msb_rlottie in.tgs out.webp 70

import os
from rlottie_python import LottieAnimation
import sys
import subprocess as sp


# pipe APNG to ffmpeg
def pipe_save_animation(self, save_path, fps: int = 0, frame_num_start: int = None, frame_num_end: int = None, buffer_size: int = None, width: int = None, height: int = None, bytes_per_line: int = None, *args, **kwargs):
    fps_orig = self.lottie_animation_get_framerate()
    duration = self.lottie_animation_get_duration()

    if not fps:
        fps = fps_orig

    if kwargs.get('loop') == None:
        kwargs['loop'] = 0

    frames = int(duration * fps)
    frame_duration = 1000 / fps

    if frame_num_start == None:
        frame_num_start = 0
    if frame_num_end == None:
        frame_num_end = frames

    im_list = []
    for frame in range(frame_num_start, frame_num_end):
        pos = frame / frame_num_end
        frame_num = self.lottie_animation_get_frame_at_pos(pos)
        image = self.render_pillow_frame(frame_num=frame_num, buffer_size=buffer_size,
                                         width=width, height=height, bytes_per_line=bytes_per_line).copy()

        im_list.append(image)

    im_list[0].save(save_path, save_all=True, append_images=im_list[1:],
                    duration=int(frame_duration), *args, **kwargs)


def main():
    f = sys.argv[1]
    fout = sys.argv[2]
    q = sys.argv[3]

    anim = LottieAnimation.from_tgs(f)
    fps = anim.lottie_animation_get_framerate()
    if fps > 25:
        fps = 25

    ext = os.path.splitext(fout)[-1].lower()

    if ext == ".gif":
        cmd_out = ['ffmpeg',
                   '-hide_banner', '-loglevel', 'warning',
                   '-f', 'apng',
                   '-i', '-',
                   "-lavfi", "split[a][b];[a]palettegen[p];[b][p]paletteuse=dither=atkinson",
                   "-gifflags", "-transdiff", "-gifflags", "-offsetting",
                   "-y", fout]

        pipe = sp.Popen(cmd_out, stdin=sp.PIPE)

        pipe_save_animation(self=anim, format='PNG', disposal=1,
                            save_path=pipe.stdin, fps=fps)

        pipe.stdin.close()
        pipe.wait()

        return 0

    elif ext == ".apng":
        anim.save_animation(save_path=fout, disposal=1, quality=int(
            q), minimize_size=True, method=4, fps=fps)
    else:
        anim.save_animation(save_path=fout, quality=int(
            q), minimize_size=True, method=4, fps=fps)

    return 0


main()
