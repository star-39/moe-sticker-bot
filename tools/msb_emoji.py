#!/usr/bin/python3

import emoji
import sys
import json

# This python script is intended to extract emoji(s) from a user input string
# which might contain emoji(s) and sparse characters.

# Usage:
# 1st cmdline arg: 'string', 'json'.
# 2nd cmdline arg: string of emojis.

# Example:
# ./msb_emoji.py string 🌸

def main():
    if len(sys.argv) < 3:
        print("wrong cmd line argument.") 
        return -1

    s = sys.argv[2]
    e = emoji.distinct_emoji_list(s)

    if sys.argv[1] == 'string':
        sys.stdout.write(''.join(e))
    elif sys.argv[1] == 'json':
        sys.stdout.write(json.dumps(e, ensure_ascii=False))

    return 0

main()
