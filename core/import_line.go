package core

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

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
	err = parseLineDetails(doc, &lineJson)
	if err != nil {
		log.Errorln("parseLineLink: ", err)
		return err
	}

	t := lineJson.Name
	i := lineJson.Sku
	u := lineJson.Url
	ls := fetchLineI18nLinks(doc)
	a := false
	c := ""
	d := "https://stickershop.line-scdn.net/stickershop/v1/product/" + i + "/iphone/"

	if strings.Contains(u, "stickershop") || strings.Contains(u, "officialaccount/event/sticker") {
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
		lineJson := &LineJson{}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page))
		if err != nil {
			continue
		}
		parseLineDetails(doc, lineJson)

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

// This function goes after parseLineLink
// Receives a gq document of LINE Store page and parse the details to *LineJson.
func parseLineDetails(doc *goquery.Document, lj *LineJson) error {
	// For typical line store page, the first script is sticker's metadata in JSON.
	// Parse to json, if OK, return nil here.
	err := json.Unmarshal([]byte(doc.Find("script").First().Text()), lj)
	if err == nil {
		return nil
	}

	// Some new line store page does not have a json metadata <script>
	log.Warnln("Failed json parsing line link!", err)
	log.Warnln("Special LINE type? Trying to guess info.")

	// Find URL.
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("property")
		if value == "og:url" {
			lj.Url, _ = s.Attr("content")
		}
	})

	// Find title.
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		oa, _ := s.Attr("data-test")
		if oa == "oa-sticker-title" {
			lj.Name = s.Text()
		}
	})
	// Find title again.
	if lj.Name == "" {
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			oa, _ := s.Attr("data-test")
			if oa == "sticker-name-title" {
				lj.Name = s.Text()
			}
		})
	}
	if lj.Name == "" {
		return errors.New("guess line: no title")
	}

	// Find ID.
	var defaultLink string
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		hreflang, _ := s.Attr("hreflang")
		if hreflang == "x-default" {
			defaultLink, _ = s.Attr("href")
		}
	})
	lj.Sku = path.Base(defaultLink)
	if lj.Sku == "" {
		return errors.New("guess line: no id(no link base)")
	}

	log.Debugln("parsed line detail:", lj)
	return nil
}

// Download and convert sticker files after parseLineLink.
func prepareLineStickers(ud *UserData, needConvert bool) error {
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	ud.stickerData.isVideo = ud.lineData.isAnimated
	ud.stickerData.id = "line_" + ud.lineData.id + secNum(4) + "_by_" + botName

	if ud.lineData.category == LINE_STICKER_MESSAGE {
		return prepareLineMessageS(ud)
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
		convertSToTGFormat(ud)
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
		if !ret {
			log.Debugln("one file sanitization ignored:", f)
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
	if l < 42 {
		return false
	}
	// byte index 37-40 must be acTL Animation Control Chunk.
	if string(bytes[37:41]) != "acTL" {
		return false
	}
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
		return false
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

func prepareLineMessageS(ud *UserData) error {
	workDir := filepath.Join(ud.workDir, ud.lineData.id)
	os.MkdirAll(workDir, 0755)

	//Only curl UA will work.
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
