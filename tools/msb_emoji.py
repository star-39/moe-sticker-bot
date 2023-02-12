#!/usr/bin/python3

import emoji
import sys
import json

# This simple python tool utilizes
# 'emoji' package in PyPI which has great funcionality in
# parsing and extracting complicated emojis from string.

# Usage:
# 1st cmdline arg: 'string', 'json'.
# 2nd cmdline arg: text containing emoji(s).

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
