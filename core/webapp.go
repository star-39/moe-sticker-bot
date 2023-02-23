package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func InitWebAppServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	u, err := url.Parse(msbconf.WebappUrl)
	if err != nil {
		log.Error("Failed parsing WebApp URL! Consider disable --webapp ?")
		log.Fatalln(err.Error())
	}
	p := u.Path

	webappApi := r.Group(path.Join(p, "api"))
	{
		//Group: /webapp/api
		webappApi.POST("/initData", apiInitData)
		webappApi.GET("/ss", apiSS)
		webappApi.POST("/edit/result", apiEditResult)
		webappApi.POST("/edit/move", apiEditMove)
		webappApi.GET("/export", apiExport)
	}

	go func() {
		err := r.Run(msbconf.WebappListenAddr)
		if err != nil {
			log.Fatalln("WebApp: Gin Run failed! Check your addr or disable webapp.\n", err)
		}
		log.Infoln("WebApp: Listening on ", msbconf.WebappListenAddr)
	}()
}

func apiExport(c *gin.Context) {
	sn := c.Query("sn")
	qid := c.Query("qid")
	hex := c.Query("hex")
	url := fmt.Sprintf("msb://app/export/%s/?qid=%s&hex=%s", sn, qid, hex)
	c.Redirect(http.StatusFound, url)
}

type webappStickerSet struct {
	//Sticker objects
	SS []webappStickerObject `json:"ss"`
	//StickerSet Name
	SSName string `json:"ssname"`
	//StickerSet Title
	SSTitle string `json:"sstitle"`
	//StickerSet PNG Thumbnail
	SSThumb string `json:"ssthumb"`
	//Is Animated WebP
	Animated bool `json:"animated"`
	Amount   int  `json:"amount"`
	//Indicates that all sticker files are ready
	Ready bool `json:"ready"`
}

type webappStickerObject struct {
	//Sticker index with offset of +1
	Id int `json:"id"`
	//Sticker emojis.
	Emoji string `json:"emoji"`
	//Sticker emoji changed on front-end.
	EmojiChanged bool `json:"emoji_changed"`
	//Sticker file path on server.
	FilePath string `json:"file_path"`
	//Sticker file ID
	FileID string `json:"file_id"`
	//Sticker unique ID
	UniqueID string `json:"unique_id"`
	//URL of sticker image.
	Surl string `json:"surl"`
}

// GET <- ?uid&qid&sn&cmd
// -------------------------------------------
// -> [webappStickerObject, ...]
// -------------------------------------------
// id starts from 1 !!!!
// surl might be 404 when preparing stickers.
func apiSS(c *gin.Context) {
	cmd := c.Query("cmd")
	sn := c.Query("sn")
	uid := c.Query("uid")
	qid := c.Query("qid")
	hex := c.Query("hex")
	var ss *tele.StickerSet
	var err error

	switch cmd {
	case "edit":
		ud, err := checkGetUd(uid, qid)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		// Refresh SS data since it might already changed.
		retrieveSSDetails(ud.lastContext, ud.stickerData.id, ud.stickerData)
		ss = ud.stickerData.stickerSet
	case "export":
		if sn == "" || qid == "" {
			c.String(http.StatusBadRequest, "no_sn_or_qid")
			return
		}
		ss, err = b.StickerSet(sn)
		if err != nil {
			c.String(http.StatusBadRequest, "bad_sn")
			return
		}
	default:
		c.String(http.StatusBadRequest, "no_cmd")
		return
	}

	wss := webappStickerSet{
		SSTitle:  ss.Title,
		SSName:   ss.Name,
		Animated: ss.Video,
	}
	sl := []webappStickerObject{}
	ready := true
	for i, s := range ss.Stickers {
		var surl string
		var fpath string
		if s.Video {
			fpath = filepath.Join(msbconf.WebappDataDir, hex, s.SetName, s.UniqueID+".webm")
		} else {
			fpath = filepath.Join(msbconf.WebappDataDir, hex, s.SetName, s.UniqueID+".webp")
		}
		surl, _ = url.JoinPath(msbconf.WebappUrl, "data", hex, s.SetName, s.UniqueID+".webp")
		sl = append(sl, webappStickerObject{
			Id:       i + 1,
			Emoji:    s.Emoji,
			Surl:     surl,
			UniqueID: s.UniqueID,
			FileID:   s.FileID,
			FilePath: fpath,
		})
		if i == 0 {
			wss.SSThumb, _ = url.JoinPath(msbconf.WebappUrl, "data", hex, s.SetName, s.UniqueID+".png")
		}
		if st, _ := os.Stat(fpath); st == nil {
			ready = false
		}
	}
	wss.SS = sl
	wss.Ready = ready

	jsonWSS, err := json.Marshal(wss)
	if err != nil {
		log.Errorln("json marshal jsonWSS in apiSS error!")
		c.String(http.StatusInternalServerError, "json marshal jsonWSS in apiSS error!")
		return
	}
	c.String(http.StatusOK, string(jsonWSS))
}

// <- ?qid&qid&sha256sum  [{"index", "emoji", "surl"}, ...]
// -------------------------------------------
// -> STATUS
func apiEditResult(c *gin.Context) {
	uid := c.Query("uid")
	qid := c.Query("qid")
	body, _ := io.ReadAll(c.Request.Body)
	// if !validateSHA256(body, sum) {
	// 	c.String(http.StatusBadRequest, "bad result csum!")
	// 	return
	// }
	if string(body) == "" {
		//user did nothing
		return
	}
	so := []webappStickerObject{}
	err := json.Unmarshal(body, &so)
	if err != nil {
		c.String(http.StatusBadRequest, "bad_json")
		return
	}
	ud, err := checkGetUd(uid, qid)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if ud.state == ST_PROCESSING {
		c.String(http.StatusOK, "already processing...")
		return
	}

	log.Debugln(so)

	c.String(http.StatusOK, "")
	ud.udSetState(ST_PROCESSING)

	go func() {
		err := commitEmojiChange(ud, so)
		if err != nil {
			sendFatalError(err, ud.lastContext)
			endManageSession(ud.lastContext)
			endSession(ud.lastContext)
		}
	}()
}

func commitEmojiChange(ud *UserData, so []webappStickerObject) error {
	waitTime := 0
	for ud.webAppWorkerPool.Waiting() > 0 {
		time.Sleep(500 * time.Millisecond)
		waitTime++
		if waitTime > 20 {
			break
		}
	}
	ud.webAppWorkerPool.ReleaseTimeout(10 * time.Second)
	// retrieveSSDetails(ud.lastContext, ud.stickerData.id, ud.stickerData)
	//copy slice
	ss := ud.stickerData.stickerSet.Stickers
	notificationSent := false
	emojiChanged := false
	for _, s := range so {
		if s.EmojiChanged {
			emojiChanged = true
		}
	}
	if !emojiChanged {
		goto NEXT
	}
	for i, s := range ss {
		if s.UniqueID != so[i].UniqueID {
			log.Error("sticker order mismatch! index:", i)
			return errors.New("sticker order mismatch, no emoji change committed")
		}
		if !so[i].EmojiChanged {
			continue
		}
		oldEmoji := findEmojis(s.Emoji)
		newEmoji := findEmojis(so[i].Emoji)
		if newEmoji == "" || newEmoji == oldEmoji {
			log.Warn("webapp: ignored one invalid emoji.")
			continue
		}
		log.Debugln("Old:", i, s.Emoji, s.FileID)
		log.Debugln("New", i, newEmoji)
		if !notificationSent {
			sendEditingEmoji(ud.lastContext)
			notificationSent = true
		}

		err := editStickerEmoji(newEmoji, i, s.FileID, so[i].FilePath, len(ss), ud)
		if err != nil {
			return err
		}
		// Have a rest.
		time.Sleep(2 * time.Second)
	}
NEXT:
	sendSEditOK(ud.lastContext)
	sendSFromSS(ud.lastContext, ud.stickerData.id, nil)
	endManageSession(ud.lastContext)
	endSession(ud.lastContext)
	return nil
}

// <- ?uid&qid POST_FORM:{"oldIndex", "newIndex"}
// -------------------------------------------
// -> STATUS
func apiEditMove(c *gin.Context) {
	uid := c.Query("uid")
	qid := c.Query("qid")
	oldIndex, _ := strconv.Atoi(c.PostForm("oldIndex"))
	newIndex, _ := strconv.Atoi(c.PostForm("newIndex"))
	ud, err := checkGetUd(uid, qid)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	smo := &StickerMoveObject{
		wg:       sync.WaitGroup{},
		sd:       ud.stickerData,
		oldIndex: oldIndex,
		newIndex: newIndex,
	}
	smo.wg.Add(1)
	ud.webAppWorkerPool.Invoke(smo)
	smo.wg.Wait()
	if smo.err != nil {
		c.String(http.StatusInternalServerError, smo.err.Error())
		return
	}
}

func apiInitData(c *gin.Context) {
	//We must verify the initData before using it
	queryID := c.PostForm("query_id")
	authDate := c.PostForm("auth_date")
	user := c.PostForm("user")
	hash := c.PostForm("hash")
	dataCheckString := strings.Join([]string{
		"auth_date=" + authDate,
		"query_id=" + queryID,
		"user=" + user}, "\n")
	if !validateHMAC(dataCheckString, hash) {
		log.Warning("WebApp DCS HMAC failed, corrupt or attack?")
		c.String(http.StatusBadRequest, "data_check_string HMAC validation failed!!")
		return
	}
	log.Debug("WebApp initData DCS HMAC OK.")

	initWebAppRequest(c)
}

func initWebAppRequest(c *gin.Context) {
	user := c.PostForm("user")
	queryID := c.PostForm("query_id")
	cmd := c.Query("cmd")
	cmd = path.Base(cmd)
	webAppUser := &WebAppUser{}
	err := json.Unmarshal([]byte(user), webAppUser)
	if err != nil {
		log.Error("json unmarshal webappuser error.")
		return
	}

	switch cmd {
	case "edit":
		ud, err := GetUd(strconv.Itoa(webAppUser.Id))
		if err != nil {
			c.String(http.StatusBadRequest, "bad_state")
			return
		}
		ud.webAppWorkerPool, _ = ants.NewPoolWithFunc(1, wSubmitSMove)
		ud.webAppQID = queryID
	case "export":
		sn := c.Query("sn")
		hex := c.Query("hex")
		if sn == "" || hex == "" {
			c.String(http.StatusBadRequest, "no_sn_or_hex")
			return
		}
		ss, err := b.StickerSet(sn)
		if err != nil {
			c.String(http.StatusBadRequest, "bad_sn")
			return
		}
		// appendSStoQIDAuthList(sn, queryID)
		prepareWebAppExportStickers(ss, hex)
	default:
		c.String(http.StatusBadRequest, "bad_or_no_cmd")
		return
	}

	c.String(http.StatusOK, "webapp init ok")
}

// Telegram WebApp Regulation.
func validateHMAC(dataCheckString string, hash string) bool {
	// This calculated secret will be used to "decrypt" DCS
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(msbconf.BotToken))
	secretByte := h.Sum(nil)

	h = hmac.New(sha256.New, secretByte)
	h.Write([]byte(dataCheckString))
	dcsHash := fmt.Sprintf("%x", h.Sum(nil))
	return hash == dcsHash
}

// func validateSHA256(dataToCheck []byte, hash string) bool {
// 	h := sha256.New()
// 	h.Write(dataToCheck)
// 	csum := fmt.Sprintf("%x", h.Sum(nil))
// 	return hash == csum
// }

func checkGetUd(uid string, qid string) (*UserData, error) {
	ud, err := GetUd(uid)
	if err != nil {
		return nil, errors.New("no such user")
	}
	if ud.webAppQID != qid {
		return nil, errors.New("qid not valid")
	}
	return ud, nil
}

func prepareWebAppEditStickers(ud *UserData, wantHQ bool) error {
	dest := filepath.Join(msbconf.WebappDataDir, ud.stickerData.id)
	os.RemoveAll(dest)
	os.MkdirAll(dest, 0755)

	for _, s := range ud.stickerData.stickerSet.Stickers {
		var f string
		if ud.stickerData.stickerSet.Video {
			f = filepath.Join(dest, s.UniqueID+".webm")
		} else {
			f = filepath.Join(dest, s.UniqueID+".webp")
		}
		obj := &StickerDownloadObject{
			bot:       b,
			dest:      f,
			sticker:   s,
			forWebApp: true,
			webAppHQ:  wantHQ,
		}
		obj.wg.Add(1)
		ud.stickerData.sDnObjects = append(ud.stickerData.sDnObjects, obj)
		go wpDownloadStickerSet.Invoke(obj)
	}
	return nil
}

func prepareWebAppExportStickers(ss *tele.StickerSet, hex string) error {
	dest := filepath.Join(msbconf.WebappDataDir, hex, ss.Name)
	// If the user is reusing the generated link to export.
	// Do not re-download for every initData.
	stat, _ := os.Stat(dest)
	if stat != nil {
		mtime := stat.ModTime().Unix()
		// Less than 5 minutes, do not re-download
		if time.Now().Unix()-mtime < 300 {
			log.Debug("prepareWebAppExportStickers: dir still fresh, don't overwrite.")
			return nil
		}
	}
	os.RemoveAll(dest)
	os.MkdirAll(dest, 0755)

	for i, s := range ss.Stickers {
		var f string
		if ss.Video {
			f = filepath.Join(dest, s.UniqueID+".webm")
		} else {
			f = filepath.Join(dest, s.UniqueID+".webp")
		}
		obj := &StickerDownloadObject{
			bot:       b,
			dest:      f,
			sticker:   s,
			forWebApp: true,
			webAppHQ:  true,
		}
		//Use first image to create a thumbnail image
		//for WhatsApp.
		if i == 0 {
			obj.webAppThumb = true
		}
		obj.wg.Add(1)
		go wpDownloadStickerSet.Invoke(obj)
	}
	return nil
}

// func appendSStoQIDAuthList(sn string, qid string) {
// 	webAppSSAuthList.mu.Lock()
// 	defer webAppSSAuthList.mu.Unlock()

// 	obj := &WebAppQIDAuthObject{sn: sn, dt: time.Now().Unix()}
// 	webAppSSAuthList.sa[qid] = obj
// }

// func removeSSfromQIDAuthList(sn string, qid string) {
// 	webAppSSAuthList.mu.Lock()
// 	defer webAppSSAuthList.mu.Unlock()

// 	_, exist := webAppSSAuthList.sa[qid]
// 	if exist {
// 		delete(webAppSSAuthList.sa, qid)
// 	}
// }

// func veriyQIDfromAuthList(sn string, qid string) bool {
// 	sa := webAppSSAuthList.sa[qid]
// 	if sa == nil {
// 		return false
// 	}
// 	if sa.sn == sn {
// 		return true
// 	} else {
// 		return false
// 	}
// }
