#!/usr/bin/python3

# Utilize rlottie-python to convert TGS images.
# Credit https://github.com/laggykiller/rlottie-python GPL-2.0 license  Copyright @laggykiller


# Example:
# msb_rlottie in.tgs out.gif

from rlottie_python import LottieAnimation
import sys

def main():
    f_in = sys.argv[1]
    f_out = sys.argv[2]

    anim = LottieAnimation.from_tgs(f_in)
    anim.save_animation(f_out)

main()
