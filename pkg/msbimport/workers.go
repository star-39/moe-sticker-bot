package msbimport

import (
	"strings"

	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
)

// Workers pool for converting webm
var wpConvertWebm, _ = ants.NewPoolWithFunc(4, wConvertWebm)

// Accepts *LineFile
func wConvertWebm(i interface{}) {
	lf := i.(*LineFile)
	defer lf.Wg.Done()
	log.Debugln("Converting in pool for:", lf)

	var err error
	//FFMpeg doest not support animated webp.
	//IM convert it to apng then feed to webm.
	if strings.HasSuffix(lf.OriginalFile, ".webp") {
		lf.OriginalFile, _ = IMToApng(lf.OriginalFile)
	}
	if lf.IsEmoji {
		lf.ConvertedFile, err = FFToWebmTGVideo(lf.OriginalFile, FORMAT_TG_EMOJI_ANIMATED)
	} else {
		lf.ConvertedFile, err = FFToWebmTGVideo(lf.OriginalFile, FORMAT_TG_REGULAR_ANIMATED)
	}

	if err != nil {
		lf.CError = err
	}
	log.Debugln("convert OK: ", lf.ConvertedFile)
}
