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


import traceback
from telegram import Update, ReplyKeyboardMarkup, Update, ReplyKeyboardRemove, InlineKeyboardButton, InlineKeyboardMarkup, Video
from telegram.callbackquery import CallbackQuery, Message
from telegram.ext.callbackcontext import CallbackContext
import main

inline_kb_ASK_EMOJI = InlineKeyboardMarkup([[InlineKeyboardButton(
    "⭐️Random/隨機", callback_data="random"), InlineKeyboardButton("Manual/手動", callback_data="manual")]])
inline_kb_ASK_TYPE = InlineKeyboardMarkup([[InlineKeyboardButton(
    "Static/靜態", callback_data="static"), InlineKeyboardButton("Animated/動態", callback_data="animated")]])
inline_kb_STATIC_SELECTED = InlineKeyboardMarkup(
    [[InlineKeyboardButton("Static selected/已選靜態", callback_data="none")]])
inline_kb_ANIMATED_SELECTED = InlineKeyboardMarkup(
    [[InlineKeyboardButton("Animated selected/已選動態", callback_data="none")]])
inline_kb_AUTO = InlineKeyboardMarkup(
    [[InlineKeyboardButton("Auto/自動", callback_data="auto")]])
inline_kb_AUTO_SELECTED = InlineKeyboardMarkup(
    [[InlineKeyboardButton("Auto selected/已選自動", callback_data="none")]])
inline_kb_MANUAL = InlineKeyboardMarkup(
    [[InlineKeyboardButton("Manual/手動", callback_data="manual")]])
inline_kb_MANUAL_SELECTED = InlineKeyboardMarkup(
    [[InlineKeyboardButton("Manual selected/已選手動", callback_data="none")]])
inline_kb_RANDOM_SELECTED = InlineKeyboardMarkup(
    [[InlineKeyboardButton("Random selected/已選隨機", callback_data="none")]])
inline_kb_MANAGE_SET = InlineKeyboardMarkup([
    [InlineKeyboardButton("Add sticker/增添貼圖", callback_data="add")],
    [InlineKeyboardButton("Delete sticker/刪除貼圖", callback_data="del")],
    [InlineKeyboardButton("Change order/調整順序", callback_data="mov")]])

reply_kb_DONE = ReplyKeyboardMarkup([['done']], one_time_keyboard=True)


def print_start_message(update: Update):
    update.effective_chat.send_message(
        """
Hello! I'm moe_sticker_bot doing sticker stuffs! Please select command below:
你好! 歡迎使用萌萌貼圖BOT, 請從下方選擇指令:
こんにちは！萌え萌えのスタンプBOTです！下からコマンドを選択してくださいね

<b>/import_line_sticker</b><code>
  匯入LINE貼圖包至Telegram
  LINEスタンプをTelegramにインポート
</code>
<b>/download_telegram_sticker</b><code>
  下載Telegram的貼圖包
  Telegramステッカーセットをダウンロード
</code>
<b>/create_sticker_set</b><code>
  創建新的Telegram的貼圖包.
  Telegramステッカーセットの新規作成
</code>
<b>/manage_sticker_set</b><code>
  管理Telegram貼圖包(增添/刪除/排序).
  Telegramステッカーセットの管理(追加/削除/順番)
</code>
<b>/faq  /about</b><code>
   常見問題/關於. よくある質問/について
</code>
<b>/cancel</b><code>
  Interrupt conversation. 中斷指令. キャンセル 
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
始めるには /start を送信してください
</b>
Advanced commands:
進階指令:
<code>alsi</code>  /download_line_sticker
<code>
BOT_VERSION: {BOT_VERSION}
</code>
""", parse_mode="HTML")


def print_faq_message(update: Update):
    update.effective_chat.send_message(
        f"""
<b>FAQ:</b>
<b>
Q:  I'm not that sure how to use this bot...
    我不太會用...</b>
A:  Your interaction with this bot is done with "conversation",
    when you send a command, a "conversation" starts, follow 
    what the bot says and you will get there.
    使用此bot的基本概念是"會話", 當您傳送一個指令後, 即進入了"會話",
    跟隨bot向您傳送的提示一步一步操作, 就可以了.
<b>
Q:  The generated sticker set ID has the bot's name as suffix.
    創建的貼圖包ID末尾有這個bot的名字.</b>
A:  This is forced by Telegram, ID of sticker set created by bot must has it's name as suffix.
    這是Telegram的強制要求, BOT創建的貼圖包ID末尾必須要有BOT的名字.
<b>
Q:  The sticker set title is in English when <code>auto</code> is used during setting title.
    當設定標題時使用了<code>auto</code>, 結果貼圖包的標題是英文的</b>
A:  The sticker set is multilingual, you should paste LINE store link with language suffix.
    有的LINE貼圖包有多種語言, 請確認LINE商店連結的末尾有指定語言.
<b>
Q:  Can I add video sticker to static sticker set or vice versa?
    我可以往靜態貼圖包加動態貼圖, 或者反之嗎?</b>
A:  Of course you can! Video will be static in static set
    and static sticker will remain static in video set.
    當然可以! 動態貼圖在靜態貼圖包裡會變成靜態, 靜態貼圖在動態貼圖包裡依然會是靜態.
<b>
Q:  Who owns the sticker sets the bot created?
    BOT創造的貼圖包由誰所有?</b>
A:  It's you of course! Albeit compulsory suffix in ID, you are the owner of the sticker sets.
    You can manage them through Telegram's official @Stickers bot.
    當然是您! 雖然ID末尾強制有BOT的名字, 但是貼圖包的擁有人是您本人.
    您可以通過Telegram官方的 @Stickers BOT管理您創進的貼圖包.
<b>
Q: No response? 沒有反應?</b>
A:  The bot might encountered an error, please try sending /cancel
    BOT可能遇到了問題, 請嘗試傳送 /cancel
""", parse_mode="HTML")


def print_import_processing(update: Update, ctx):
    try:
        return update.effective_chat.send_message("Preparing stickers, please wait...\n"
                                                  "正在準備貼圖, 請稍後...\n"
                                                  "作業が開始しています、少々お時間を...\n\n"
                                                  "<code>"
                                                  f"LINE TYPE: {ctx.user_data['line_sticker_type']}\n"
                                                  f"LINE ID: {ctx.user_data['line_sticker_id']}\n"
                                                  f"TG ID: {ctx.user_data['telegram_sticker_id']}\n"
                                                  f"TG TITLE: {ctx.user_data['telegram_sticker_title']}\n"
                                                  f"TG LINK: </code>https://t.me/addstickers/{ctx.user_data['telegram_sticker_id']}\n\n"
                                                  "<b>Progress / 進度</b>\n"
                                                  "<code>Preparing stickers...</code>\n",
                                                  parse_mode="HTML")
    except:
        pass


def edit_message_progress(message_progress: Message, ctx: CallbackContext, current, total):
    progress_1 = '[=>                  ]'
    progress_25 = '[====>               ]'
    progress_50 = '[=========>          ]'
    progress_75 = '[==============>     ]'
    progress_100 = '[====================]'
    try:
        message_header = message_progress.text_html[:message_progress.text_html.rfind(
            '<code>')]
        if current == 1:
            message_progress.edit_text(message_header + "<code>" + progress_1 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) +
                                       "     </code>",
                                       parse_mode="HTML")
        if current == int(0.25 * total):
            message_progress.edit_text(message_header + "<code>" + progress_25 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == int(0.5 * total):
            message_progress.edit_text(message_header + "<code>" + progress_50 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == int(0.75 * total):
            message_progress.edit_text(message_header + "<code>" + progress_75 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
        if current == total:
            message_progress.edit_text(message_header + "<code>" + progress_100 + "</code>\n"
                                       "<code>       " +
                                       str(current) + " of " +
                                       str(total) + "     </code>",
                                       parse_mode="HTML")
        if current > total:
            message_progress.edit_text(message_header + "\n"
                                       "√  " +
                                       ctx.user_data['in_command'] + "  /start"
                                       "\nCommand success. 成功完成指令.",
                                       parse_mode="HTML")

    except:
        print(traceback.format_exc())


def print_sticker_done(update: Update, ctx: CallbackContext):
    update.effective_chat.send_message("The sticker set has been successfully created!\n"
                                       "貼圖包已經成功創建!\n"
                                       "ステッカーセットの作成が成功しました！\n\n"
                                       "https://t.me/addstickers/" + ctx.user_data['telegram_sticker_id'])
    ctx.bot.send_sticker(update.effective_chat.id, ctx.bot.get_sticker_set(
        ctx.user_data['telegram_sticker_id']).stickers[0])


def print_ask_id(update: Update):
    update.effective_chat.send_message(
        "Please enter an ID for this sticker set, used for share link.\n"
        "Can contain only english letters, digits and underscores.\n"
        "Must begin with a letter, can't contain consecutive underscores.\n\n"
        "請給此貼圖包設定一個ID, 用於分享連結.\n"
        "ID只可以由英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.",
        reply_markup=inline_kb_AUTO)


def print_file_too_big(update: Update):
    update.effective_chat.send_message(
        "This file is too big, skipping...\n",
        "這個檔案太大了, 已略去...")


def print_ask_sticker_set(update):
    update.effective_chat.send_message(
        "Send a sticker from the sticker set that want to edit,\n"
        "or send its share link or ID.\n\n"
        "您想要修改哪個貼圖包? 請傳送那個貼圖包內任意一張貼圖,\n"
        "或者是它的分享連結或ID.")


def print_ask_what_to_edit(update: Update):
    update.effective_chat.send_message(
        "What do you want to edit? Please select below:\n"
        "您想要修改貼圖包的甚麼內容? 請選擇:",
        reply_markup=inline_kb_MANAGE_SET
        )

def print_ask_which_to_delete(update: Update):
    update.effective_chat.send_message(
        "Which sticker do you want to delete? Please send it.\n"
        "您想要刪除哪一個貼圖? 請傳送那個貼圖"
    )


def print_ask_which_to_move(update: Update):
    update.effective_chat.send_message(
        "Please send the sticker that you want to move.\n"
        "請傳送您想要移動位置的那個貼圖."
    )

def print_ask_where_to_move(update: Update):
    update.effective_chat.send_message(
        "Where do you want to move this sticker to?\n"
        "Send a sticker, then the previous sticker will be inserted to that position.\n"
        "您想要把貼圖移動到哪裡?\n"
        "請傳送一個貼圖, 原先的貼圖便會插入到那個位置上."
    )


def print_wrong_id_syntax(update):
    update.effective_chat.send_message(
        "Wrong ID syntax!! Try again. ID格式錯誤!! 請再試一次.\n\n"
        "Can contain only less than 64 english letters, digits and underscores.\n"
        "Must begin with a letter, can't contain consecutive underscores.\n"
        "ID只可以由少於64個英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.")


def print_ask_emoji(update: Update):
    update.effective_chat.send_message("Please send emoji representing this sticker set\n"
                                       "請傳送用於表示整個貼圖包的emoji\n"
                                       "このスタンプセットにふさわしい絵文字を送信してください\n"
                                       "eg. ☕ \n\n"
                                       "To manually assign different emoji for each sticker, press Manual button\n"
                                       "如果想要為每個貼圖分別設定不同的emoji, 請按下Manual按鈕\n"
                                       "一つずつ絵文字を付けたいなら、Manualボタンを押してください",
                                       reply_markup=inline_kb_ASK_EMOJI)


def print_ask_emoji_for_single_sticker(update: Update, ctx: CallbackContext):
    if ".webp" in ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']]:
        ctx.bot.send_photo(chat_id=update.effective_chat.id,
                           caption="Please send emoji(s) representing this sticker\n"
                           "請傳送代表這個貼圖的emoji(可以多個)\n"
                           "このスタンプにふさわしい絵文字を送信してください(複数可)\n" +
                           f"{ctx.user_data['manual_emoji_index'] + 1} of {len(ctx.user_data['img_files_path'])}",
                           photo=open(ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']], 'rb'))
    elif ".webm" in ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']]:
        ctx.bot.send_video(chat_id=update.effective_chat.id,
                           caption="Please send emoji(s) representing this sticker\n"
                           "請傳送代表這個貼圖的emoji(可以多個)\n"
                           "このスタンプにふさわしい絵文字を送信してください(複数可)\n" +
                           f"{ctx.user_data['manual_emoji_index'] + 1} of {len(ctx.user_data['img_files_path'])}",
                           video=open(ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']], 'rb'))
    else:
        update.effective_chat.send_sticker(
            sticker=ctx.user_data['img_files_path'][ctx.user_data['manual_emoji_index']])
        update.effective_chat.send_message("Please send emoji(s) representing this sticker\n"
                                           "請傳送代表這個貼圖的emoji(可以多個)\n"
                                           "このスタンプにふさわしい絵文字を送信してください(複数可)\n")


def print_ask_title(update: Update, title: str):
    if title != "":
        update.effective_chat.send_message(
            "Please set a title for this sticker set. Press Auto button to set title from LINE Store as shown below:\n"
            "請設定貼圖包的標題.按下Auto按鈕可以自動設為LINE Store中的標題如下:\n"
            "スタンプのタイトルを送信してください。Autoボタンを押すと、LINE STOREに表記されているタイトルが設定されます。" + "\n\n" +
            "<code>" + title + "</code>",
            reply_markup=inline_kb_AUTO,
            parse_mode="HTML")
    else:
        update.effective_chat.send_message(
            "Please set a title for this sticker set.\n"
            "請設定貼圖包的標題.\n"
            "スタンプのタイトルを送信してください。")


def edit_inline_kb_auto_selected(query: CallbackQuery):
    query.edit_message_reply_markup(inline_kb_AUTO_SELECTED)


def edit_inline_kb_manual_selected(query: CallbackQuery):
    query.edit_message_reply_markup(inline_kb_MANUAL_SELECTED)


def edit_inline_kb_random_selected(query: CallbackQuery):
    query.edit_message_reply_markup(inline_kb_RANDOM_SELECTED)


def print_ask_line_store_link(update):
    update.effective_chat.send_message("Please enter LINE store URL of the sticker set\n"
                                       "請傳送貼圖包的LINE STORE連結\n"
                                       "スタンプのLINE STOREリンクを送信してください\n\n"
                                       "<code>eg. https://store.line.me/stickershop/product/9961437/ja</code>",
                                       parse_mode="HTML")


def print_fatal_error(update, err_msg):
    update.effective_chat.send_message("<b>"
                                       "Fatal error! Please try again. /start\n"
                                       "發生致命錯誤! 請您從頭再試一次. /start\n"
                                       "致命的なエラーが発生しました！もう一度やり直してください /start\n\n"
                                       "</b>"
                                       "<code>" + err_msg.replace('<', '＜').replace('>', '＞') + "</code>", parse_mode="HTML")


def print_use_start_command(update):
    update.effective_chat.send_message("Please use /start to see available commands!\n"
                                       "請先傳送 /start 來看看可用的指令\n"
                                       "/start を送信してコマンドで始めましょう")


def print_suggest_import(update):
    update.effective_chat.send_message("You have sent a LINE Store link, guess you want to import LINE sticker to Telegram? Please send /import_line_sticker\n"
                                       "您傳送了一個LINE商店連結, 是想要把LINE貼圖包匯入至Telegram嗎? 請使用 /import_line_sticker\n"
                                       "LINEスタンプをインポートしたいんですか？ /import_line_sticker で始めてください")


def print_suggest_download(update):
    update.effective_chat.send_message("Guess you want to download this sticker set? Please use /download_telegram_sticker\n"
                                       "如果您想要下載這個Telegram貼圖包, 請使用 /download_telegram_sticker\n"
                                       "このステッカーセットをダウンロードしようとしていますか？ /download_telegram_sticker で始めてください")


def print_ask_user_sticker(update: Update, ctx: CallbackContext):
    if ctx.user_data['telegram_sticker_is_animated'] is True:
        update.effective_chat.send_message("Please send videos/stickers(less than 120 in total)(don't group items),\n"
                                           "wait until upload complete, then send <code>done</code>\n"
                                           "Video can be in any format, should shorter than 3 seconds.\n\n"
                                           "請傳送任意格式的短片(視訊)(少於120張)(不要合併成組), 時長應短於3秒鐘\n"
                                           "請等候所有檔案上載完成, 然後傳送<code>done</code>\n",
                                           parse_mode="HTML",
                                           reply_markup=reply_kb_DONE)
    else:
        update.effective_chat.send_message("Please send images/photos/stickers(less than 120 in total)(don't group items),\n"
                                           "or send an archive containing image files,\n"
                                           "wait until upload complete, then send <code>done</code>\n\n"
                                           "請傳送任意格式的圖片/照片/貼圖(少於120張)(不要合併成組)\n"
                                           "或者傳送內有貼圖檔案的歸檔,\n"
                                           "請等候所有檔案上載完成, 然後傳送<code>done</code>\n",
                                           parse_mode="HTML",
                                           reply_markup=reply_kb_DONE)


def print_do_not_send_media_group(update: Update, ctx: CallbackContext):
    update.effective_chat.send_message("Please do not group media files,\n"
                                        "send files separately. \n"
                                        "Skipping this one... Please try again.\n\n"
                                        "請不要合併成組傳送(group items),\n"
                                        "請分開傳送這些檔案.\n"
                                        "已略去這個組訊息, 請再試一次.\n",
                                        reply_to_message_id=update.message.message_id,
                                        reply_markup=reply_kb_DONE)


def print_command_done(update, ctx):
    update.effective_chat.send_message(
        ctx.user_data['in_command'] + "  /start \nCommand success. 成功完成指令.")


def print_in_conv_warning(update, ctx):
    if 'in_command' in ctx.user_data:
        update.effective_chat.send_message("Oops, please follow the bot's instructions to send messages.\n"
                                           "If you encountered a problem, try to send /cancel and start over.\n"
                                           "請跟隨bot的提示傳送相應的訊息喔."
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


def print_wrong_LINE_STORE_URL(update, err_msg):
    update.effective_chat.send_message('Make sure you sent a correct LINE Store link and try again.\n'
                                       '請確認傳送的是正確的LINE商店URL連結後再試一次.\n'
                                       '正しいLINEスタンプストアのリンクを送信してください\n\n' + err_msg)


def print_command_canceled(update):
    update.effective_chat.send_message("Command terminated.\n"
                                       "已中斷指令.\n"
                                       "コマンドは中止されました")


def print_no_user_sticker_received(update):
    update.effective_chat.send_message('Please send photos/images/stickers/videos/archive first, then send "done" to continue\n'
                                       '請先傳送圖片/照片/貼圖/短片/歸檔後, 再傳送"done"來繼續.', reply_markup=reply_kb_DONE)


def print_user_sticker_done(update, ctx):
    update.effective_chat.send_message(f"Done. {len(ctx.user_data['user_sticker_files'])} stickers received.\n"
                                       f"成功收到 {len(ctx.user_data['user_sticker_files'])} 張貼圖",
                                       reply_markup=ReplyKeyboardRemove())


def print_ask_type_to_create(update: Update):
    update.effective_chat.send_message("What kind of sticker set you want to create?\n"
                                       "您想要創建何種貼圖包?",
                                       reply_markup=inline_kb_ASK_TYPE)
