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


from telegram import Update, ReplyKeyboardMarkup, Update, ReplyKeyboardRemove
from telegram.ext.callbackcontext import CallbackContext
import main


def print_start_message(update: Update):
    update.message.reply_text(
        """
Hello! I'm moe_sticker_bot doing sticker stuffs! Please select command below:
你好! 歡迎使用萌萌貼圖BOT, 請從下方選擇指令:
こんにちは！　萌え萌えのスタンプBOTです！下からコマンドを選択してくださいね

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
<b>/help  /faq</b><code>
   幫助訊息／常見問題. ヘルプ／よくある質問
</code>
<b>/cancel</b><code>
  Cancel conversation. 中斷指令. キャンセル 
</code>
""", parse_mode="HTML")


def print_help_message(update: Update, BOT_NAME, BOT_VERSION):
    update.message.reply_text(
        f"""
@{BOT_NAME} by @plow283
https://github.com/star-39/moe-sticker-bot
Thank you @StickerGroup for feedbacks and advices!
<code>
This free(as in freedom) software is released under the GPLv3 License.
Comes with ABSOLUTELY NO WARRANTY! All rights reserved.
PRIVACY NOTICE:
  This software does not collect or save any kind of your personal information.
  Complies with Hong Kong Legislation Cap. 486 Personal Data (Privacy) Ordinance.
本BOT為免費提供的自由軟體, 您可以自由使用/分發, 惟無任何保用服務(warranty)!
本軟體授權於通用公共許可證(GPL)v3, 保留所有權利.
私隱聲明:
  本軟體不會採集或存儲任何用戶數據, 遵守香港法例第四百八十六章「私隱條例」.
</code><b>
Please send /start to start using!
請傳送 /start 來開始!
始めるには　/start　を入力してください！
</b>
Advanced commands:
進階指令:<code>
alsi</code>
<code>
BOT_VERSION: {BOT_VERSION}
</code>
""", parse_mode="HTML")


def print_faq_message(update: Update):
    update.message.reply_text(
        f"""
<b>FAQ:</b>
<b>=>Q: The generated sticker set ID has the bot's name as suffix! 
     創建的貼圖包ID末尾有這個bot的名字!
     出来立てのステッカーセットのIDの最後にBOTの名前がはいてる！</b>
=>A: This is compulsory by Telegram, ID of sticker set created by a bot must has it's name as suffix.
     這是Telegram的強制要求, BOT創建的貼圖包ID末尾必須要有BOT的名字.
     これはTelegramからのおきてです。BOTで生成されたステッカーセットのIDの最後に必ずBOTの名前が入っています。
   
<b>=>Q: The sticker set title is in English when <code>auto</code> is used during setting title.
     當設定標題時使用了<code>auto</code>, 結果貼圖包的標題是英文的
     タイトルを入力している時に<code>auto</code>を入力すると、タイトルは英語になっちゃうん</b>
=>A: Line sometimes has sticker set in multiple languages, you should paste LINE store link with language suffix.
     有的LINE貼圖包有多種語言, 請確認輸入LINE商店連結的時候末尾有指定語言.
     LINEのスタンプは時々多言語対応です、リンクを入力するとき、最後に言語コードがあるかどうかを確認してください。

<b>=>Q: No response? 沒有反應? 応答なし？</b>
=>A: The bot might encountered an error, please try sending /cancel
     BOT可能遇到了問題, 請嘗試傳送 /cancel
     問題が発生したかもしれません、 /cancel を送信してください
""", parse_mode="HTML")


def print_import_starting(update, ctx):
    try:
        update.message.reply_text("Now starting, please wait...\n"
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
            update.message.reply_text("You are importing LINE Message Stickers which needs more time to complete.\n"
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
            return update.message.reply_text("<b>Current Status 當前進度</b>\n"
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

    main.retry_do(lambda: update.message.reply_text("The sticker set has been successfully created!\n"
                                                    "貼圖包已經成功創建!\n"
                                                    "ステッカーセットの作成が成功しました！\n\n"
                                                    "https://t.me/addstickers/" + ctx.user_data['telegram_sticker_id']),
                  lambda: False)
    if ctx.user_data['line_sticker_type'] == "sticker_animated":
        update.message.reply_text("It seems the sticker set you imported also has a animated version\n"
                                  "You can use /get_animated_line_sticker to have their GIF version\n"
                                  "您匯入的貼圖包還有動態貼圖版本\n"
                                  "可以使用 /get_animated_line_sticker 獲取GIF版動態貼圖\n"
                                  "このスタンプの動くバージョンもございます。\n"
                                  "/get_animated_line_sticker を使ってGIF版のスタンプを入手できます")
    ctx.bot.send_sticker(update.message.chat_id, ctx.bot.get_sticker_set(ctx.user_data['telegram_sticker_id']).stickers[0])
    update.message.reply_text(
        ctx.user_data['in_command'] + " done! 指令成功完成! /start")


def print_ask_id(update):
    update.message.reply_text(
        "Please enter an ID for this sticker set, used for share link.\n"
        "Can contain only english letters, digits and underscores.\n"
        "Must begin with a letter, can't contain consecutive underscores.\n\n"
        "請給此貼圖包設定一個ID, 用於分享連結.\n"
        "ID只可以由英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.")


def print_wrong_id_syntax(update):
    update.message.reply_text(
        "Wrong ID syntax!! Try again. ID格式錯誤!! 請再試一次.\n\n"
        "Can contain only english letters, digits and underscores.\n"
        "Must begin with a letter, can't contain consecutive underscores.\n"
        "ID只可以由英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.")


def print_ask_emoji(update):
    update.message.reply_text("Please enter emoji representing this sticker set\n"
                              "請傳送用於表示整個貼圖包的emoji\n"
                              "このスタンプセットにふさわしい絵文字を入力してください\n"
                              "eg. ☕ \n\n"
                              "If you want to manually assign different emoji for each sticker, send <code>manual</code>\n"
                              "如果您想要為貼圖包內每個貼圖設定不同的emoji, 請傳送<code>manual</code>\n"
                              "一つずつ絵文字を付けたいなら、<code>manual</code>を送信してください。",
                              reply_markup=ReplyKeyboardMarkup(
                                  [['manual']], one_time_keyboard=True),
                              parse_mode="HTML")


def print_ask_title(update, title: str):
    if title != "":
        update.message.reply_text(
            "Please set a title for this sticker set. Send <code>auto</code> to automatically set original title from LINE Store as shown below:\n"
            "請設定貼圖包的標題, 也就是名字. 傳送<code>auto</code>可以自動設為LINE Store中原版的標題如下:\n"
            "スタンプのタイトルを入力してください。<code>auto</code>を入力すると、LINE STORE上に表記されているのタイトルが自動的以下の通りに設定されます。" + "\n\n" +
            "<code>" + title + "</code>",
            reply_markup=ReplyKeyboardMarkup(
                [['auto']], one_time_keyboard=True),
            parse_mode="HTML")
    else:
        update.message.reply_text(
            "Please set a title for this sticker set.\n"
            "請設定貼圖包的標題, 也就是名字.\n"
            "スタンプのタイトルを入力してください。",
            reply_markup=ReplyKeyboardRemove())


def print_ask_line_store_link(update):
    update.message.reply_text("Please enter LINE store URL of the sticker set\n"
                              "請輸入貼圖包的LINE STORE連結\n"
                              "スタンプのLINE STOREリンクを入力してください\n\n"
                              "<code>eg. https://store.line.me/stickershop/product/9961437/ja</code>",
                              parse_mode="HTML")


def print_not_animated_warning(update):
    update.message.reply_text("Sorry! This LINE Sticker set is NOT animated! Please check again.\n"
                              "抱歉! 這個LINE貼圖包沒有動態版本! 請檢查連結是否有誤.\n"
                              "このスタンプの動くバージョンはございません。もう一度ご確認してください。")


def print_fatal_error(update, err_msg):
    update.message.reply_text("<b>"
                              "Fatal error! Please try again. /start\n"
                              "發生致命錯誤! 請您從頭再試一次. /start\n"
                              "致命的なエラーが発生しました！もう一度やり直してください /start\n\n"
                              "</b>"
                              "<code>" + err_msg + "</code>", parse_mode="HTML")


def print_use_start_command(update):
    update.message.reply_text("Please use /start to see available commands!\n"
                              "請先傳送 /start 來看看可用的指令\n"
                              "/start を送信してコマンドで始めましょう")


def print_suggest_import(update):
    update.message.reply_text("You have sent a LINE Store link, guess you want to import LINE sticker to Telegram? Please send /import_line_sticker\n"
                              "您傳送了一個LINE商店連結, 是想要把LINE貼圖包匯入至Telegram嗎? 請使用 /import_line_sticker\n"
                              "LINEスタンプをインポートしたいんですか？ /import_line_sticker で始めてください")


def print_suggest_download(update):
    update.message.reply_text("You have sent a sticker, guess you want to download this sticker set? Please send /download_telegram_sticker\n"
                              "您傳送了一個貼圖, 是想要下載這個Telegram貼圖包嗎? 請使用 /download_telegram_sticker\n"
                              "このステッカーセットを丸ごとダウンロードしようとしていますか？ /download_telegram_sticker で始めてください")


def print_ask_sticker_archive(update):
    update.message.reply_text("Please send an archive file containing image files.\n"
                              "The archive could be any archive format, eg. <code>ZIP RAR 7z</code>\n"
                              "Image could be any image format, eg. <code>PNG JPG WEBP HEIC</code>\n\n"
                              "請傳送一個內含貼圖圖片的歸檔檔案\n"
                              "檔案可以是任意歸檔格式, 比如 <code>ZIP RAR 7z</code>\n"
                              "圖片可以是任意圖片格式, 比如 <code>PNG JPG WEBP HEIC</code>\n",
                              parse_mode="HTML")


def print_command_done(update, ctx):
    update.message.reply_text(ctx.user_data['in_command'] + " done! 指令成功完成!")


def print_in_conv_warning(update, ctx):
    update.message.reply_text("You are already in command : " + str(ctx.user_data['in_command']).removeprefix("/") + "\n\n"
                              "If you encountered a problem, please send /cancel and start over.\n"
                              "如果您遇到了問題, 請傳送 /cancel 來試試重新開始.")


def print_ask_telegram_sticker(update):
    update.message.reply_text("Please send a sticker.\n"
                              "請傳送一張Telegram貼圖.\n"
                              "ステッカーを一つ送信してください。")
