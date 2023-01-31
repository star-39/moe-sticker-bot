package core

import (
	"errors"
	"net/url"
	"strings"

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
