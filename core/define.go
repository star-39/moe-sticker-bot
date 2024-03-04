package core

import (
	"context"
	"sync"

	"github.com/go-co-op/gocron"
	"github.com/panjf2000/ants/v2"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
)

var BOT_VERSION = "2.4.0-RC4"

var b *tele.Bot
var cronScheduler *gocron.Scheduler

var dataDir string
var botName string

// ['uid'] -> bool channels
var autocommitWorkersList = make(map[int64][]chan bool)
var users Users

const (
	CB_DN_WHOLE           = "dall"
	CB_DN_SINGLE          = "dsingle"
	CB_OK_IMPORT          = "yesimport"
	CB_OK_IMPORT_EMOJI    = "yesimportemoji"
	CB_OK_DN              = "yesd"
	CB_BYE                = "bye"
	CB_MANAGE             = "manage"
	CB_DONE_ADDING        = "done"
	CB_YES                = "yes"
	CB_NO                 = "no"
	CB_DEFAULT_TITLE      = "titledefault"
	CB_EXPORT_WA          = "exportwa"
	CB_ADD_STICKER        = "adds"
	CB_DELETE_STICKER     = "dels"
	CB_DELETE_STICKER_SET = "delss"
	CB_CHANGE_TITLE       = "changetitle"

	ST_WAIT_WEBAPP = "waitWebApp"
	ST_PROCESSING  = "process"

	FID_KAKAO_SHARE_LINK = "AgACAgEAAxkBAAEjezVj3_YXwaQ8DM-107IzlLSaXyG6yAACfKsxG3z7wEadGGF_gJrcnAEAAwIAA3kAAy4E"

	LINK_TG     = "t.me"
	LINK_LINE   = "line.me"
	LINK_KAKAO  = "kakao.com"
	LINK_IMPORT = "IMPORT"
)

// Object for quering database for Line Sticker.
type LineStickerQ struct {
	Line_id   string
	Line_link string
	Tg_id     string
	Tg_title  string
	Ae        bool
}

// Object for quering database for User Sticker.
type UserStickerQ struct {
	tg_id     string
	tg_title  string
	timestamp int64
}

// Telegram API JSON.
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

// Unique user data for one user and one session.
type UserData struct {
	//waitgroup for sticker set, wait before commit.
	wg sync.WaitGroup
	//commit channel for emoji assign
	commitChans []chan bool
	ctx         context.Context
	cancel      context.CancelFunc

	//Current conversational state.
	state     string
	sessionID string
	workDir   string
	//Current command.
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

// Map for users, identified by user id.
// All temporal user data are stored in this struct.
type Users struct {
	mu   sync.Mutex
	data map[int64]*UserData
}

// Object for ants worker function.
// wg must be initiated with wg.Add(1)
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

// Object for ants worker function.
// wg must be initialized with wg.Add(1)
type StickerFile struct {
	wg sync.WaitGroup
	// path of original file
	oPath string
	// path of converted filea
	cPath  string
	cError error
}

// General sticker data for internal use.
type StickerData struct {
	id string
	// link     string
	title          string
	emojis         []string
	stickers       []*StickerFile
	stickerSet     *tele.StickerSet
	sDnObjects     []*StickerDownloadObject
	stickerSetType tele.StickerSetType
	//either static or video, used for CreateNewStickerSet
	// getFormat     StickerData
	isVideo       bool
	isAnimated    bool
	isCustomEmoji bool
	pos           int
	// amount of local files
	lAmount int
	// amount on cloud
	cAmount int
	// amout of flood error encounterd
	flCount int
}

func (sd StickerData) getFormat() string {
	if sd.isVideo {
		return "video"
	} else {
		return "static"
	}
}

type StickerDownloadObject struct {
	wg      sync.WaitGroup
	sticker tele.Sticker
	dest    string
	//Convert to conventional format?
	needConvert bool
	//Shrink oversized GIF?
	shrinkGif bool
	//need to convert to WebApp use case
	forWebApp bool
	//need to convert to WhatsApp format
	forWhatsApp bool
	//need 96px PNG thumb for WhatsApp
	forWhatsAppThumb bool
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
