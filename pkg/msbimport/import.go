package msbimport

import (
	"context"
	"errors"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// This function serves as an entrypoint for this package.
// Parse a LINE or Kakao link and fetch metadata.
// The metadata (which means the LineData struct) can be used to call prepareImportStickers.
// Returns a string and an error. String act as a warning message, empty string means no warning yield.
//
// Attention: After this function returns, ld.Amount, ld.Files will NOT be available!
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

// Prepare stickers files.
// Should be called after calling ParseImportLink().
// A context is provided, which can be used to interrupt the process.
// Even if this function returns, file preparation might still in progress.
// LineData.Amount, LineData.Files will be produced after return.
// wg.Wait() is required for individual LineData.Files
//
// convertToTGFormat: Convert original stickers to Telegram sticker format.
// convertToTGEmoji: If present sticker set is Emoji(LINE), convert to 100x100 Telegram CustomEmoji.
func PrepareImportStickers(ctx context.Context, ld *LineData, workDir string, convertToTGFormat bool, convertToTGEmoji bool) error {
	switch ld.Store {
	case "line":
		return prepareLineStickers(ctx, ld, workDir, convertToTGFormat, convertToTGEmoji)
	case "kakao":
		return prepareKakaoStickers(ctx, ld, workDir, convertToTGFormat)
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
			s.ConvertedFile, err = IMToWebpTGStatic(s.OriginalFile, s.ConvertToEmoji)
			if err != nil {
				s.CError = err
			}
			s.Wg.Done()
		}
	}
}
