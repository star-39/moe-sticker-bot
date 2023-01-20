package core

import (
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
)

// Workers pool for converting webm
var wpConvertWebm *ants.PoolWithFunc
var wpDownloadStickerSet *ants.PoolWithFunc

// var wpDownloadSticker *ants.PoolWithFunc
// var wpDownloadTGSSticker *ants.PoolWithFunc

func initWorkersPool() {
	wpConvertWebm, _ = ants.NewPoolWithFunc(4, wConvertWebm)
	wpDownloadStickerSet, _ = ants.NewPoolWithFunc(
		8, wDownloadStickerObject)
}

// *StickerFile
func wConvertWebm(i interface{}) {
	sf := i.(*StickerFile)
	defer sf.wg.Done()
	log.Debugln("Converting in pool for:", sf)

	var err error
	sf.cPath, err = ffToWebm(sf.oPath)
	if err != nil {
		sf.cError = err
	}
	log.Debugln("convert OK: ", sf.cPath)
}

// *StickerDownloadObject
func wDownloadStickerObject(i interface{}) {
	obj := i.(*StickerDownloadObject)
	defer obj.wg.Done()
	log.Debugf("Downloading in pool: %s -> %s", obj.sticker.FileID, obj.dest)

	//WebApp does not need special conversion.
	if obj.forWebApp {
		err := obj.bot.Download(&obj.sticker.File, obj.dest)
		if err != nil {
			log.Warnln("download: error downloading sticker:", err)
			obj.err = err
			return
		}
		if obj.sticker.Video {
			if obj.webAppHQ {
				obj.err = ffToAnimatedWebpWA(obj.dest)
			} else {
				obj.err = imToAnimatedWebpLQ(obj.dest)
			}
		} else {
			imToWebpWA(obj.dest)
		}
		if obj.webAppThumb {
			obj.err = imToPNGThumb(obj.dest)
		}
		return
	}

	var f string
	var cf string
	var err error
	if obj.sticker.Video {
		f = obj.dest + ".webm"
		err = obj.bot.Download(&obj.sticker.File, f)
		if obj.needConvert {
			if obj.shrinkGif {
				cf, _ = ffToGifShrink(f)
			} else {
				cf, _ = ffToGif(f)
			}
		}
	} else if obj.sticker.Animated {
		f = obj.dest + ".tgs"
		err = obj.bot.Download(&obj.sticker.File, f)
		if obj.needConvert {
			cf, _ = lottieToGIF(f)
		}
	} else {
		f = obj.dest + ".webp"
		err = obj.bot.Download(&obj.sticker.File, f)
		if obj.needConvert {
			cf, _ = imToPng(f)
		}
	}
	if err != nil {
		log.Warnln("download: error downloading sticker:", err)
		obj.err = err
		return
	}

	obj.of = f
	obj.cf = cf

}

// *StickerMoveObject
func wSubmitSMove(i interface{}) {
	obj := i.(*StickerMoveObject)
	defer obj.wg.Done()
	sid := obj.sd.stickerSet.Stickers[obj.oldIndex].FileID
	log.Debugf("Moving in pool %d(%s) -> %d", obj.oldIndex, sid, obj.newIndex)
	err := b.SetStickerPosition(sid, obj.newIndex)
	if err != nil {
		log.Debugln("SMove failed!!", err)
		obj.err = err
	} else {
		log.Debugf("Sticker move OK for %s", obj.sd.stickerSet.Name)
		obj.sd.stickerSet.Stickers =
			sliceMove(obj.oldIndex, obj.newIndex, obj.sd.stickerSet.Stickers)
	}
}
