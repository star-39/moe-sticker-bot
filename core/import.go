package core

import (
	"errors"
	"net/url"
	"strings"
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
