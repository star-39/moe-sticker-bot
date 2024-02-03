package core

import (
	"strings"

	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
)

// Workers pool for converting webm
var wpDownloadStickerSet *ants.PoolWithFunc

func initWorkersPool() {
	// wpConvertWebm, _ = ants.NewPoolWithFunc(4, wConvertWebm)
	wpDownloadStickerSet, _ = ants.NewPoolWithFunc(
		8, wDownloadStickerObject)
}

// *StickerDownloadObject
func wDownloadStickerObject(i interface{}) {
	obj := i.(*StickerDownloadObject)
	defer obj.wg.Done()
	log.Debugf("Downloading in pool: %s -> %s", obj.sticker.FileID, obj.dest)

	if obj.forWebApp || obj.forWhatsApp {
		err := teleDownload(&obj.sticker.File, obj.dest)
		if err != nil {
			log.Warnln("download: error downloading sticker:", err)
			obj.err = err
			return
		}
		if obj.forWhatsApp {
			if obj.sticker.Video {
				obj.err = msbimport.FFToAnimatedWebpWA(obj.dest)
			} else if obj.sticker.Animated {
				_, obj.err = msbimport.RlottieToWebp(obj.dest)

			} else {
				obj.err = msbimport.IMToWebpWA(obj.dest)
			}

			if obj.forWhatsAppThumb {
				if obj.sticker.Animated {
					f := strings.ReplaceAll(obj.dest, ".tgs", ".webp")
					obj.err = msbimport.IMToPNGThumb(f)
				} else {
					obj.err = msbimport.IMToPNGThumb(obj.dest)
				}
			}
		} else {
			//TGS set is not managable, no need to convert.
			if obj.sticker.Video {
				obj.err = msbimport.FFToAnimatedWebpLQ(obj.dest)
			}
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
			cf, _ = msbimport.FFToGif(f)
		}
	} else if obj.sticker.Animated {
		f = obj.dest + ".tgs"
		err = teleDownload(&obj.sticker.File, f)
		if obj.needConvert {
			cf, _ = msbimport.RlottieToGIF(f)
		}
	} else {
		f = obj.dest + ".webp"
		err = teleDownload(&obj.sticker.File, f)
		if obj.needConvert {
			cf, _ = msbimport.IMToPng(f)
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
		log.Errorln("SMove failed!!", err)
		obj.err = err
	} else {
		log.Debugf("Sticker move OK for %s", obj.sd.stickerSet.Name)
		obj.sd.stickerSet.Stickers =
			sliceMove(obj.oldIndex, obj.newIndex, obj.sd.stickerSet.Stickers)
	}
}
