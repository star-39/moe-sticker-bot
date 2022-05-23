package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
)

func sendStartMessage(c tele.Context) error {
	message := `
Hello! I'm <a href="https://github.com/star-39/moe-sticker-bot">moe_sticker_bot</a> doing sticker stuffs!
Send me links or stickers to import or download them, or, use a command below:
你好! 歡迎使用萌萌貼圖BOT, 請傳送連結或貼圖來匯入或下載貼圖喔,
您也可以從下方選擇指令:

<b>/import</b> LINE or kakao stickers to Telegram<code>
  匯入LINE/kakao貼圖包至Telegram
</code>
<b>/download</b> Telegram/LINE/kakao sticker(s)<code>
  下載Telegram/LINE/kakao貼圖包
</code>
<b>/create</b> new sticker set<code>
  創建新的Telegram的貼圖包.
</code>
<b>/manage</b> exsting sticker set (add/delete/reorder)<code>
  管理Telegram貼圖包(增添/刪除/排序).
</code>
<b>/faq  /about</b><code>
   常見問題/關於.
</code>
<b>/exit /quit /cancel</b><code>
  Interrupt conversation. 中斷指令.
</code>
`
	return c.Send(message, tele.ModeHTML, tele.NoPreview)
}

func sendAboutMessage(c tele.Context) {
	c.Send(fmt.Sprintf(`
@%s by @plow283
<b>Please hit Star for this project on Github if you like this bot!
如果您喜歡這個bot, 歡迎Github給本專案標Star喔!
https://github.com/star-39/moe-sticker-bot</b>
Thank you @StickerGroup for feedbacks and advices!
<code>
This free(as in freedom) software is released under the GPLv3 License.
Comes with ABSOLUTELY NO WARRANTY! All rights reserved.
本BOT為免費提供的自由軟體, 您可以自由使用/分發, 惟無任何保用(warranty)!
本軟體授權於通用公眾授權條款(GPL)v3, 保留所有權利.
</code><b>
Please send /start to start using
請傳送 /start 來開始
始めるには /start を送信してください
</b><code>
BOT_VERSION: %s
</code>
`, botName, botVersion), tele.ModeHTML)
}

func sendFAQ(c tele.Context) {
	c.Send(fmt.Sprintf(`
@%s by @plow283
<b>Please hit Star for this project on Github if you like this bot!
如果您喜歡這個bot, 歡迎在Github給本專案標Star喔!
https://github.com/star-39/moe-sticker-bot</b>
------------------------------------
<b>Q: Why ID has suffix: _by_%s ?
為甚麼ID的末尾有: _by_%s ?</b>
A: It's forced by Telegram, bot created sticker set must have its name in ID suffix.
因為這個是Telegram的強制要求, 由bot創造的貼圖ID末尾必須有bot名字.

<b>Q:  Can I add video sticker to static sticker set or vice versa?
    我可以往靜態貼圖包加動態貼圖, 或者反之嗎?</b>
A:  Yes. Video will be static in static set
    and static sticker will remain static in video set.
    當然. 動態貼圖在靜態貼圖包裡會變成靜態, 靜態貼圖在動態貼圖包裡依然會是靜態.

<b>Q:  Who owns the sticker sets the bot created?
    BOT創造的貼圖包由誰所有?</b>
A:  It's you of course. You can manage them through /manage or Telegram's official @Stickers bot.
    當然是您. 您可以通過 /manage 指令或者Telegram官方的 @Stickers 管理您的貼圖包.

`, botName, botName, botName), tele.ModeHTML)
}

func sendAskEmoji(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnManu := selector.Data("Assign separately/分別設定", "manual")
	btnRand := selector.Data(`Batch assign as/一併設定為 "⭐"`, "random")
	selector.Inline(selector.Row(btnManu), selector.Row(btnRand))

	return c.Send(`
Telegram sticker requires emoji to represent it.
Press "Assign separately" to assign emoji one by one.
You can also do batch assign, send a emoji or press button below.
Telegram要求為貼圖設定emoji來表示它.
按下"分別設定"來為每個貼圖都分別設定相應的emoji.
您也一口氣為全部貼圖設定一樣的emoji, 傳送emoji或者按下方按鈕.
`, selector)
}

func sendAskSDownloadChoice(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnRand := selector.Data("This sticker/這張貼圖", "single")
	btnManu := selector.Data("Whole sticker set/整個貼圖包", "whole")
	selector.Inline(selector.Row(btnRand), selector.Row(btnManu))
	return c.Reply(`
You can download this sticker or the whole sticker set, please select below.
您可以下載這個貼圖或者其所屬的整個貼圖包, 請選擇:
`, selector)
}

func sendAskSChoice(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnRand := selector.Data("This sticker/下載這張貼圖", "single")
	btnManu := selector.Data("Whole sticker set/下載整個貼圖包", "whole")
	btnMan := selector.Data("Manage sticker set/管理這個貼圖包", "manage")
	selector.Inline(selector.Row(btnRand), selector.Row(btnManu), selector.Row(btnMan))
	return c.Reply(`
You own this sticker set. You can download or manage this sticker set, please select below.
您擁有這個貼圖包. 您可以下載或者管理這個貼圖包, 請選擇:
`, selector)
}

func sendAskTGLinkChoice(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnManu := selector.Data("Whole sticker set/下載整個貼圖包", "whole")
	btnMan := selector.Data("Manage sticker set/管理這個貼圖包", "manage")
	selector.Inline(selector.Row(btnManu), selector.Row(btnMan))
	return c.Reply(`
You own this sticker set. You can download or manage this sticker set, please select below.
您擁有這個貼圖包. 您可以下載或者管理這個貼圖包, 請選擇:
`, selector)
}

func sendAskWantSDown(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btn1 := selector.Data("Yes", "whole")
	btnNo := selector.Data("No", "bye")
	selector.Inline(selector.Row(btn1), selector.Row(btnNo))
	return c.Reply(`
You can download this sticker set. Press Yes to continue.
您可以下載這個貼圖包, 按下Yes來繼續.
`, selector)
}

func sendAskWantImport(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btn1 := selector.Data("Yes", "yesimport")
	btnNo := selector.Data("No", "bye")
	selector.Inline(selector.Row(btn1), selector.Row(btnNo))
	return c.Reply(`
You can import this sticker set. Press Yes to continue.
您可以匯入這個貼圖包, 按下Yes來繼續.
`, selector)
}

func sendAskWhatToDownload(c tele.Context) error {
	return c.Send("Please send a sticker that you want to download, or its share link(can be either Telegram or LINE ones)\n" +
		"請傳送想要下載的貼圖, 或者是貼圖包的分享連結(可以是Telegram或LINE連結).")
}

func sendAskTitle_Import(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnAuto := selector.Data("Auto", "autoTitle")
	selector.Inline(selector.Row(btnAuto))
	lineTitle := escapeTagMark(users.data[c.Sender().ID].lineData.title) + " @" + botName

	return c.Send("Please set a title for this sticker set. Press Auto button to set title from LINE Store as shown below:\n"+
		"請設定貼圖包的標題.按下Auto按鈕可以自動設為LINE Store中的標題如下:\n\n"+
		"<code>"+lineTitle+"</code>", selector, tele.ModeHTML)
}

func sendAskTitle(c tele.Context) {
	c.Send("Please set a title for this sticker set.\n" +
		"請設定貼圖包的標題.\n" +
		"スタンプのタイトルを送信してください。")
}

func sendAskID(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnAuto := selector.Data("Auto Generate/自動生成", "auto")
	selector.Inline(selector.Row(btnAuto))
	return c.Send(
		fmt.Sprintf(`
Please set an ID for sticker set, used in share link.
Can contain only english letters, digits and underscores.
Must begin with a letter, can't contain consecutive underscores.
請設定貼圖包的ID, 用於分享連結.
ID只可以由英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.",
For example, share link below: 例如以下分享連結:<code>
https://t.me/addstickers/LoveRinneForever_by_%s</code>
<code>LoveRinneForever</code> is the ID you will set.
<code>LoveRinneForever</code> 便是您將要設定的ID.

This is usually not important, it's recommended to press "Auto Generate" button.
ID通常不重要, 建議您按下下方的"自動生成"按鈕.
`, botName), selector, tele.ModeHTML)
}

func sendAskImportLink(c tele.Context) error {
	return c.Send("Please send LINE/kakao store link of the sticker set.\n"+
		"請傳送貼圖包的LINE/kakao Store連結. 您可以在LINE App貼圖商店按右上角的分享->複製連結來取得連結.\n\n"+
		"For example: 例如:\n"+
		"<code>https://store.line.me/stickershop/product/7673/ja</code>\n"+
		"<code>https://e.kakao.com/t/pretty-all-friends</code>", tele.ModeHTML)
}

func sendNotifySExist(c tele.Context, lineID string) bool {
	lines := queryLineS(lineID)
	if len(lines) == 0 {
		return false
	}
	message := "This sticker set exists in our database, you can continue import or just use them if you want.\n" +
		"此套貼圖包已經存在於資料庫中, 您可以繼續匯入, 或者使用下列現成的貼圖包\n\n"

	var entries []string
	for _, l := range lines {
		if l.ae {
			entries = append(entries, fmt.Sprintf(`<a href="%s">%s</a>`, "https://t.me/addstickers/"+l.tg_id, l.tg_title))
		} else {
			// append to top
			entries = append([]string{fmt.Sprintf(`★ <a href="%s">%s</a>`, "https://t.me/addstickers/"+l.tg_id, l.tg_title)}, entries...)
		}
	}
	if len(entries) > 5 {
		entries = entries[:5]
	}
	message += strings.Join(entries, "\n")
	println(message)
	c.Send(message, tele.ModeHTML)
	return true
}

func sendAskStickerFile(c tele.Context) error {
	c.Send("Please send images/photos/stickers(less than 120 in total),\n" +
		"or send an archive containing image files,\n" +
		"wait until upload complete, then send a # mark.\n\n" +
		"請傳送任意格式的圖片/照片/貼圖(少於120張)\n" +
		"或者傳送內有貼圖檔案的歸檔,\n" +
		"請等候所有檔案上載完成, 然後傳送 # 記號\n")

	if users.data[c.Sender().ID].stickerData.isVideo {
		c.Send("Special note: Sending GIF with transparent background will lose transparent due to Telegram client problem.\n" +
			"You can compress your GIF into a ZIP file then send it to bot to bypass.\n" +
			"特別提示: 傳送帶有透明背景的GIF動圖會被Telegram客戶端強制轉換為MP4並且丟失透明層.\n" +
			"您可以將貼圖放入ZIP歸檔中再傳送給bot來繞過這個限制.")
	}
	return nil
}

func sendInStateWarning(c tele.Context) error {
	command := users.data[c.Sender().ID].command
	state := users.data[c.Sender().ID].state

	return c.Send(fmt.Sprintf("Please follow instructions.\n"+
		"Current command: %s\nCurrent state: %s\nYou can also send /quit to terminate session.", command, state))

}

func sendNoSessionWarning(c tele.Context) error {
	return c.Send("Please use /start or send LINE/kakao/Telegram links or stickers.\n請使用 /start 或者傳送LINE/kakao/Telegram連結或貼圖.")
}

func sendAskSTypeToCreate(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnStatic := selector.Data("Static/靜態", "static")
	btnAnimated := selector.Data("Animated/動態", "video")
	selector.Inline(selector.Row(btnStatic, btnAnimated))
	return c.Send("What kind of sticker set you want to create?\n"+
		"您想要創建何種類型的貼圖包?", selector)
}

func sendAskEmojiAssign(c tele.Context) error {
	sd := users.data[c.Sender().ID].stickerData
	caption := fmt.Sprintf(`
Send emoji(s) representing this sticker.
請傳送代表這個貼圖的emoji(可以多個).

%d of %d
`, sd.pos+1, sd.lAmount)

	err := c.Send(&tele.Photo{
		File:    tele.FromDisk(sd.stickers[sd.pos].oPath),
		Caption: caption,
	})
	if err != nil {
		err2 := c.Send(&tele.Video{
			File:    tele.FromDisk(sd.stickers[sd.pos].oPath),
			Caption: caption,
		})
		if err2 != nil {
			err3 := c.Send(&tele.Document{
				File:     tele.FromDisk(sd.stickers[sd.pos].oPath),
				FileName: filepath.Base(sd.stickers[sd.pos].oPath),
				Caption:  caption,
			})
			if err3 != nil {
				return err3
			}
		}
	}
	return nil
}

func sendFatalError(err error, c tele.Context) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "500") {
		errMsg += "\nThis is an internal error of Telegram server, we could do nothing but wait for its recover. Please try again later.\n" +
			"此錯誤為Telegram伺服器之內部錯誤, 無法由bot解決, 只能等候官方修復. 建議您稍後再嘗試一次.\n"
	}
	log.Error("User encountered fatal error!")
	log.Errorln("Raw error:", err)

	c.Send("<b>Fatal error! Please try again. /start\n"+
		"發生嚴重錯誤! 請您從頭再試一次. /start\n"+
		"深刻なエラーが発生しました！もう一度やり直してください /start </b>\n\n"+
		"You can report this error to @plow283 or https://github.com/star-39/moe-sticker-bot/issues\n\n"+
		"<code>"+errMsg+"</code>", tele.ModeHTML, tele.NoPreview)
}

func sendProcessStarted(ud *UserData, c tele.Context, optMsg string) (string, *tele.Message, error) {
	message := fmt.Sprintf(`
Preparing stickers, please wait...
正在準備貼圖, 請稍後...
作業が開始しています、少々お時間を...
<code>
LINE Cat:%s
LINE ID:%s
TG ID:%s
TG Title:</code><a href="%s">%s</a>

<b>Progress / 進展</b>
<code>%s</code>
`, ud.lineData.category,
		ud.lineData.id,
		ud.stickerData.id,
		"https://t.me/addstickers/"+ud.stickerData.id,
		escapeTagMark(ud.stickerData.title),
		optMsg)
	ud.progress = message

	teleMsg, err := c.Bot().Send(c.Recipient(), message, tele.ModeHTML)
	ud.progressMsg = teleMsg
	return message, teleMsg, err
}

func editProgressMsg(cur int, total int, sp string, originText string, teleMsg *tele.Message, c tele.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("editProgressMsg encountered panic! ignoring...")
		}
	}()

	ud, exist := users.data[c.Sender().ID]
	if teleMsg == nil {
		if !exist {
			return nil
		}
		select {
		case <-ud.ctx.Done():
			log.Warn("editProgressMsg received ctxDone!")
			return nil
		default:
		}
		originText = ud.progress
		teleMsg = ud.progressMsg
	}

	header := originText[:strings.LastIndex(originText, "<code>")]
	prog := ""

	if sp != "" {
		prog = sp
		goto SEND
	}
	cur = cur + 1
	if cur == 1 {
		prog = fmt.Sprintf("<code>[=>                  ]\n       %d of %d</code>", cur, total)
	} else if cur == int(float64(0.25)*float64(total)) {
		prog = fmt.Sprintf("<code>[====>               ]\n       %d of %d</code>", cur, total)
	} else if cur == int(float64(0.5)*float64(total)) {
		prog = fmt.Sprintf("<code>[=========>          ]\n       %d of %d</code>", cur, total)
	} else if cur == int(float64(0.75)*float64(total)) {
		prog = fmt.Sprintf("<code>[==============>     ]\n       %d of %d</code>", cur, total)
	} else if cur == total {
		prog = fmt.Sprintf("<code>[====================]\n       %d of %d</code>", cur, total)
	} else {
		return nil
	}
SEND:
	messageText := header + prog
	c.Bot().Edit(teleMsg, messageText, tele.ModeHTML)
	return nil
}

func sendAskSToManage(c tele.Context) error {
	return c.Send("Send a sticker from the sticker set that want to edit,\n" +
		"or send its share link.\n\n" +
		"您想要修改哪個貼圖包? 請傳送那個貼圖包內任意一張貼圖,\n" +
		"或者是它的分享連結.")
}

func sendUserOwnedS(c tele.Context) error {
	usq := queryUserS(c.Sender().ID)
	if usq == nil {
		return errors.New("no sticker owned")
	}

	var entries []string

	for _, us := range usq {
		date := time.Unix(us.timestamp, 0).Format("2006-01-02 15:04")
		title := strings.TrimSuffix(us.tg_title, " @"+botName)
		entry := fmt.Sprintf(`<a href="https://t.me/addstickers/%s">%s</a>`, us.tg_id, title)
		entry += " | " + date
		entries = append(entries, entry)
	}

	if len(entries) > 30 {
		eChunks := chunkSlice(entries, 30)
		for _, eChunk := range eChunks {
			message := "You own following stickers:\n"
			message += strings.Join(eChunk, "\n")
			c.Send(message, tele.ModeHTML)
		}
	} else {
		message := "You own following stickers:\n"
		message += strings.Join(entries, "\n")
		c.Send(message, tele.ModeHTML)
	}
	return nil
}

func chunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for {
		if len(slice) == 0 {
			break
		}

		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}
	return chunks
}

func sendAskEditChoice(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnAdd := selector.Data("Add sticker/增添貼圖", "add")
	btnDel := selector.Data("Delete sticker/刪除貼圖", "del")
	btnMov := selector.Data("Change order/調整順序", "mov")
	btnEmoji := selector.Data("Change emoji/修改Emoji", "emoji")
	btnDelset := selector.Data("Delete sticker set/刪除貼圖包", "delset")
	btnExit := selector.Data("Exit/退出", "bye")
	selector.Inline(selector.Row(btnAdd), selector.Row(btnDel), selector.Row(btnDelset), selector.Row(btnMov), selector.Row(btnEmoji), selector.Row(btnExit))

	return c.Send(fmt.Sprintf("<code>ID: %s</code>\n\n", users.data[c.Sender().ID].stickerData.id)+
		"What do you want to edit? Please select below:\n"+
		"您想要修改貼圖包的甚麼內容? 請選擇:", selector, tele.ModeHTML)
}

func sendAskSFrom(c tele.Context) error {
	return c.Send("Which sticker do you want to move? Please send it.\n" +
		"傳送您想要移動的那個貼圖.")
}

func sendAskMovTarget(c tele.Context) error {
	return c.Send("Where do you want to move the sticker to? Please send the sticker that holds the target position.\n" +
		"請傳送目標位置上所在的貼圖.")
}

func sendAskSDel(c tele.Context) error {
	return c.Send("Which sticker do you want to delete? Please send it.\n" +
		"您想要刪除哪一個貼圖? 請傳送那個貼圖")
}

func sendSEditEmoji(c tele.Context) error {
	return c.Send("Please the sticker that you want to edit.\n請傳送想要修改的貼圖")
}

func sendAskEmojiEdit(c tele.Context) error {
	return c.Send("Please send emoji(s)\n請傳送emoji(可以多個)")
}

func sendConfirmDelset(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnYes := selector.Data("Yes", "yes")
	btnNo := selector.Data("No", "no")
	selector.Inline(selector.Row(btnYes), selector.Row(btnNo))

	return c.Send("You are attempting to delete the whole sticker set, please confirm.\n"+
		"您將要刪除整個貼圖包, 請確認.", selector)

}

func sendSFromSS(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	id := ud.stickerData.id
	ss, _ := c.Bot().StickerSet(id)
	c.Send(&ss.Stickers[0])
	return nil
}

func sendFLWarning(c tele.Context) error {
	return c.Send("It might take longer to process this request (< 1min), please wait...\n" +
		"此貼圖可能需要更長時間處理(少於1分鐘), 請耐心等待...")
}

func sendTooManyFloodLimits(c tele.Context) error {
	return c.Send("Sorry, you have triggered Telegram's flood limit for too many times, it's recommended try again after a while.\n" +
		"抱歉, 您暫時超過了Telegram的貼圖製作次數限制, 建議您過一段時間後再試一次.")
}

func sendNoCbWarn(c tele.Context) error {
	return c.Send("Please press a button! /quit\n請選擇按鈕!")
}

func sendBadIDWarn(c tele.Context) error {
	return c.Send("Bad ID! try again or press Auto Generate. ID錯誤, 請試多一次或按下'自動生成'按鈕.")
}

func sendIDOccupiedWarn(c tele.Context) error {
	return c.Send("ID already occupied! try another one. ID已經被占用, 請試試另一個.")
}

func sendBadImportLinkWarn(c tele.Context) error {
	return c.Send("Invalid import link, make sure its a LINE Store link or kakao store link. Try again or /quit\n"+
		"無效的連結, 請檢視是否為LINE貼圖商店的連結, 或是kakao emoticon的連結.\n\n"+
		"For example: 例如:\n"+
		"<code>https://store.line.me/stickershop/product/7673/ja</code>\n"+
		"<code>https://e.kakao.com/t/pretty-all-friends</code>", tele.ModeHTML)
}

func sendNotifySDOnBackground(c tele.Context) error {
	return c.Send("Download has been started on the background. You can continue using other features. /start\n" +
		"下載任務已開始在背景處理, 您可以繼續使用bot的其他功能. /start")
}

func sendNoSToManage(c tele.Context) error {
	return c.Send("Sorry, you have not created any sticker set yet. You can use /import or /create .\n" +
		"抱歉, 您還未創建過貼圖包, 您可以使用 /create 或 /import .")
}

// func sendEditEmojiOK(c tele.Context) error {
// 	return c.Send("Edit emoji OK. Note that you might not be able to see the change immediately.\n" +
// 		"成功修改emoji. 更新後的emoji可能無法即刻看到, 需稍等其更新.")
// }

func sendPromptStopAdding(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnDone := selector.Data("Done adding/停止添加", "done")
	selector.Inline(selector.Row(btnDone))
	return c.Send("Continue sending files or press button below to stop adding.\n"+
		"請繼續傳送檔案. 或者按下方按鈕來停止增添.", selector)
}

func replySFileOK(c tele.Context, count int) error {
	selector := &tele.ReplyMarkup{}
	btnDone := selector.Data("Done adding/停止添加", "done")
	selector.Inline(selector.Row(btnDone))
	return c.Reply(
		fmt.Sprintf("File OK. Got %d stickers. Continue sending files or press button below to stop adding.\n"+
			"檔案OK. 已收到%d份貼圖. 請繼續傳送檔案. 或者按下方按鈕來停止增添.", count, count), selector)
}

func sendSEditOK(c tele.Context) error {
	return c.Send("It might take some time for your client to reflect the change if on iOS/macOS.\n" +
		"請知悉如果使用iOS/macOS, 您可能無法即刻看到修改後的變化, 可能需要稍等一下.")
}
