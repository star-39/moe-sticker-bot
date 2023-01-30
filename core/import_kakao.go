package core

import (
	"encoding/json"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

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

func prepKakaoStickers(ud *UserData, needConvert bool) error {
	ud.udWg.Add(1)
	defer ud.udWg.Done()
	ud.stickerData.id = "kakao_" + ud.lineData.id + secHex(2) + "_by_" + botName

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
		f := filepath.Join(workDir, path.Base(l)+".png")
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
