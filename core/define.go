package core

import (
	"context"
	"sync"

	"github.com/go-co-op/gocron"
	"github.com/panjf2000/ants/v2"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
)

var BOT_VERSION = "2.3.3-GO"

// The telegram bot.
var b *tele.Bot
var cronScheduler *gocron.Scheduler

var dataDir string
var botName string

var downloadQueue DownloadQueue
var webAppSSAuthList WebAppQIDAuthList

// ['uid'] -> bool channels
var autocommitWorkersList = make(map[int64][]chan bool)
var users Users

const (
	CB_DN_WHOLE      = "dall"
	CB_DN_SINGLE     = "dsingle"
	CB_OK_IMPORT     = "yesimport"
	CB_OK_DN         = "yesd"
	CB_BYE           = "bye"
	CB_MANAGE        = "manage"
	CB_DONE_ADDING   = "done"
	CB_YES           = "yes"
	CB_NO            = "no"
	CB_DEFAULT_TITLE = "titledefault"
	CB_EXPORT_WA     = "exportwa"

	ST_WAIT_WEBAPP = "waitWebApp"
	ST_PROCESSING  = "process"

	FID_KAKAO_SHARE_LINK = "AgACAgEAAxkBAAEjezVj3_YXwaQ8DM-107IzlLSaXyG6yAACfKsxG3z7wEadGGF_gJrcnAEAAwIAA3kAAy4E"

	LINK_TG     = "t.me"
	LINK_LINE   = "line.me"
	LINK_KAKAO  = "kakao.com"
	LINK_IMPORT = "IMPORT"
)

type LineStickerQ struct {
	Line_id   string
	Line_link string
	Tg_id     string
	Tg_title  string
	Ae        bool
}

type UserStickerQ struct {
	tg_id     string
	tg_title  string
	timestamp int64
}

type WebAppUser struct {
	Id            int
	Is_bot        bool
	First_name    string
	Last_name     string
	Username      string
	Language_code string
	Is_premium    bool
	Photo_url     string
}

type UserData struct {
	//waitgroup for sticker set, wait before commit.
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	state            string
	sessionID        string
	workDir          string
	command          string
	progress         string
	progressMsg      *tele.Message
	lineData         *msbimport.LineData
	stickerData      *StickerData
	webAppUser       *WebAppUser
	webAppQID        string
	webAppWorkerPool *ants.PoolWithFunc
	lastContext      tele.Context
}

type DownloadQueue struct {
	mu sync.Mutex
	ss map[string]bool
}

type WebAppQIDAuthList struct {
	mu sync.Mutex
	sa map[string]*WebAppQIDAuthObject
}

type WebAppQIDAuthObject struct {
	sn string
	dt int64
}

type Users struct {
	mu   sync.Mutex
	data map[int64]*UserData
}

type StickerMoveObject struct {
	wg       sync.WaitGroup
	err      error
	sd       *StickerData
	oldIndex int
	newIndex int
}

func (ud *UserData) udSetState(state string) {
	ud.state = state
}

type StickerFile struct {
	wg sync.WaitGroup
	// path of original file
	oPath string
	// path of converted filea
	cPath  string
	cError error
}

type StickerData struct {
	id string
	// link     string
	title      string
	emojis     []string
	sticker    *tele.Sticker
	stickers   []*StickerFile
	stickerSet *tele.StickerSet
	sDnObjects []*StickerDownloadObject
	isVideo    bool
	pos        int
	// amount of local files
	lAmount int
	// amount on cloud
	cAmount int
}

type StickerDownloadObject struct {
	wg      sync.WaitGroup
	bot     *tele.Bot
	sticker tele.Sticker
	dest    string
	//Convert to conventional format?
	needConvert bool
	//Shrink oversized GIF?
	shrinkGif bool
	//Sticker is for WebApp?
	forWebApp bool
	//need HQ animated sticker for WhatsApp
	webAppHQ bool
	//need 96px PNG thumb for WhatsApp
	webAppThumb bool
	/*
		Following fields are yielded by worker after wg is done.
	*/
	//Original sticker file downloaded.
	of string
	//Converted sticker file.
	cf string
	//Returned error.
	err error
}
