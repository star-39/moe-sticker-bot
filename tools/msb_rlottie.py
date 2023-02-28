#!/usr/bin/python3

# Utilize rlottie-python to convert TGS images.
# Credit https://github.com/laggykiller/rlottie-python GPL-2.0 license  Copyright @laggykiller


# Example:
# msb_rlottie in.tgs out.webp 70

from rlottie_python import LottieAnimation
import sys


def main():
    f = sys.argv[1]
    fout = sys.argv[2]
    q = sys.argv[3]

    anim = LottieAnimation.from_tgs(f)

    anim.save_animation(fout, quality=int(q), minimize_size=True, method=4)
    return 0

main()
