package msbimport

import (
	"context"
	"errors"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Parse a LINE or Kakao link and fetch metadata.
// The metadata can be used to call prepareImportStickers.
// Returns a string and an error. String act as a warning message, empty string means no warning yield.
//
// Attention: During this step, ld.Amount, ld.Files and ld.IsAnimated will NOT be available!
func ParseImportLink(link string, ld *LineData) (string, error) {
	var warn string

	u, err := url.Parse(link)
	if err != nil {
		return warn, err
	}
	switch {
	case strings.HasSuffix(u.Host, "line.me"):
		ld.Store = "line"
		return parseLineLink(link, ld)
	case strings.HasSuffix(u.Host, "kakao.com"):
		ld.Store = "kakao"
		return parseKakaoLink(link, ld)
	default:
		return warn, errors.New("unknow import")
	}
}

// Prepare import stickers files.
// A context is provided, which can be used to interrupt the process.
// When this function returns, stickers are ready to be sent to execAutoCommit.
// However, wg inside each LineFile might still not being done yet,
// wg.Wait() is required for individual sticker file.
//
// ld.Amount, ld.Files and ld.IsAnimated will be produced after return.
func PrepareImportStickers(ctx context.Context, ld *LineData, workDir string, needConvert bool) error {
	switch ld.Store {
	case "line":
		return prepareLineStickers(ctx, ld, workDir, needConvert)
	case "kakao":
		return prepareKakaoStickers(ctx, ld, workDir, needConvert)
	}
	return nil
}

// Convert imported sticker to Telegram format,
// which means WEBM for animated and WEBP for static
// with 512x512 dimension.
func convertSToTGFormat(ctx context.Context, ld *LineData) {
	for _, s := range ld.Files {
		select {
		case <-ctx.Done():
			log.Warn("convertSToTGFormat received ctxDone!")
			return
		default:
		}
		var err error
		// If lineS is animated, commit to worker pool
		// since encoding vp9 is time and resource costy.
		if ld.IsAnimated {
			wpConvertWebm.Invoke(s)
		} else {
			s.ConvertedFile, err = IMToWebpTGStatic(s.OriginalFile, ld.IsEmoji)
			if err != nil {
				s.CError = err
			}
			s.Wg.Done()
		}
	}
}
