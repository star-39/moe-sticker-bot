package core

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func sendStartMessage(c tele.Context) error {
	message := `
<b>/import</b>  <b>/search</b> LINE/Kakao stickers.<code>
匯入或搜尋LINE/Kaka貼圖包.</code>
<b>/download</b>  <b>/create</b>  <b>/manage</b> Telegram stickers.<code>
下載、創建、管理Telegram貼圖包.</code>
<b>/faq  /about  /changelog</b><code>
常見問題 / 關於 / 更新紀錄.</code>

Hello! I'm <a href="https://github.com/star-39/moe-sticker-bot">moe_sticker_bot</a>! Please:
• Send <b>LINE/Kakao sticker share link</b> to import or download.
• Send <b>Telegram sticker/link/GIF</b> to download or export to WhatsApp.
• Send <b>keywords</b> to search titles.
• Send <b>/create</b> or <b>/manage</b> to create or manage Telegram sticker set.
or use a command above.

你好, 歡迎使用<a href="https://github.com/star-39/moe-sticker-bot">萌萌貼圖BOT</a>! 請：
• 傳送<b>LINE/kakao貼圖包的分享連結</b>來匯入或下載.
• 傳送<b>Telegram貼圖/連結/GIF</b>來下載或匯出到WhatsApp.
• 傳送<b>關鍵字</b>來搜尋貼圖包.
• 傳送 <b>/create</b> 或 <b>/manage</b> 來創建或管理Telegram貼圖包。
或 從上方點選指令.
`
	return c.Send(message, tele.ModeHTML, tele.NoPreview)
}

func sendAboutMessage(c tele.Context) {
	c.Send(fmt.Sprintf(`
@%s by @plow283
<b>Please star for this project on Github if you like this bot!
如果您喜歡這個bot, 歡迎在Github給本專案標Star喔!
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
</b><code>
Version:版本: %s
</code>
`, botName, BOT_VERSION), tele.ModeHTML)
}

func sendFAQ(c tele.Context) {
	c.Send(fmt.Sprintf(`
@%s by @plow283
<b>Please hit Star for this project on Github if you like this bot!
如果您喜歡這個bot, 歡迎在Github給本專案標Star喔!
https://github.com/star-39/moe-sticker-bot</b>
------------------------------------
<b>Q: I got stucked! I can't quit from command!
我卡住了! 我沒辦法從指令中退出!</b>
A: Please send /quit to interrupt.
請傳送 /quit 來中斷.

<b>Q: Why ID has suffix: _by_%s ?
為甚麼ID的末尾有: _by_%s ?</b>
A: It's forced by Telegram, bot created sticker set must have its name in ID suffix.
因為這個是Telegram的強制要求, 由bot創造的貼圖ID末尾必須有bot名字.

<b>Q: Can I add video sticker to static sticker set or vice versa?
    我可以往靜態貼圖包加動態貼圖, 或者反之嗎?</b>
A: Yes, however, video will be static in static set
    可以. 惟動態貼圖在靜態貼圖包裡會變成靜態.

<b>Q: Who owns the sticker sets the bot created?
    BOT創造的貼圖包由誰所有?</b>
A: It's you of course. You can manage them through /manage or Telegram's official @Stickers bot.
    當然是您. 您可以通過 /manage 指令或者Telegram官方的 @Stickers 管理您的貼圖包.
`, botName, botName, botName), tele.ModeHTML)
}

func sendChangelog(c tele.Context) error {
	return c.Send(`
Details: 詳細:
https://github.com/star-39/moe-sticker-bot#changelog
v2.3.x (20230208)
  * Fix flood limit error during import.
  * Improved performance.
  * 修復匯入貼圖時flood limit錯誤。
  * 提升了性能.
  
v2.2.0 (20230131)
  * Support animated kakao sticker.
  * 支援動態kakao貼圖。

v2.1.0 (20230129)
  * Support exporting sticker to WhatsApp.
  * 支援匯出貼圖到WhatsApp

v2.0.0 (20230105)
  * Use new WebApp from /manage command to edit sticker set with ease.
  * Send text or use /search command to search imported LINE/kakao sticker sets by all users.
  * Auto import now happens on backgroud.
  * Downloading sticker set is now lot faster.
  * Fix many LINE import issues.
  * 通過 /manage 指令使用新的WebApp輕鬆管理貼圖包.
  * 直接傳送文字或使用 /search 指令來搜尋所有用戶匯入的LINE/KAKAO貼圖包.
  * 自動匯入現在會在背景處理.
  * 下載整個貼圖包的速度現在會快許多.
  * 修復了許多LINE貼圖匯入的問題.
	`, tele.NoPreview)
}

func sendAskEmoji(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnManu := selector.Data("Assign separately/分別設定", "manual")
	btnRand := selector.Data(`Batch assign as/一併設定為 "⭐"`, "random")
	selector.Inline(selector.Row(btnManu), selector.Row(btnRand))

	return c.Send(`
Telegram sticker requires emoji to represent it.
Press "Assign separately" to assign emoji one by one.
You can also do batch assign, send an emoji or press button below.
Telegram要求為貼圖設定emoji來表示它.
按下"分別設定"來為每個貼圖都分別設定相應的emoji.
您也可以一口氣為全部貼圖設定一樣的emoji, 請傳送一個emoji, 抑或是點選下方按鈕.
`, selector)
}

func sendConfirmExportToWA(c tele.Context, sn string, hex string) error {
	selector := &tele.ReplyMarkup{}
	baseUrl, _ := url.JoinPath(msbconf.WebappUrl, "export")
	webAppUrl := fmt.Sprintf("%s?sn=%s&hex=%s", baseUrl, sn, hex)
	log.Debugln("webapp export link is:", webAppUrl)
	webapp := tele.WebApp{URL: webAppUrl}
	btnExport := selector.WebApp("Continue export/繼續匯出 →", &webapp)
	selector.Inline(selector.Row(btnExport))

	return c.Reply(`
Exporting to WhatsApp requires <a href="https://github.com/star-39/msb_app">Msb App</a> being installed.
匯出到WhatsApp需要手機上安裝<a href="https://github.com/star-39/msb_app">Msb App</a>

<b>iPhone:</b> AppStore(N/A.暫無), <a href="https://github.com/star-39/msb_app/releases/latest/download/msb_app.ipa">IPA</a>
<b>Android:</b> GooglePlay(N/A.暫無), <a href="https://github.com/star-39/msb_app/releases/latest/download/msb_app.apk">APK</a>
`, tele.ModeHTML, tele.NoPreview, selector)
}

func genSDnMnEInline(canManage bool, sn string) *tele.ReplyMarkup {
	selector := &tele.ReplyMarkup{}
	btnSingle := selector.Data("Download this sticker/下載這張貼圖", CB_DN_SINGLE)
	btnAll := selector.Data("Download sticker set/下載整個貼圖包", CB_DN_WHOLE)
	btnMan := selector.Data("Manage sticker set/管理這個貼圖包", CB_MANAGE)
	btnExport := selector.Data("Export to WhatsApp/匯出到WhatsApp", CB_EXPORT_WA)
	if canManage {
		selector.Inline(selector.Row(btnSingle), selector.Row(btnAll),
			selector.Row(btnMan), selector.Row(btnExport))
	} else {
		selector.Inline(selector.Row(btnSingle), selector.Row(btnAll),
			selector.Row(btnExport))
	}
	return selector
}

func sendAskSDownloadChoice(c tele.Context, sn string) error {
	selector := genSDnMnEInline(false, sn)
	return c.Reply(`
You can download this sticker or the whole sticker set, please select below.
您可以下載這個貼圖或者其所屬的整個貼圖包, 請選擇:
`, selector)
}

func sendAskSChoice(c tele.Context, sn string) error {
	selector := genSDnMnEInline(true, sn)
	return c.Reply(`
You own this sticker set. You can download or manage this sticker set, please select below.
您擁有這個貼圖包. 您可以下載或者管理這個貼圖包, 請選擇:
`, selector)
}

func sendAskTGLinkChoice(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnManu := selector.Data("Download sticker set/下載整個貼圖包", CB_DN_WHOLE)
	btnMan := selector.Data("Manage sticker set/管理這個貼圖包", CB_MANAGE)
	selector.Inline(selector.Row(btnManu), selector.Row(btnMan))
	return c.Reply(`
You own this sticker set. You can download or manage this sticker set, please select below.
您擁有這個貼圖包. 您可以下載或者管理這個貼圖包, 請選擇:
`, selector)
}

func sendAskWantSDown(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btn1 := selector.Data("Yes", CB_DN_WHOLE)
	btnNo := selector.Data("No", CB_BYE)
	selector.Inline(selector.Row(btn1), selector.Row(btnNo))
	return c.Reply(`
You can download this sticker set. Press Yes to continue.
您可以下載這個貼圖包, 按下Yes來繼續.
`, selector)
}

func sendAskWantImportOrDownload(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btn1 := selector.Data("Import to Telegram/匯入到Telegram", CB_OK_IMPORT)
	btn2 := selector.Data("Download/下載", CB_OK_DN)
	// btnNo := selector.Data("Bye", CB_BYE)
	selector.Inline(selector.Row(btn1), selector.Row(btn2))
	return c.Reply(`
You can import or download this sticker set. Please choose.
您可以匯入或下載這個貼圖包, 請選擇.
`, selector)
}

func sendAskWhatToDownload(c tele.Context) error {
	return c.Send("Please send a sticker that you want to download, or its share link(can be either Telegram or LINE ones)\n" +
		"請傳送想要下載的貼圖, 或者是貼圖包的分享連結(可以是Telegram或LINE連結).")
}

func sendAskTitle_Import(c tele.Context) error {
	ld := users.data[c.Sender().ID].lineData
	ld.TitleWg.Wait()
	log.Debug("titles are::")
	log.Debugln(ld.I18nTitles)
	selector := &tele.ReplyMarkup{}

	var titleButtons []tele.Row
	var titleText string
	for i, t := range ld.I18nTitles {
		if t == "" {
			continue
		}
		title := escapeTagMark(t) + " @" + botName
		btn := selector.Data(title, strconv.Itoa(i))
		row := selector.Row(btn)
		titleButtons = append(titleButtons, row)
		titleText = titleText + "\n<code>" + title + "</code>"
	}

	if len(titleButtons) == 0 {
		btnDefault := selector.Data(escapeTagMark(ld.Title)+" @"+botName, CB_DEFAULT_TITLE)
		titleButtons = []tele.Row{selector.Row(btnDefault)}
	}
	selector.Inline(titleButtons...)

	return c.Send("Please send a title for this sticker set. You can also select a appropriate original title below:\n"+
		"請傳送貼圖包的標題.您也可以按下面的按鈕自動填上合適的原版標題:\n"+
		titleText, selector, tele.ModeHTML)
}

func sendAskTitle(c tele.Context) {
	c.Send("Please set a title for this sticker set.\n" +
		"請設定貼圖包的標題.")
}

func sendAskID(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnAuto := selector.Data("Auto Generate/自動生成", "auto")
	selector.Inline(selector.Row(btnAuto))
	return c.Send(`
Please set an ID for sticker set, used in share link.
Can contain only english letters, digits and underscores.
Must begin with a letter, can't contain consecutive underscores.
請設定貼圖包的ID, 用於分享連結.
ID只可以由英文字母, 數字, 下劃線記號組成, 由英文字母開頭, 不可以有連續下劃線記號.",
For example: 例如:
<code>My_favSticker21</code>

This is usually not important, it's recommended to press "Auto Generate" button.
ID通常不重要, 建議您按下下方的"自動生成"按鈕.`, selector, tele.ModeHTML)
}

func sendAskImportLink(c tele.Context) error {
	return c.Send(`
Please send LINE/kakao store link of the sticker set. You can obtain this link from App by going to sticker store and tapping Share->Copy Link.
請傳送貼圖包的LINE/kakao Store連結. 您可以在App裡的貼圖商店按右上角的分享->複製連結來取得連結.
For example: 例如:
<code>https://store.line.me/stickershop/product/7673/ja</code>
<code>https://e.kakao.com/t/pretty-all-friends</code>
<code>https://emoticon.kakao.com/items/lV6K2fWmU7CpXlHcP9-ysQJx9rg=?referer=share_link</code>
`, tele.ModeHTML)
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
		if l.Ae {
			entries = append(entries, fmt.Sprintf(`<a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title))
		} else {
			// append to top
			entries = append([]string{fmt.Sprintf(`★ <a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title)}, entries...)
		}
	}
	if len(entries) > 5 {
		entries = entries[:5]
	}
	message += strings.Join(entries, "\n")
	c.Send(message, tele.ModeHTML)
	return true
}

func sendSearchResult(entriesWant int, lines []LineStickerQ, c tele.Context) error {
	var entries []string
	message := "Search Results: 搜尋結果：\n"

	for _, l := range lines {
		l.Tg_title = strings.TrimSuffix(l.Tg_title, " @"+botName)
		if l.Ae {
			entries = append(entries, fmt.Sprintf(`<a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title))
		} else {
			// append to top
			entries = append([]string{fmt.Sprintf(`★ <a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title)}, entries...)
		}
	}

	if entriesWant == -1 && len(entries) > 120 {
		c.Send("Too many results, please narrow your keyword, truncated to 120 entries.\n" +
			"搜尋結果過多，已縮減到120個，請使用更準確的搜尋關鍵字。")
		entries = entries[:120]
	}
	if entriesWant != -1 && len(entries) > entriesWant {
		entries = entries[:entriesWant]
	}
	if len(entries) > 30 {
		eChunks := chunkSlice(entries, 30)
		for _, eChunk := range eChunks {
			msgToSend := message + strings.Join(eChunk, "\n")
			c.Send(msgToSend, tele.ModeHTML)
		}
	} else {
		message += strings.Join(entries, "\n")
		c.Send(message, tele.ModeHTML)
	}

	return nil
}

func sendAskStickerFile(c tele.Context) error {

	if users.data[c.Sender().ID].stickerData.isVideo {
		c.Send("Please send images/photos/stickers/videos(less than 50 in total),\n" +
			"or send an archive containing image files,\n" +
			"wait until upload complete, then tap 'Done adding'.\n\n" +
			"請傳送任意格式的圖片/照片/貼圖/影片(少於50張)\n" +
			"或者傳送內有貼圖檔案的歸檔,\n" +
			"請等候所有檔案上載完成, 然後按下「停止增添」\n")
		c.Send("Special note: Sending GIF with transparent background will lose transparency due to client issue.\n" +
			"You can compress your GIF into a ZIP file then send it to bot to bypass.\n" +
			"特別提示: 傳送帶有透明背景的GIF會被Telegram客戶端強制轉換並且丟失透明層.\n" +
			"您可以將貼圖放入ZIP歸檔中再傳送給bot來繞過這個限制.")
	} else {
		c.Send("Please send images/photos/stickers(less than 120 in total),\n" +
			"or send an archive containing image files,\n" +
			"wait until upload complete, then tap 'Done adding'.\n\n" +
			"請傳送任意格式的圖片/照片/貼圖(少於120張)\n" +
			"或者傳送內有貼圖檔案的歸檔,\n" +
			"請等候所有檔案上載完成, 然後按下「停止增添」\n")
	}
	return nil
}

func sendInStateWarning(c tele.Context) error {
	command := users.data[c.Sender().ID].command
	state := users.data[c.Sender().ID].state

	return c.Send(fmt.Sprintf(`
Please send content according to instructions.
請按照bot提示傳送相應內容.
Current command: %s
Current state: %s
You can also send /quit to terminate session.
您也可以傳送 /quit 來中斷對話.
`, command, state))
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
	if c == nil {
		return
	}
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
		if strings.Contains(errMsg, "500") {
			errMsg += "\nThis is an internal error of Telegram server, we could do nothing but wait for its recover. Please try again later.\n" +
				"此錯誤為Telegram伺服器之內部錯誤, 無法由bot解決, 只能等候官方修復. 建議您稍後再嘗試一次.\n"
		}
	}

	c.Send("<b>Fatal error! Please try again. /start\n"+
		"發生嚴重錯誤! 請您從頭再試一次. /start </b>\n\n"+
		"You can report this error to @plow283 or https://github.com/star-39/moe-sticker-bot/issues\n\n"+
		"<code>"+errMsg+"</code>", tele.ModeHTML, tele.NoPreview)
}

// Return:
// string: Text of the message.
// *tele.Message: The pointer of the message.
// error: error
func sendProcessStarted(ud *UserData, c tele.Context, optMsg string) (string, *tele.Message, error) {
	message := fmt.Sprintf(`
Preparing stickers, please wait...
正在準備貼圖, 請稍後...
<code>
LINE Cat:%s
LINE ID:%s
TG ID:%s
TG Title:</code><a href="%s">%s</a>

<b>Progress / 進展</b>
<code>%s</code>
`, ud.lineData.Category,
		ud.lineData.Id,
		ud.stickerData.id,
		"https://t.me/addstickers/"+ud.stickerData.id,
		escapeTagMark(ud.stickerData.title),
		optMsg)
	ud.progress = message

	teleMsg, err := c.Bot().Send(c.Recipient(), message, tele.ModeHTML)
	ud.progressMsg = teleMsg
	return message, teleMsg, err
}

// if progressText is empty, a progress bar will be generated based on cur and total.
func editProgressMsg(cur int, total int, progressText string, originalText string, teleMsg *tele.Message, c tele.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("editProgressMsg encountered panic! ignoring...", string(debug.Stack()))
		}
	}()

	header := originalText[:strings.LastIndex(originalText, "<code>")]
	prog := ""

	if progressText != "" {
		prog = progressText
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
		//workaround for empty title.
		if title == "" || title == " " {
			title = "_"
		}
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

func sendAskEditChoice(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	selector := &tele.ReplyMarkup{}
	btnAdd := selector.Data("Add sticker/增添貼圖", CB_ADD_STICKER)
	btnDel := selector.Data("Delete sticker/刪除貼圖", CB_DELETE_STICKER)
	btnDelset := selector.Data("Delete sticker set/刪除貼圖包", CB_DELETE_STICKER_SET)
	btnChangeTitle := selector.Data("Change title/修改標題", CB_CHANGE_TITLE)
	btnExit := selector.Data("Exit/退出", "bye")

	if msbconf.WebApp {
		baseUrl, _ := url.JoinPath(msbconf.WebappUrl, "edit")
		url := fmt.Sprintf("%s?ss=%s&dt=%d",
			baseUrl,
			ud.stickerData.id,
			time.Now().Unix())
		log.Debugln("WebApp URL is : ", url)
		webApp := &tele.WebApp{
			URL: url,
		}
		btnEdit := selector.WebApp("Change order or emoji/修改順序或Emoji", webApp)
		selector.Inline(
			selector.Row(btnAdd), selector.Row(btnDel), selector.Row(btnDelset), selector.Row(btnEdit), selector.Row(btnChangeTitle), selector.Row(btnExit))
	} else {
		selector.Inline(
			selector.Row(btnAdd), selector.Row(btnDel), selector.Row(btnDelset), selector.Row(btnChangeTitle), selector.Row(btnExit))
	}

	return c.Send(fmt.Sprintf(`
ID: <code>%s</code>
Title: <a href="https://t.me/addstickers/%s">%s</a>

What do you want to edit? Please select below:
您想要修改貼圖包的甚麼內容? 請選擇:`,
		users.data[c.Sender().ID].stickerData.id,
		ud.stickerData.id,
		ud.stickerData.title),
		selector, tele.ModeHTML)
}

func sendAskSDel(c tele.Context) error {
	return c.Send("Which sticker do you want to delete? Please send it.\n" +
		"您想要刪除哪一個貼圖? 請傳送那個貼圖")
}

func sendConfirmDelset(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnYes := selector.Data("Yes", CB_YES)
	btnNo := selector.Data("No", CB_NO)
	selector.Inline(selector.Row(btnYes), selector.Row(btnNo))

	return c.Send("You are attempting to delete the whole sticker set, please confirm.\n"+
		"您將要刪除整個貼圖包, 請確認.", selector)
}

func sendSFromSS(c tele.Context, ssid string, reply *tele.Message) error {
	ss, _ := c.Bot().StickerSet(ssid)
	if reply != nil {
		c.Bot().Reply(reply, &ss.Stickers[0])
	} else {
		c.Send(&ss.Stickers[0])
	}
	return nil
}

func sendFLWarning(c tele.Context) error {
	return c.Send(`
It might take longer to process this sticker set (2-8 minutes)... 
This warning indicates that you might triggered Telegram's flood limit, and bot is trying to re-submit.
Due to this mechanism, resulted sticker set might contains duplicate or missing sticker, please check manually after done.

此貼圖包可能需要更長時間處理(2-8分鐘)...
看到這一條警告表示Telegram可能限制了您創建貼圖包的頻度, 且bot正在自動嘗試重新製作, 因此得出的貼圖包可能會重複或缺失貼圖, 請在完成製作後再檢查一下.
`)
}

func sendTooManyFloodLimits(c tele.Context) error {
	return c.Send("Sorry, it seems that you have triggered Telegram's flood limit for too many times, it's recommended try again after a while.\n" +
		"抱歉, 您似乎觸發了Telegram的貼圖製作次數限制, 建議您過一段時間後再試一次.")
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

func sendNotifyWorkingOnBackground(c tele.Context) error {
	return c.Send("Work has been started on the background. You can continue using other features. /start\n" +
		"工作已開始在背景處理, 您可以繼續使用bot的其他功能. /start")
}

func sendNoSToManage(c tele.Context) error {
	return c.Send("Sorry, you have not created any sticker set yet. You can use /import or /create .\n" +
		"抱歉, 您還未創建過貼圖包, 您可以使用 /create 或 /import .")
}

func sendPromptStopAdding(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnDone := selector.Data("Done adding/停止添加", CB_DONE_ADDING)
	selector.Inline(selector.Row(btnDone))
	return c.Send("Continue sending files or press button below to stop adding.\n"+
		"請繼續傳送檔案. 或者按下方按鈕來停止增添.", selector)
}

func replySFileOK(c tele.Context, count int) error {
	selector := &tele.ReplyMarkup{}
	btnDone := selector.Data("Done adding/停止添加", CB_DONE_ADDING)
	selector.Inline(selector.Row(btnDone))
	return c.Reply(
		fmt.Sprintf("File OK. Got %d stickers. Continue sending files or press button below to stop adding.\n"+
			"檔案OK. 已收到%d份貼圖. 請繼續傳送檔案. 或者按下方按鈕來停止增添.", count, count), selector)
}

func sendSEditOK(c tele.Context) error {
	return c.Send(
		"Successfully edited sticker set. /start\n" +
			"成功修改貼圖包. /start")
}

func sendStickerSetFullWarning(c tele.Context) error {
	return c.Send(
		"Warning: Your sticker set is already full. You cannot add new sticker or edit emoji.\n" +
			"提示：當前貼圖包已滿，您將不能增添貼圖和修改emoji。")
}

func sendEditingEmoji(c tele.Context) error {
	return c.Send("Commiting changes...\n正在套用變更，請稍候...")
}

func sendAskSearchKeyword(c tele.Context) error {
	return c.Send("Please send a word that you want to search\n請傳送想要搜尋的內容")
}

func sendSearchNoResult(c tele.Context) error {
	message := "Sorry, no result.\n抱歉, 搜尋沒有結果."
	if c.Chat().Type == tele.ChatPrivate {
		message += "\nTry again or /quit\n請試試別的關鍵字或 /quit"
	}
	return c.Send(message)
}

func sendNotifyNoSessionSearch(c tele.Context) error {
	return c.Send("Here are some search results, use /search to dig deeper or /start to see available commands.\n" +
		"這些是貼圖包搜尋結果，使用 /search 詳細搜尋或 /start 來看看可用的指令。")
}

func sendUnsupportedCommandForGroup(c tele.Context) error {
	return c.Send("This command is not supported in group chat, please chat with bot directly.\n" +
		"此指令無法於群組內使用, 請與bot直接私訊.")
}

func sendBadSearchKeyword(c tele.Context) error {
	return c.Send(fmt.Sprintf(`
Please specify keyword
請指定搜尋關鍵字.

Example: 例如:
/search@%s keyword1 keyword2 ...
/search@%s nekomimi mia
`, botName, botName))
}

func sendNeedKakaoAnimatedShareLinkWarning(c tele.Context) error {
	msg := `
Importing animated kakao stickers requires a share link from KakaoTalk app.
You can still continue import, in that case, static ones will be imported.
You can obtain share link from sticker store in KakaoTalk app by tapping share->copy link.

此貼圖包含有動態貼圖，您需要傳送KakaoTalk app分享連結來匯入動態貼圖。
您也可以繼續匯入，但是匯入的貼圖將會是靜態。
您可以在KakaoTalk App內的貼圖商店點選 分享->複製連結 來取得連結。

eg: <code>https://emoticon.kakao.com/items/lV6K2fWmU7CpXlHcP9-ysQJx9rg=?referer=share_link</code>
`
	err := c.Reply(&tele.Photo{
		File:    tele.File{FileID: FID_KAKAO_SHARE_LINK},
		Caption: msg,
	}, tele.ModeHTML)
	if err != nil {
		c.Reply(msg, tele.ModeHTML)
	}
	return nil
}

func sendUseCommandToImport(c tele.Context) error {
	return c.Send("Please use /create to create sticker set using your own photos and videos. /start\n" +
		"請使用 /create 指令來使用自己的圖片和影片和創建貼圖包. /start")
}

func sendOneStickerFailedToAdd(c tele.Context, pos int, err error) error {
	return c.Reply(fmt.Sprintf(`
Failed to add one sticker.
一張貼圖添加失敗.
Index: %d
Error: %s
`, pos, err.Error()))
}

func sendBadSNWarn(c tele.Context) error {
	return c.Reply("Wrong sticker or link!\n貼圖或連結錯誤!")
}

func sendHowToChangeSSTitle(c tele.Context) error {
	sid := users.data[c.Sender().ID].stickerData.id
	msg := fmt.Sprintf(`
Currently, changing title can only be done through Telegram's official @Stickers bot.
Please chat with @Stickers bot, then 
1. Send <code>/renamepack</code>
2. Send <code>%s</code>
3. Send your new title.

修改貼圖包的標題目前只能通過Telegram官方的 @Stickers 進行，
請您與 @Stickers 對話，然後：
1. 傳送 <code>/renamepack</code>
2. 傳送 <code>%s</code>
3. 傳送您想設定的新標題。`, sid, sid)

	// err := c.Send(&tele.Photo{
	// 	File:    tele.File{FileID: FID_CHANGE_TITLE_TUTORIAL},
	// 	Caption: msg,
	// }, tele.ModeHTML)
	return c.Reply(msg, tele.ModeHTML)
}
