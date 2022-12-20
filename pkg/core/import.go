package core

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
	url, _ := url.Parse(link)
	kakaoID := path.Base(url.Path)

	apiUrl := "https://e.kakao.com/api/v1/items/t/" + kakaoID
	page, err := httpGet(apiUrl)
	if err != nil {
		return err
	}

	var kakaoJson KakaoJson
	err = json.Unmarshal([]byte(page), &kakaoJson)
	if err != nil {
		log.Errorln("Failed json parsing kakao link!", err)
		return err
	}
	kakaoID = kakaoJson.Result.TitleUrl

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

func fetchLineI18nLinks(doc *goquery.Document) []string {
	var i18nLinks []string
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		hreflang, exist := s.Attr("hreflang")
		if !exist {
			return
		}
		href, exist2 := s.Attr("href")
		if !exist2 {
			return
		}
		switch hreflang {
		case "zh-Hant":
			fallthrough
		case "ja":
			fallthrough
		case "en":
			i18nLinks = append(i18nLinks, href)
		}
	})
	return i18nLinks
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

	var lineJson LineJson
	// For LINE STORE, the first script is always sticker's metadata in JSON.
	err = json.Unmarshal([]byte(doc.Find("script").First().Text()), &lineJson)
	if err != nil {
		log.Errorln("Failed json parsing line link!", err)
		return err
	}

	t := lineJson.Name
	i := lineJson.Sku
	u := lineJson.Url
	ls := fetchLineI18nLinks(doc)
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
			// According to collected logs, LINE ID befre exact 775 have special PNG encodings,
			// which are not parsable with libpng.
			// You will get -> CgBI: unhandled critical chunk <- from IM.
			// Workaround is to take the lower resolution "android" ones.
			if id, _ := strconv.Atoi(i); id < 775 && id != 0 {
				d = "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/android/" +
					"stickers.zip"
			} else {
				d += "stickers@2x.zip"
			}
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
	ld.i18nLinks = ls
	ld.category = c
	ld.dLink = d
	ld.id = i
	ld.title = t
	ld.isAnimated = a
	log.Debugln("line data parsed:", ld)
	ld.titleWg.Add(1)
	go fetchLineI18nTitles(ld)
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
	// ud.stickerData.title = ud.lineData.title + " @" + botName

	workDir := filepath.Join(ud.workDir, ud.lineData.id)
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
	// ud.stickerData.title = ud.lineData.title + " @" + botName

	if ud.lineData.category == LINE_STICKER_MESSAGE {
		return prepLineMessageS(ud)
	}

	workDir := filepath.Join(ud.workDir, ud.lineData.id)
	savePath := filepath.Join(workDir, "line.zip")
	os.MkdirAll(workDir, 0755)

	ud.wg.Add(1)
	err := fDownload(ud.lineData.dLink, savePath)
	if err != nil {
		return err
	}

	pngFiles := lineZipExtract(savePath, ud.lineData)
	if len(pngFiles) == 0 {
		return errors.New("no line image")
	}
	ud.wg.Done()

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
	sanitizeLinePNGs(files)
	return files
}

// Some LINE png contains a 'tEXt' textual information after `fcTL`
// FFMpeg's APNG demuxer could not parse it properly.
// I have patched ffmpeg, however, compiling it for AArch64 is not easy.
// Therefore, we are going to tackle the source itself by removing tEXt chunk.
func sanitizeLinePNGs(files []string) bool {
	for _, f := range files {
		ret := removeAPNGtEXtChunk(f)
		if ret == false {
			//do nothing.
		}
	}
	return true
}

func removeAPNGtEXtChunk(f string) bool {
	bytes, err := os.ReadFile(f)
	if err != nil {
		return false
	}
	l := len(bytes)
	textStart := 0
	textEnd := 0
	for i, _ := range bytes {
		if i > l-10 {
			break
		}
		tag := bytes[i : i+4]
		// only probe the first appearence of 'tEXt' tag.
		if string(tag) == "tEXt" && textStart == 0 {
			// 4 bytes before tag represents chunk length.
			textStart = i - 4
			// first IDAT after tEXt should be what we want.
		} else if string(tag) == "IDAT" && textStart != 0 {
			textEnd = i - 4
			break
		}
	}
	if textStart == 0 || textEnd == 0 {
		return true
	}
	newBytes := bytes[:textStart]
	newBytes = append(newBytes, bytes[textEnd:]...)

	os.Remove(f)
	fo, err := os.Create(f)
	if err != nil {
		return false
	}
	defer fo.Close()
	fo.Write(newBytes)
	log.Infoln("Sanitized one APNG file, path:", f)
	log.Infof("Length from %d to %d.", l, len(newBytes))
	return true
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
	workDir := filepath.Join(ud.workDir, ud.lineData.id)
	os.MkdirAll(workDir, 0755)

	page, err := httpGetCurlUA(ud.lineData.link)
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
		httpDownloadCurlUA(b, bPath)
		httpDownloadCurlUA(overlayImages[i], oPath)
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

func fetchLineI18nTitles(ld *LineData) {
	log.Debugln("Fetching LINE i18n titles...")
	log.Debugln(ld.i18nLinks)
	defer ld.titleWg.Done()

	var i18nTitles []string

	for _, l := range ld.i18nLinks {
		page, err := httpGet(l)
		if err != nil {
			continue
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
		var lineJson LineJson
		err = json.Unmarshal([]byte(doc.Find("script").First().Text()), &lineJson)
		if err != nil {
			continue
		}

		for _, t := range i18nTitles {
			// if title duplicates, skip loop
			if t == lineJson.Name {
				goto CONTINUE
			}
		}

		i18nTitles = append(i18nTitles, lineJson.Name)
	CONTINUE:
		continue
	}

	ld.i18nTitles = i18nTitles
	log.Debugln("I18N titles are:")
	log.Debugln(ld.i18nTitles)
}
