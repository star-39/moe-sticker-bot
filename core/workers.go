package core

import (
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/convert"
)

// Workers pool for converting webm
var wpConvertWebm *ants.PoolWithFunc
var wpDownloadStickerSet *ants.PoolWithFunc

// var wpDownloadSticker *ants.PoolWithFunc
// var wpDownloadTGSSticker *ants.PoolWithFunc

func initWorkersPool() {
	// wpConvertWebm, _ = ants.NewPoolWithFunc(4, wConvertWebm)
	wpDownloadStickerSet, _ = ants.NewPoolWithFunc(
		8, wDownloadStickerObject)
}

// // *StickerFile
// func wConvertWebm(i interface{}) {
// 	sf := i.(*StickerFile)
// 	defer sf.wg.Done()
// 	log.Debugln("Converting in pool for:", sf)

// 	var err error
// 	//FFMpeg doest not support animated webp.
// 	//IM convert it to apng then feed to webm.
// 	if strings.HasSuffix(sf.oPath, ".webp") {
// 		sf.oPath, _ = convert.IMToApng(sf.oPath)
// 	}
// 	sf.cPath, err = convert.FFToWebm(sf.oPath)

// 	if err != nil {
// 		sf.cError = err
// 	}
// 	log.Debugln("convert OK: ", sf.cPath)
// }

// *StickerDownloadObject
func wDownloadStickerObject(i interface{}) {
	obj := i.(*StickerDownloadObject)
	defer obj.wg.Done()
	log.Debugf("Downloading in pool: %s -> %s", obj.sticker.FileID, obj.dest)

	//WebApp does not need special conversion.
	if obj.forWebApp {
		err := teleDownload(&obj.sticker.File, obj.dest)
		if err != nil {
			log.Warnln("download: error downloading sticker:", err)
			obj.err = err
			return
		}
		if obj.sticker.Video {
			if obj.webAppHQ {
				obj.err = convert.FFToAnimatedWebpWA(obj.dest)
			} else {
				obj.err = convert.IMToAnimatedWebpLQ(obj.dest)
			}
		} else {
			convert.IMToWebpWA(obj.dest)
		}
		if obj.webAppThumb {
			obj.err = convert.IMToPNGThumb(obj.dest)
		}
		return
	}

	var f string
	var cf string
	var err error
	if obj.sticker.Video {
		f = obj.dest + ".webm"
		err = teleDownload(&obj.sticker.File, f)
		if obj.needConvert {
			if obj.shrinkGif {
				cf, _ = convert.FFToGifShrink(f)
			} else {
				cf, _ = convert.FFToGif(f)
			}
		}
	} else if obj.sticker.Animated {
		f = obj.dest + ".tgs"
		err = teleDownload(&obj.sticker.File, f)
		if obj.needConvert {
			cf, _ = convert.LottieToGIF(f)
		}
	} else {
		f = obj.dest + ".webp"
		err = teleDownload(&obj.sticker.File, f)
		if obj.needConvert {
			cf, _ = convert.IMToPng(f)
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
