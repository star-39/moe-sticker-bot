#!/usr/bin/python3

# Utilize rlottie-python to convert TGS images.
# Credit https://github.com/laggykiller/rlottie-python GPL-2.0 license  Copyright @laggykiller


# Example:
# msb_rlottie in.tgs out.webp

from rlottie_python import LottieAnimation
import sys


def main():
    if len(sys.argv) < 3:
        print("wrong cmd line argument.") 
        return -1

    f = sys.argv[1]
    fout = sys.argv[2]

    anim = LottieAnimation.from_tgs(f)

    anim.save_animation(fout)
    return 0

main()
