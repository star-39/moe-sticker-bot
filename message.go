package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

func sendStartMessage(c tele.Context) error {
	message := `
Hello! I'm moe_sticker_bot doing sticker stuffs!
Send me links or stickers to import or download them, or, use a command below:
你好! 歡迎使用萌萌貼圖BOT, 請傳送連結或貼圖給我來匯入或下載貼圖喔,
您也可以從下方選擇指令:

<b>/import</b> LINE stickers to Telegram<code>
  匯入LINE貼圖包至Telegram
</code>
<b>/download</b> Telegram sticker(s)<code>
  下載Telegram的貼圖包
</code>
<b>/create</b> new sticker set<code>
  創建新的Telegram的貼圖包.
</code>
<b>/manage</b> exsting sticker set<code>
  管理Telegram貼圖包(增添/刪除/排序).
</code>
<b>/faq  /about</b><code>
   常見問題/關於.
</code>
<b>/exit /quit /cancel</b><code>
  Interrupt conversation. 中斷指令.
</code>
`
	return c.Send(message, tele.ModeHTML)
}

func sendAboutMessage(c tele.Context) {
	c.Send(fmt.Sprintf(`
@%s by @plow283
https://github.com/star-39/moe-sticker-bot
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

func sendAskEmoji(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnRand := selector.Data("Random", "random")
	btnManu := selector.Data("Manual", "manual")
	selector.Inline(selector.Row(btnRand, btnManu))

	return c.Send(`
Please send an emoji representing all stickers in this sticker set,
to assign different emoji for each sticker, press Manual button below.
請傳送用於表示整個貼圖包的emoji,
如果想要為每個貼圖分別設定不同的emoji, 請按下Manual按鈕.
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

func sendAskWantSDown(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btn1 := selector.Data("Yes", "yes")
	selector.Inline(selector.Row(btn1))
	return c.Reply(`
You can download this sticker set. Press Yes to continue.
您可以下載這個貼圖包, 按下Yes來繼續.
`, selector)
}

func sendAskWantImport(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btn1 := selector.Data("Yes", "yes")
	selector.Inline(selector.Row(btn1))
	return c.Reply(`
You can import this sticker set. Press Yes to continue.
您可以匯入這個貼圖包, 按下Yes來繼續.
`, selector)
}

func sendAskTitle_Import(c tele.Context) {
	selector := &tele.ReplyMarkup{}
	btnAuto := selector.Data("Auto", "autoTitle")
	selector.Inline(selector.Row(btnAuto))

	c.Send("Please set a title for this sticker set. Press Auto button to set title from LINE Store as shown below:\n"+
		"請設定貼圖包的標題.按下Auto按鈕可以自動設為LINE Store中的標題如下:\n"+
		"スタンプのタイトルを送信してください。Autoボタンを押すと、LINE STOREに表記されているタイトルが設定されます。\n\n"+
		users.data[c.Sender().ID].lineData.title, selector)
}

func sendAskTitle(c tele.Context) {
	c.Send("Please set a title for this sticker set.\n" +
		"請設定貼圖包的標題.\n" +
		"スタンプのタイトルを送信してください。")
}

func sendAskImportLink(c tele.Context) error {
	return c.Send("Please send LINE store link of the sticker set\n" +
		"請傳送貼圖包的LINE Store連結.\n" +
		"スタンプのLINE Storeリンクを送信してください")
}

func sendNotifySExist(c tele.Context) error {
	links := queryLinksByLineID(users.data[c.Sender().ID].lineData.id)
	if len(links) > 0 {
		linkText := strings.Join(links, "\n")
		return c.Send(linkText)
	}
	return nil
}

func sendAskStickerFile(c tele.Context) error {
	return c.Send("Please send images/photos/stickers(less than 120 in total)(don't group items),\n" +
		"or send an archive containing image files,\n" +
		"wait until upload complete, then send a # mark.\n\n" +
		"請傳送任意格式的圖片/照片/貼圖(少於120張)(不要合併成組)\n" +
		"或者傳送內有貼圖檔案的歸檔,\n" +
		"請等候所有檔案上載完成, 然後傳送 # 記號\n")
}

func sendInStateWarning(c tele.Context) error {
	command := users.data[c.Sender().ID].command
	state := users.data[c.Sender().ID].state

	return c.Send(fmt.Sprintf("Please follow instructions.\n"+
		"Current command: %s\nCurrent state: %s\nYou can also send /quit to terminate session.", command, state))

}

func sendNoStateWarning(c tele.Context) error {
	return c.Send("Please use /start .")
}

func sendAskSTypeToCreate(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnStatic := selector.Data("Static", "static")
	btnAnimated := selector.Data("Animated", "video")
	selector.Inline(selector.Row(btnStatic, btnAnimated))
	return c.Send("What kind of sticker set you want to create?", selector)
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
	c.Send("<b>Fatal error! Please try again. /start\n"+
		"發生致命錯誤! 請您從頭再試一次. /start\n"+
		"致命的なエラーが発生しました！もう一度やり直してください /start </b>\n\n"+
		"<code>"+err.Error()+"</code>", tele.ModeHTML)
}

func sendProcessStarted(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	message := fmt.Sprintf(`
Preparing stickers, please wait...
正在準備貼圖, 請稍後...
作業が開始しています、少々お時間を...
<code>
LINE Cat:%s
LINE ID:%s
TG ID:%s
TG Title:%s
TG Link:</code>%s

<b>Progress / 進展</b>
<code>%s</code>
`, ud.lineData.category,
		ud.lineData.id,
		ud.stickerData.id,
		escapeTagMark(ud.stickerData.title),
		"https://t.me/addstickers/"+ud.stickerData.id, "")
	ud.progress = message

	teleMsg, err := c.Bot().Send(c.Recipient(), message, tele.ModeHTML)
	ud.progressMsg = teleMsg
	return err
}

func editProgressMsg(cur int, total int, sp string, c tele.Context) error {
	ud := users.data[c.Sender().ID]
	origin := ud.progress
	header := origin[:strings.LastIndex(origin, "<code>")]
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
	message := header + prog
	c.Bot().Edit(ud.progressMsg, message, tele.ModeHTML)
	return nil
}

func sendAskSToManage(c tele.Context) error {
	return c.Send("Send a sticker from the sticker set that want to edit,\n" +
		"or send its share link.\n\n" +
		"您想要修改哪個貼圖包? 請傳送那個貼圖包內任意一張貼圖,\n" +
		"或者是它的分享連結.")
}

func sendUserOwnedS(c tele.Context) error {
	ids, titles, timestamps := queryUserS(c.Sender().ID)
	if ids == nil {
		return errors.New("no sticker owned")
	}

	var entries []string

	for i, id := range ids {
		date := time.Unix(timestamps[i], 0).Format("2006-01-02 15:04")
		title := strings.TrimSuffix(titles[i], " @"+botName)
		entry := fmt.Sprintf(`<a href="https://t.me/addstickers/%s">%s</a>`, id, title)
		entry += " | " + date
		entries = append(entries, entry)
	}
	message := "You own following stickers:\n"
	message += strings.Join(entries, "\n")

	return c.Send(message, tele.ModeHTML)
}

func sendAskEditChoice(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnAdd := selector.Data("Add sticker/增添貼圖", "add")
	btnDel := selector.Data("Delete sticker/刪除貼圖", "del")
	btnDelset := selector.Data("Delete sticker set/刪除貼圖包", "delset")
	btnExit := selector.Data("Exit/退出", "bye")
	selector.Inline(selector.Row(btnAdd), selector.Row(btnDel), selector.Row(btnDelset), selector.Row(btnExit))

	return c.Send("What do you want to edit? Please select below:\n"+
		"您想要修改貼圖包的甚麼內容? 請選擇:", selector)
}

func sendAskSDel(c tele.Context) error {
	return c.Send("Which sticker do you want to delete? Please send it.\n" +
		"您想要刪除哪一個貼圖? 請傳送那個貼圖")
}

func sendConfirmDelset(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnYes := selector.Data("Yes", "yes")
	btnNo := selector.Data("No", "no")
	selector.Inline(selector.Row(btnYes), selector.Row(btnNo))

	return c.Send("You are attempting to delete the whole sticker set, please confirm.\n"+
		"您將要刪除整個貼圖包, 請確認.", selector)

}
