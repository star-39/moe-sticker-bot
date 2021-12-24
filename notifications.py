# moe-sticker-bot @moe_sticker_bot
# Copyright (c) 2020-2021, @plow283 @star-39. All rights reserved
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.


from telegram import Update, ReplyKeyboardMarkup, Update, ReplyKeyboardRemove, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.callbackquery import CallbackQuery
from telegram.ext.callbackcontext import CallbackContext
import main

inline_kb_ASK_EMOJI=InlineKeyboardMarkup([[InlineKeyboardButton("⭐️Random/隨機", callback_data="random"), InlineKeyboardButton("Manual/手動", callback_data="manual")]])
inline_kb_AUTO=InlineKeyboardMarkup([[InlineKeyboardButton("Auto/自動", callback_data="auto")]])
inline_kb_AUTO_SELECTED=InlineKeyboardMarkup([[InlineKeyboardButton("Auto selected/已選自動", callback_data="none")]])
inline_kb_MANUAL=InlineKeyboardMarkup([[InlineKeyboardButton("Manual/手動", callback_data="manual")]])
inline_kb_MANUAL_SELECTED=InlineKeyboardMarkup([[InlineKeyboardButton("Manual selected/已選手動", callback_data="none")]])
inline_kb_RANDOM_SELECTED=InlineKeyboardMarkup([[InlineKeyboardButton("Random selected/已選隨機", callback_data="none")]])


def print_start_message(update: Update):
    update.effective_chat.send_message(
        """
Hello! I'm moe_sticker_bot doing sticker stuffs! Please select command below:
你好! 歡迎使用萌萌貼圖BOT, 請從下方選擇指令:
こんにちは！萌え萌えのスタンプBOTです！下からコマンドを選択してくださいね

<b>/import_line_sticker</b><code>
  從LINE STORE將貼圖包匯入至Telegram
  LINE STOREからスタンプをTelegramにインポート
</code>
<b>/download_line_sticker</b><code>
  從LINE STORE下載貼圖包
  LINE STOREからスタンプをダウンロード
</code>
<b>/get_animated_line_sticker</b><code>
  獲取GIF版LINE STORE動態貼圖
  LINE STOREから動くスタンプをGIF形式で入手
</code>
<b>/download_telegram_sticker</b><code>
  下載Telegram的貼圖包.(webp png)
  Telegramのステッカーセットをダウンロード(webp png)
</code>
<b>/create_sticker_set</b><code>
  創建新的Telegram的貼圖包.
  Telegramステッカーセット新規作成
</code>
<b>/faq  /about</b><code>
   常見問題/關於. よくある質問/について
</code>
<b>/cancel</b><code>
  Cancel conversation. 中斷指令. キャンセル 
</code>
""", parse_mode="HTML")


def print_about_message(update: Update, BOT_NAME, BOT_VERSION):
    update.effective_chat.send_message(
        f"""
@{BOT_NAME} by @plow283
https://github.com/star-39/moe-sticker-bot
Thank you @StickerGroup for feedbacks and advices!
<code>
This free(as in freedom) software is released under the GPLv3 License.
Comes with ABSOLUTELY NO WARRANTY! All rights reserved.
PRIVACY NOTICE:
  This software does not collect or save any kind of your personal information.
本BOT為免費提供的自由軟體, 您可以自由使用/分發, 惟無任何保用服務(warranty)!
本軟體授權於通用公共許可證(GPL)v3, 保留所有權利.
私隱聲明: 本軟體不會採集或存儲任何用戶數據.
</code><b>
Please send /start to start using
請傳送 /start 來開始
始めるには /start を入力してください
</b>
Advanced commands:
進階指令:<code>
alsi</code>
<code>
BOT_VERSION: {BOT_VERSION}
</code>
""", parse_mode="HTML")


def print_faq_message(update: Update):
    update.effective_chat.send_message(
        f"""
<b>FAQ:</b>
<b>
Q:  I not that sure how to use this bot...
    我不太會用...
</b>
A:  Your interaction with this bot is done with "conversation",
    when you send a command, a "conversation" starts, follow 
    what the bot says and you will get there.
    使用此bot的基本概念是"會話", 當您傳送一個指令後, 即進入了"會話",
    跟隨bot向您傳送的提示消息一步一步操作, 就可以了.

<b>
Q:  The generated sticker set ID has the bot's name as suffix.
    創建的貼圖包ID末尾有這個bot的名字..
</b>
A:  This is forced by Telegram, ID of sticker set created by bot must has it's name as suffix.
    這是Telegram的強制要求, BOT創建的貼圖包ID末尾必須要有BOT的名字.
   
<b>
Q:  The sticker set title is in English when <code>auto</code> is used during setting title.
    當設定標題時使用了<code>auto</code>, 結果貼圖包的標題是英文的
</b>
A:  The sticker set is multilingual, you should paste LINE store link with language suffix.
    有的LINE貼圖包有多種語言, 請確認LINE商店連結的末尾有指定語言.

<b>
Q: No response? 沒有反應?
</b>
A:  The bot might encountered an error, please try sending /cancel
    BOT可能遇到了問題, 請嘗試傳送 /cancel
""", parse_mode="HTML")


def print_import_starting(update, ctx):
    try:
        update.effective_chat.send_message("Now starting, please wait...\n"
                                  "正在開始, 請稍後...\n"
                                  "作業がまもなく開始します、少々お時間を...\n\n"
                                  "<code>"
                                  f"LINE TYPE: {ctx.user_data['line_sticker_type']}\n"
                                  f"LINE ID: {ctx.user_data['line_sticker_id']}\n"
                                  f"TG ID: {ctx.user_data['telegram_sticker_id']}\n"
                                  f"TG TITLE: {ctx.user_data['telegram_sticker_title']}\n"
                                  "</code>",
                                  parse_mode="HTML", reply_markup=ReplyKeyboardRemove())
        if ctx.user_data['line_sticker_type'] == "sticker_message":
            update.effective_chat.send_message("You are importing LINE Message Stickers which needs more time to complete.\n"
                                      "您正在匯入LINE訊息貼圖, 這需要更長的等候時間.")
    except:
        pass


def print_progress(message_progress, current, total, update=None):
    progress_1 = '[=>                  ]'
    progress_25 = '[====>               ]'
    progress_50 = '[=========>          ]'
    progress_75 = '[==============>     ]'
    progress_100 = '[====================]'
    try:
        if update is not None:
            return update.effective_chat.send_message("<b>Current Status 當前進度</b>\n"
                                             "<code>" + progress_1 + "</code>\n"
                                             "<code>       " +
                                             str(current) + " of " +
                                             str(total) + "     </code>",
                                             parse_mode="HTML")
        if current == int(0.25 * total):
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_25 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == int(0.5 * total):
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_50 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == int(0.75 * total):
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_75 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == total:
            message_progress.edit_text("<b>Current Status 當前進度</b>\n"
                                       "<code>" + progress_100 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
    except:
        pass


def print_sticker_done(update: Update, ctx: CallbackContext):

    update.effective_chat.send_message("The sticker set has been successfully created!\n"
                              "貼圖包已經成功創建!\n"
                              "ステッカーセットの作成が成功しました！\n\n"
                              "https://t.me/addstickers/" + ctx.user_data['telegram_sticker_id'])
    if ctx.user_data['line_sticker_type'] == "sticker_animated":
        update.effective_chat.send_message("It seems the sticker set you imported also has a animated version\n"
                                  "You can use /get_animated_line_sticker to have their GIF version\n"
                                  "您匯入的貼圖包還有動態貼圖版本\n"
                                  "可以使用 /get_animated_line_sticker 獲取GIF版動態貼圖\n"
                                  "このスタンプの動くバージョンもございます。\n"
                                  "/get_animated_line_sticker を使ってGIF版のスタンプを入手できます")
    ctx.bot.send_sticker(update.effective_chat.id, ctx.bot.get_sticker_set(ctx.user_data['telegram_sticker_id']).stickers[0])
    update.effective_chat.send_message(
        ctx.user_data['in_command'] + " done! 指令成功完成! /start")


def print_ask_id(update):
    update.effective_chat.send_message(
        "Please enter an ID for this sticker set, used for share link.\n"
        "Can contain only english letters, digits and underscores.\n"
        "Must begin with a letter, can't contain consecutive underscores.\n\n"
        "請給此貼圖包設定一個ID, 用於分享連結.\n"
        "ID只可以由英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.")


def print_wrong_id_syntax(update):
    update.effective_chat.send_message(
        "Wrong ID syntax!! Try again. ID格式錯誤!! 請再試一次.\n\n"
        "Can contain only english letters, digits and underscores.\n"
        "Must begin with a letter, can't contain consecutive underscores.\n"
        "ID只可以由英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.")


def print_ask_emoji(update):
    update.effective_chat.send_message("Please send emoji representing this sticker set\n"
                              "請傳送用於表示整個貼圖包的emoji\n"
                              "このスタンプセットにふさわしい絵文字を入力してください\n"
                              "eg. ☕ \n\n"
                              "To manually assign different emoji for each sticker, press Manual button\n"
                              "如果想要為每個貼圖分別設定不同的emoji, 請按下Manual按鈕\n"
                              "一つずつ絵文字を付けたいなら、Manualボタンを押してください",
                              reply_markup=inline_kb_ASK_EMOJI)


def print_ask_title(update: Update, title: str):
    if title != "":
        update.effective_chat.send_message(
            "Please set a title for this sticker set. Press Auto button to set original title from LINE Store as shown below:\n"
            "請設定貼圖包的標題, 也就是名字.按下Auto按鈕可以自動設為LINE Store中原版的標題如下:\n"
            "スタンプのタイトルを入力してください。Autoボタンを押すと、LINE STORE上に表記されているのタイトルが自動的以下の通りに設定されます。" + "\n\n" +
            "<code>" + title + "</code>",
            reply_markup=inline_kb_AUTO,
            parse_mode="HTML")
    else:
        update.effective_chat.send_message(
            "Please set a title for this sticker set.\n"
            "請設定貼圖包的標題, 也就是名字.\n"
            "スタンプのタイトルを入力してください。")

def edit_inline_kb_auto_selected(query: CallbackQuery):
    query.edit_message_reply_markup(inline_kb_AUTO_SELECTED)
    

def edit_inline_kb_manual_selected(query: CallbackQuery):
    query.edit_message_reply_markup(inline_kb_MANUAL_SELECTED)


def edit_inline_kb_random_selected(query: CallbackQuery):
    query.edit_message_reply_markup(inline_kb_RANDOM_SELECTED)


def print_ask_line_store_link(update):
    update.effective_chat.send_message("Please enter LINE store URL of the sticker set\n"
                              "請輸入貼圖包的LINE STORE連結\n"
                              "スタンプのLINE STOREリンクを入力してください\n\n"
                              "<code>eg. https://store.line.me/stickershop/product/9961437/ja</code>",
                              parse_mode="HTML")


def print_not_animated_warning(update):
    update.effective_chat.send_message("Sorry! This LINE Sticker set is NOT animated! Please check again.\n"
                              "抱歉! 這個LINE貼圖包沒有動態版本! 請檢查連結是否有誤.\n"
                              "このスタンプの動くバージョンはございません。もう一度ご確認してください。")


def print_fatal_error(update, err_msg):
    update.effective_chat.send_message("<b>"
                              "Fatal error! Please try again. /start\n"
                              "發生致命錯誤! 請您從頭再試一次. /start\n"
                              "致命的なエラーが発生しました！もう一度やり直してください /start\n\n"
                              "</b>"
                              "<code>" + err_msg + "</code>", parse_mode="HTML")


def print_use_start_command(update):
    update.effective_chat.send_message("Please use /start to see available commands!\n"
                              "請先傳送 /start 來看看可用的指令\n"
                              "/start を送信してコマンドで始めましょう")


def print_suggest_import(update):
    update.effective_chat.send_message("You have sent a LINE Store link, guess you want to import LINE sticker to Telegram? Please send /import_line_sticker\n"
                              "您傳送了一個LINE商店連結, 是想要把LINE貼圖包匯入至Telegram嗎? 請使用 /import_line_sticker\n"
                              "LINEスタンプをインポートしたいんですか？ /import_line_sticker で始めてください")


def print_suggest_download(update):
    update.effective_chat.send_message("You have sent a sticker, guess you want to download this sticker set? Please send /download_telegram_sticker\n"
                              "您傳送了一個貼圖, 是想要下載這個Telegram貼圖包嗎? 請使用 /download_telegram_sticker\n"
                              "このステッカーセットを丸ごとダウンロードしようとしていますか？ /download_telegram_sticker で始めてください")


def print_ask_sticker_archive(update):
    update.effective_chat.send_message("Please send an archive file containing image files.\n"
                              "The archive could be any archive format, eg. <code>ZIP RAR 7z</code>\n"
                              "Image could be any image format, eg. <code>PNG JPG WEBP HEIC</code>\n\n"
                              "請傳送一個內含貼圖圖片的歸檔檔案\n"
                              "檔案可以是任意歸檔格式, 比如 <code>ZIP RAR 7z</code>\n"
                              "圖片可以是任意圖片格式, 比如 <code>PNG JPG WEBP HEIC</code>\n",
                              parse_mode="HTML")


def print_command_done(update, ctx):
    update.effective_chat.send_message(ctx.user_data['in_command'] + " done! 指令成功完成!")


def print_in_conv_warning(update, ctx):
    update.effective_chat.send_message("You are already in command : " + str(ctx.user_data['in_command']).removeprefix("/") + "\n\n"
                              "If you encountered a problem, please send /cancel and start over.\n"
                              "如果您遇到了問題, 請傳送 /cancel 來試試重新開始.")


def print_ask_telegram_sticker(update):
    update.effective_chat.send_message("Please send a sticker.\n"
                              "請傳送一張Telegram貼圖.\n"
                              "ステッカーを一つ送信してください。")


def print_timeout_message(update):
    update.effective_chat.send_message("Timeout has been reached due to long time inactivity. Please start over.\n"
                              "指令因為長時無操作而超時, 請重新開始.\n"
                              "長い間操作がないためタイムアウトしました、もう一度やり直してください。\n\n"
                              "/start",
                              disable_notification=True)


def print_preparing_tg_sticker(update, title, name, amount):
    update.effective_chat.send_message("This might take some time, please wait...\n"
                            "此項作業可能需時較長, 請稍等...\n"""
                            "少々お待ちください...\n"
                            "<code>\n"
                            f"Title: {title}\n"
                            f"ID: {name}\n"
                            f"Amount: {amount}\n"
                            "</code>",
                            parse_mode="HTML")


def print_wrong_LINE_STORE_URL(update, err_msg):
    update.effective_chat.send_message('Make sure you sent a correct LINE Store link and again please.\n'
                            '請確認傳送的是正確的LINE商店URL連結後再試一次.\n'
                            '正しいLINEスタンプストアのリンクを送信してください\n\n' + err_msg)


def print_command_canceled(update):
    update.effective_chat.send_message("Command terminated.\n"
                              "已中斷指令.\n"
                              "コマンドは中止されました")

