package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
)

func Init(conf ConfigTemplate) {
	msbconf = conf
	initLogrus()
	msbimport.InitConvert()
	b = initBot(conf)
	initWorkspace(b)
	initWorkersPool()
	go initGoCron()
	if msbconf.WebApp {
		InitWebAppServer()
	} else {
		log.Info("WebApp not enabled.")
	}

	log.WithFields(log.Fields{"botName": botName, "dataDir": dataDir}).Info("Bot OK.")

	// complies to telebot v3.1
	b.Use(Recover())

	b.Handle("/quit", cmdQuit)
	b.Handle("/cancel", cmdQuit)
	b.Handle("/exit", cmdQuit)
	b.Handle("/faq", cmdFAQ)
	b.Handle("/changelog", cmdChangelog)
	b.Handle("/privacy", cmdPrivacy)
	b.Handle("/help", cmdStart)
	b.Handle("/about", cmdAbout)
	b.Handle("/command_list", cmdCommandList)
	b.Handle("/import", cmdImport, checkState)
	b.Handle("/download", cmdDownload, checkState)
	b.Handle("/create", cmdCreate, checkState)
	b.Handle("/manage", cmdManage, checkState)
	b.Handle("/search", cmdSearch, checkState)

	// b.Handle("/register", cmdRegister, checkState)
	b.Handle("/sitrep", cmdSitRep, checkState)
	b.Handle("/getfid", cmdGetFID, checkState)

	b.Handle("/start", cmdStart, checkState)

	b.Handle(tele.OnText, handleMessage)
	b.Handle(tele.OnVideo, handleMessage)
	b.Handle(tele.OnAnimation, handleMessage)
	b.Handle(tele.OnSticker, handleMessage)
	b.Handle(tele.OnDocument, handleMessage)
	b.Handle(tele.OnPhoto, handleMessage)
	b.Handle(tele.OnCallback, handleMessage, autoRespond, sanitizeCallback)

	b.Start()
}

// Recover returns a middleware that recovers a panic happened in
// the handler.
func Recover(onError ...func(error)) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			var f func(error)
			if len(onError) > 0 {
				f = onError[0]
			} else {
				f = func(err error) {
					c.Bot().OnError(err, c)
				}
			}

			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						f(err)
					} else if s, ok := r.(string); ok {
						f(errors.New(s))
					}
				}
			}()

			return next(c)
		}
	}
}

// This one never say goodbye.
func endSession(c tele.Context) {
	cleanUserDataAndDir(c.Sender().ID)
}

// This one will say goodbye.
func terminateSession(c tele.Context) {
	cleanUserDataAndDir(c.Sender().ID)
	c.Send("Bye. /start")
}

func endManageSession(c tele.Context) {
	ud, exist := users.data[c.Sender().ID]
	if !exist {
		return
	}
	if ud.stickerData.id == "" {
		return
	}
	path := filepath.Join(msbconf.WebappDataDir, ud.stickerData.id)
	os.RemoveAll(path)
}

func onError(err error, c tele.Context) {
	log.Error("User encountered fatal error!")
	log.Errorln("Raw error:", err)
	log.Errorln(string(debug.Stack()))

	defer func() {
		if r := recover(); r != nil {
			log.Errorln("Recovered from onError!!", r)
		}
	}()
	if c == nil {
		return
	}
	sendFatalError(err, c)
	cleanUserDataAndDir(c.Sender().ID)
}

func initBot(conf ConfigTemplate) *tele.Bot {
	var poller tele.Poller
	url := tele.DefaultApiURL

	if msbconf.WebhookListenAddr != "" {
		poller = &tele.Webhook{
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: msbconf.WebhookPublicAddr,
			},
			Listen: msbconf.WebhookListenAddr,
		}
		if msbconf.BotApiAddr != "" {
			url = msbconf.BotApiAddr
		}
	} else {
		poller = &tele.LongPoller{Timeout: 10 * time.Second}
	}

	pref := tele.Settings{
		URL:         url,
		Token:       msbconf.BotToken,
		Poller:      poller,
		Synchronous: false,
		// Genrally, issues are tackled inside each state, only fatal error should be returned to framework.
		// onError will terminate current session and log to terminal.
		OnError: onError,
	}
	log.WithField("token", msbconf.BotToken).Info("Attempting to initialize...")
	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func initWorkspace(b *tele.Bot) {
	botName = b.Me.Username
	if msbconf.DataDir != "" {
		dataDir = msbconf.DataDir
	} else {
		dataDir = botName + "_data"
	}
	users = Users{data: make(map[int64]*UserData)}
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if msbconf.UseDB {
		dbName := botName + "_db"
		err = initDB(dbName)
		if err != nil {
			log.Fatalln("Error initializing database!!", err)
		}
	} else {
		log.Warn("Not using database because --use_db is not set.")
	}
}

func initGoCron() {
	time.Sleep(15 * time.Second)
	cronScheduler = gocron.NewScheduler(time.UTC)
	cronScheduler.Every(2).Days().Do(purgeOutdatedStorageData)
	if msbconf.UseDB {
		cronScheduler.Every(1).Weeks().Do(curateDatabase)
	}
	cronScheduler.StartAsync()
}

func initLogrus() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
	})

	level, err := log.ParseLevel(msbconf.LogLevel)
	if err != nil {
		println("Error parsing log_level! Defaulting to DEBUG level.\n")
		log.SetLevel(log.DebugLevel)
	}
	log.SetLevel(level)

	fmt.Printf("Log level is set to: %s\n", log.GetLevel())
	log.Debug("Warning: Log level below DEBUG might print sensitive information, including passwords.")
}
