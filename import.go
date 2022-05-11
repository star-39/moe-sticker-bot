package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

func parseImportLink(link string, ld *LineData) error {
	u, err := url.Parse(link)
	if err != nil {
		return err
	}
	switch {
	case strings.HasSuffix(u.Host, "line.me"):
		ld.store = "line"
		return parseLineLink(link, ld)
	case strings.HasSuffix(u.Host, "kakao.com"):
		ld.store = "kakao"
		return parseKakaoLink(link, ld)
	default:
		return errors.New("unknow import")
	}
}

func parseKakaoLink(link string, ld *LineData) error {
	kakaoID := path.Base(link)
	apiUrl := "https://e.kakao.com/api/v1/items/t/" + kakaoID
	page, err := httpGet(apiUrl)
	if err != nil {
		return err
	}

	type KakaoJsonResult struct {
		Title         string
		ThumbnailUrls []string
	}
	type KakaoJson struct {
		Result KakaoJsonResult
	}

	var kakaoJson KakaoJson
	err = json.Unmarshal([]byte(page), &kakaoJson)
	if err != nil {
		log.Errorln("Failed json parsing kakao link!", err)
		return err
	}

	log.Debugln("Parsed kakao link:", link)
	log.Debugln(kakaoJson.Result)
	t := kakaoJson.Result.Title
	i := strings.ReplaceAll(kakaoID, "-", "_")

	ld.dLinks = kakaoJson.Result.ThumbnailUrls
	ld.title = t
	ld.id = i
	ld.link = link
	ld.amount = len(ld.dLinks)
	return nil
}

func parseLineLink(link string, ld *LineData) error {
	page, err := httpGet(link)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		log.Errorln("Failed gq parsing line link!", err)
		return err
	}

	var jsonData map[string]interface{}
	// For LINE STORE, the first script is always sticker's metadata in JSON.
	err = json.Unmarshal([]byte(doc.Find("script").First().Text()), &jsonData)
	if err != nil {
		log.Errorln("Failed json parsing line link!", err)
		return err
	}
	t := jsonData["name"].(string)
	i := jsonData["sku"].(string)
	u := jsonData["url"].(string)
	a := false
	c := ""
	d := "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/iphone/"

	if strings.Contains(u, "stickershop") {
		if strings.Contains(page, "MdIcoPlay_b") || strings.Contains(page, "MdIcoAni_b") {
			c = LINE_STICKER_ANIMATION
			d += "stickerpack@2x.zip"
			a = true
		} else if strings.Contains(page, "MdIcoMessageSticker_b") {
			d = u
			c = LINE_STICKER_MESSAGE
		} else if strings.Contains(page, "MdIcoNameSticker_b") {
			d += "sticker_name_base@2x.zip"
			c = LINE_STICKER_NAME
		} else if strings.Contains(page, "MdIcoFlash_b") || strings.Contains(page, "MdIcoFlashAni_b") {
			c = LINE_STICKER_POPUP
			d += "stickerpack@2x.zip"
			a = true
		} else if strings.Contains(page, "MdIcoEffectSoundSticker_b") || strings.Contains(page, "MdIcoEffectSticker_b") {
			c = LINE_STICKER_POPUP_EFFECT
			d += "stickerpack@2x.zip"
			a = true
		} else {
			c = LINE_STICKER_STATIC
			d += "stickers@2x.zip"
		}
	} else if strings.Contains(u, "emojishop") {
		if strings.Contains(page, "MdIcoPlay_b") {
			c = LINE_EMOJI_ANIMATION
			d = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + i + "/iphone/package_animation.zip"
			a = true
		} else {
			c = LINE_EMOJI_STATIC
			d = "https://stickershop.line-scdn.net/sticonshop/v1/sticon/" + i + "/iphone/package.zip"
		}
	} else {
		return errors.New("unknown line store category")
	}
	if ld == nil {
		return nil
	}
	ld.link = u
	ld.category = c
	ld.dLink = d
	ld.id = i
	ld.title = t
	ld.isAnimated = a
	log.Debugln("line data parsed:", ld)
	return nil
}

func prepImportStickers(ud *UserData, needConvert bool) error {
	switch ud.lineData.store {
	case "line":
		return prepLineStickers(ud, needConvert)
	case "kakao":
		return prepKakaoStickers(ud, needConvert)
	}
	return nil
}

func prepKakaoStickers(ud *UserData, needConvert bool) error {
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	ud.stickerData.id = "kakao_" + ud.lineData.id + secHex(2) + "_by_" + botName
	ud.stickerData.title = ud.lineData.title + " @" + botName

	workDir := filepath.Join(ud.userDir, ud.lineData.id)
	os.MkdirAll(workDir, 0755)
	for range ud.lineData.dLinks {
		sf := &StickerFile{}
		sf.wg.Add(1)
		ud.stickerData.stickers = append(ud.stickerData.stickers, sf)
		ud.stickerData.lAmount += 1
	}
	for i, l := range ud.lineData.dLinks {
		select {
		case <-ud.ctx.Done():
			log.Warn("prepKakaoStickers received ctxDone!")
			return nil
		default:
		}
		f := filepath.Join(workDir, path.Base(l))
		err := httpDownload(l, f)
		if err != nil {
			return err
		}
		cf, _ := imToWebp(f)
		ud.lineData.files = append(ud.lineData.files, f)
		ud.stickerData.stickers[i].oPath = f
		ud.stickerData.stickers[i].cPath = cf
		ud.stickerData.stickers[i].wg.Done()

		log.Debug("Done process one kakao emoticon")
		log.Debugf("f:%s, cf:%s", f, cf)
	}
	log.Debug("Done process ALL kakao emoticons")
	return nil
}

func prepLineStickers(ud *UserData, needConvert bool) error {
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	ud.stickerData.isVideo = ud.lineData.isAnimated
	ud.stickerData.id = "line_" + ud.lineData.id + secNum(4) + "_by_" + botName
	ud.stickerData.title = ud.lineData.title + " @" + botName

	if ud.lineData.category == LINE_STICKER_MESSAGE {
		return prepLineMessageS(ud)
	}

	workDir := filepath.Join(ud.userDir, ud.lineData.id)
	savePath := filepath.Join(workDir, "line.zip")
	os.MkdirAll(workDir, 0755)

	err := fDownload(ud.lineData.dLink, savePath)
	if err != nil {
		return err
	}

	pngFiles := lineZipExtract(savePath, ud.lineData)
	if len(pngFiles) == 0 {
		return errors.New("no line image")
	}

	ud.lineData.files = pngFiles
	ud.lineData.amount = len(pngFiles)
	ud.stickerData.lAmount = len(pngFiles)

	for _, f := range pngFiles {
		sf := &StickerFile{oPath: f}
		sf.wg = sync.WaitGroup{}
		ud.stickerData.stickers = append(ud.stickerData.stickers, sf)
	}

	if needConvert {
		log.Debugln("start converting...")
		doConvert(ud)
	}

	log.Debug("Done preparing line files:")
	log.Debugln(ud.lineData, ud.stickerData)

	return nil
}

func lineZipExtract(f string, ld *LineData) []string {
	var files []string
	workDir := fExtract(f)
	if workDir == "" {
		return nil
	}
	log.Debugln("scanning workdir:", workDir)

	switch ld.category {
	case LINE_STICKER_ANIMATION:
		files, _ = filepath.Glob(filepath.Join(workDir, "animation@2x", "*.png"))
	case LINE_STICKER_POPUP:
		files, _ = filepath.Glob(filepath.Join(workDir, "popup", "*.png"))
	case LINE_STICKER_POPUP_EFFECT:
		pfs, _ := filepath.Glob(filepath.Join(workDir, "popup", "*.png"))
		for _, pf := range pfs {
			os.Rename(pf, filepath.Join(workDir, strings.TrimSuffix(filepath.Base(pf), ".png")+"@99.png"))
		}
		files = lsFiles(workDir, []string{".png"}, []string{"tab", "key", "json"})
	default:
		files = lsFiles(workDir, []string{".png"}, []string{"tab", "key", "json"})
	}

	return files
}

func doConvert(ud *UserData) {
	sf := ud.stickerData.stickers
	for _, s := range sf {
		select {
		case <-ud.ctx.Done():
			log.Warn("doConvert received ctxDone!")
			return
		default:
		}
		var err error
		s.wg.Add(1)
		// If lineS is animated, commit to worker pool
		// since encoding vp9 is time and resource costy.
		if ud.lineData.isAnimated {
			wpConvertWebm.Invoke(s)
		} else {
			s.cPath, err = imToWebp(s.oPath)
			if err != nil {
				s.cError = err
			}
			s.wg.Done()
		}
	}
}

func prepLineMessageS(ud *UserData) error {
	workDir := filepath.Join(ud.userDir, ud.lineData.id)
	os.MkdirAll(workDir, 0755)

	page, err := httpGet(ud.lineData.link)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		log.Errorln("Failed gq parsing line link!", err)
		return err
	}

	var baseImages []string
	var overlayImages []string
	var jsonData map[string]interface{}
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		jsonDP, exist := s.Attr("data-preview")
		if !exist {
			return
		}
		log.Debugln("Got one json data-preview:", jsonDP)

		err := json.Unmarshal([]byte(jsonDP), &jsonData)
		if err != nil {
			log.Warnln("Json parse failed!", jsonDP)
			return
		}
		baseImages = append(baseImages, jsonData["staticUrl"].(string))
		overlayImages = append(overlayImages, jsonData["customOverlayUrl"].(string))
	})
	log.Debugln("base images:", baseImages)
	log.Debugln("overlay images:", overlayImages)

	for range baseImages {
		sf := &StickerFile{}
		sf.wg = sync.WaitGroup{}
		sf.wg.Add(1)
		ud.stickerData.stickers = append(ud.stickerData.stickers, sf)
	}

	ud.lineData.amount = len(baseImages)
	ud.stickerData.lAmount = ud.lineData.amount

	for i, b := range baseImages {
		select {
		case <-ud.ctx.Done():
			log.Warn("prepLineMessageS received ctxDone!")
			return nil
		default:
		}
		log.Debugln("Preparing one message sticker... index:", i)
		bPath := filepath.Join(workDir, strconv.Itoa(i)+".base.png")
		oPath := filepath.Join(workDir, strconv.Itoa(i)+".overlay.png")
		httpDownload(b, bPath)
		httpDownload(overlayImages[i], oPath)
		f, err := imStackToWebp(bPath, oPath)
		if err != nil {
			return err
		}
		ud.lineData.files = append(ud.lineData.files, f)
		ud.stickerData.stickers[i].oPath = f
		ud.stickerData.stickers[i].cPath = f
		ud.stickerData.stickers[i].wg.Done()
		log.Debugln("one message sticker OK:", f)
	}

	return nil
}
