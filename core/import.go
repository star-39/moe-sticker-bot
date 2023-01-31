package core

import (
	"errors"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func parseImportLink(c tele.Context, link string, ld *LineData) error {
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
		return parseKakaoLink(c, link, ld)
	default:
		return errors.New("unknow import")
	}
}

func prepareImportStickers(ud *UserData, needConvert bool) error {
	switch ud.lineData.store {
	case "line":
		return prepareLineStickers(ud, needConvert)
	case "kakao":
		return prepareKakaoStickers(ud, needConvert)
	}
	return nil
}

func convertSToTGFormat(ud *UserData) {
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
