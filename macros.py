import platform

# Line sticker types
LINE_STICKER_STATIC = "line_s"  #普通貼圖
LINE_STICKER_ANIMATION = "line_a"  #動態貼圖
LINE_STICKER_POPUP = "line_p"  #全螢幕
LINE_STICKER_POPUP_EFFECT = "line_f" #特效
LINE_EMOJI_STATIC = "line_e"  #表情貼
LINE_EMOJI_ANIMATION = "line_i"  #動態表情貼
LINE_STICKER_MESSAGE = "line_m"  #訊息
LINE_STICKER_NAME = "line_n"  #隨你填
KAKAO_EMOTICON = "kakao_e"  #KAKAOTALK普通貼圖

FFMPEG_BIN = ['ffmpeg']
MOGRIFY_BIN = ['mogrify'] if platform.system() == "Linux" else ['magick', 'mogrify']
CONVERT_BIN = ['convert'] if platform.system() == "Linux" else ['magick', 'convert']
BSDTAR_BIN = ['bsdtar'] if platform.system() == "Linux" else ['tar']  #On win32 and darwin, tar implies bsdtar
