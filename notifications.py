import telegram.error
from telegram import Update, Bot, ReplyKeyboardMarkup, Update, ReplyKeyboardRemove
from telegram.ext import Updater, CommandHandler, CallbackContext, ConversationHandler, MessageHandler, Filters
from main import GlobalConfigs
import main


def start(update: Update, ctx: CallbackContext) -> None:
    update.message.reply_text(
        """
Hello! I'm moe_sticker_bot doing sticker stuffs! Please select command below:
你好! 歡迎使用萌萌貼圖BOT, 請從下方選擇指令:
こんにちは！　萌え萌えのスタンプBOTです！下からコマンドを選択してくださいね

<b>/import_line_sticker</b><code>
    Import sticker set from LINE Store to Telegram
    從LINE STORE將貼圖包匯入至Telegram
    LINE STOREからスタンプをTelegramにインポート
</code>
<b>/download_line_sticker</b><code>
    Download sticker set from LINE Store
    從LINE STORE下載貼圖包
    LINE STOREからスタンプをダウンロード
</code>
<b>/get_animated_line_sticker</b><code>
    Get GIF sticker from LINE Store animated sticker set
    獲取GIF版LINE STORE動態貼圖
    LINE STOREから動くスタンプをGIF形式で入手
</code>
<b>/download_telegram_sticker</b><code>
    Download Telegram sticker set.(webp png)
    下載Telegram的貼圖包.(webp png)
    Telegramのステッカーセットをダウンロード(webp png)
</code>
<b>/help  /faq</b><code>
    Get help. 幫助訊息. ヘルプ
</code>
<b>/cancel</b><code>
    Cancel conversation. 中斷指令. キャンセル 
</code>
""", parse_mode="HTML")


def command_help(update: Update, ctx: CallbackContext) -> None:
    update.message.reply_text(
        f"""
@{GlobalConfigs.BOT_NAME} by @plow283
https://github.com/star-39/moe-sticker-bot
Thank you @StickerGroup for feedbacks and advices!
<code>
This free(as in freedom) software comes with ABSOLUTELY NO WARRANTY!
Released under the GPL v3 License. All rights reserved.
PRIVACY NOTICE:
    This software(bot) does not collect or save any kind of your personal information.
    Complies with Hong Kong Legislation Cap. 486 Personal Data (Privacy) Ordinance.
本軟體為免費提供的自由軟體, 您可以自由使用/分發本軟體, 惟無任何保用服務!
本軟體授權於公共通用許可證(GPL)v3, 保留所有權利.
私隱聲明:
   本軟體(bot)不會採集任何用戶數據, 遵守香港法例第四百八十六章「私隱條例」.
</code>
<b>
Please send /start to start using!
請傳送 /start 來開始!
始めるには　/start　を入力してください！
</b>
Advanced mode commands:
進階模式指令:
<code>alsi</code>

/faq FAQ 檢閱常見問題
<code>
BOT_VERSION: {GlobalConfigs.BOT_VERSION}
</code>
""", parse_mode="HTML")


def command_faq(update: Update, ctx: CallbackContext):
    update.message.reply_text(
        f"""
<b>FAQ:</b>
=>Q: The generated sticker set ID has the bot's name as suffix! 
　　　創建的貼圖包ID末尾有這個bot的名字!
　　　出来立てのステッカーセットのIDの最後にBOTの名前がはいてる！
=>A: This is compulsory by Telegram, ID of sticker set created by a bot must has it's name as suffix.
   　這是Telegram的強制要求, BOT創建的貼圖包ID末尾必須要有BOT的名字.
   　これはTelegramからのおきてです。BOTで生成されたステッカーセットのIDの最後に必ずBOTの名前が入っています。
   
=>Q: The sticker set title is in English when <code>auto</code> is used during setting title.
   　當設定標題時使用了<code>auto</code>, 結果貼圖包的標題是英文的
   　タイトルを入力している時に<code>auto</code>を入力すると、タイトルは英語になっちゃうん
=>A: Line sometimes has sticker set in multiple languages, you should paste LINE store link with language suffix.
   　有的LINE貼圖包有多種語言, 請確認輸入LINE商店連結的時候末尾有指定語言.
   　LINEのスタンプは時々多言語対応です、リンクを入力するとき、最後に言語コードがあるかどうかを確認してください。

=>Q: No response? 沒有反應? 応答なし？
=>A: The bot might encountered an error, please report to GitHub issue:
     BOT可能遇到了問題, 請報告至GitHub issue網頁:
     https://github.com/star-39/moe-sticker-bot/issues
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


def print_sticker_done(update, ctx):

    main.retry_do(lambda: update.message.reply_text("The sticker set has been successfully created!\n"
                                                    "貼圖包已經成功創建!\n"
                                                    "ステッカーセットの作成が成功しました！\n\n"
                                                    "https://t.me/addstickers/" + ctx.user_data['telegram_sticker_id']),
                  lambda: False,
                  ctx)

    if ctx.user_data['line_sticker_type'] == "sticker_animated":
        update.message.reply_text("It seems the sticker set you imported also has a animated version\n"
                                  "You can use /get_animated_line_sticker to have their GIF version\n"
                                  "您匯入的貼圖包還有動態貼圖版本\n"
                                  "可以使用 /get_animated_line_sticker 獲取GIF版動態貼圖\n"
                                  "このスタンプの動くバージョンもございます。\n"
                                  "/get_animated_line_sticker を使ってGIF版のスタンプを入手できます")
    update.message.reply_text(
        ctx.user_data['in_command'] + " done! 指令成功完成!")


def print_ask_line_store_link(update):
    update.message.reply_text("Please enter LINE store URL of the sticker set\n"
                              "請輸入貼圖包的LINE STORE連結\n"
                              "スタンプのLINE STOREリンクを入力してください\n\n"
                              "<code>eg. https://store.line.me/stickershop/product/9961437/ja</code>",
                              parse_mode="HTML")



def print_fatal_error(update, err_msg):
    update.message.reply_text("Fatal error! Please try again.\n"
                              "發生致命錯誤! 請您重新再試一次.\n\n" +
                              err_msg)
