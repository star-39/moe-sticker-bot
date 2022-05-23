package main

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	tele "github.com/star-39/telebot"
)

// main.go should only handle states and basic response,
// complex operations are done in other files.

func main() {
	initLogrus()

	log.Debug("Warn: Log level below DEBUG might print sensitive information, including passwords. Use with care.")
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Please set BOT_TOKEN environment variable!! Exiting...")
		return
	}
	pref := tele.Settings{
		Token:       token,
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
		// Genrally, issues are tackled inside each state, only fatal error should be returned to framework.
		// onError will terminate current session and log to terminal.
		OnError: onError,
	}
	log.WithField("token", token).Info("Attempting to initialize...")
	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	initWorkspace(b)
	log.WithFields(log.Fields{"botName": botName, "dataDir": dataDir}).Info("Bot OK.")

	b.Handle("/quit", cmdQuit)
	b.Handle("/cancel", cmdQuit)
	b.Handle("/exit", cmdQuit)
	b.Handle("/help", cmdStart, checkState)
	b.Handle("/about", cmdAbout, checkState)
	b.Handle("/faq", cmdFAQ, checkState)
	b.Handle("/import", cmdImport, checkState)
	b.Handle("/download", cmdDownload, checkState)
	b.Handle("/create", cmdCreate, checkState)
	b.Handle("/manage", cmdManage, checkState)

	b.Handle("/register", cmdRegister, checkState)
	b.Handle("/sanitize", cmdSanitize, checkState)

	b.Handle("/start", cmdStart, checkState)
	// Handle contents.
	b.Handle(tele.OnText, handleMessage)
	b.Handle(tele.OnVideo, handleMessage)
	b.Handle(tele.OnAnimation, handleMessage)
	b.Handle(tele.OnSticker, handleMessage)
	b.Handle(tele.OnDocument, handleMessage)
	b.Handle(tele.OnPhoto, handleMessage)
	b.Handle(tele.OnCallback, handleMessage, autoRespond, sanitizeCallback)

	b.Start()
}

func handleMessage(c tele.Context) error {
	var err error
	command, state := getState(c)
	if command == "" {
		return handleNoSession(c)
	}
	switch command {
	case "import":
		switch state {
		case "waitImportLink":
			err = waitImportLink(c)
		case "waitCbImportChoice":
			err = waitCbImportChoice(c)
		case "waitSTitle":
			err = waitSTitle(c)
		case "waitEmojiChoice":
			err = waitEmojiChoice(c)
		case "process":
			err = stateProcessing(c)
		case "waitSEmojiAssign":
			err = waitSEmojiAssign(c)
		}
	case "download":
		switch state {
		case "waitSDownload":
			err = waitSDownload(c)
		case "process":
			err = stateProcessing(c)
		}
	case "create":
		switch state {
		case "waitSType":
			err = waitSType(c)
		case "waitSTitle":
			err = waitSTitle(c)
		case "waitSID":
			err = waitSID(c)
		case "waitSFile":
			err = waitSFile(c)
		case "waitEmojiChoice":
			err = waitEmojiChoice(c)
		case "waitSEmojiAssign":
			err = waitSEmojiAssign(c)
		case "process":
			err = stateProcessing(c)
		}
	case "manage":
		switch state {
		case "waitSManage":
			err = waitSManage(c)
		case "waitCbEditChoice":
			err = waitCbEditChoice(c)
		case "waitSMovFrom":
			err = waitSMovFrom(c)
		case "waitSMovTo":
			err = waitSMovTo(c)
		case "waitSEmojiEdit":
			err = waitSEmojiEdit(c)
		case "waitEmojiEdit":
			err = waitEmojiEdit(c)
		case "waitSFile":
			err = waitSFile(c)
		case "waitEmojiChoice":
			err = waitEmojiChoice(c)
		case "waitSEmojiAssign":
			err = waitSEmojiAssign(c)
		case "waitSDel":
			err = waitSDel(c)
		case "waitCbDelset":
			err = waitCbDelset(c)
		case "process":
			err = stateProcessing(c)
		}

	case "register":
		switch state {
		case "waitRegLineLink":
			err = waitRegLineLink(c)
		case "waitRegS":
			err = waitRegS(c)
		}
	}
	return err
}

func initWorkspace(b *tele.Bot) {
	botName = b.Me.Username
	dataDir = botName + "_data"
	users = Users{data: make(map[int64]*UserData)}
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		log.Fatal(err)
		return
	}
	go initUserDirGCTimer(dataDir)

	if os.Getenv("USE_DB") == "1" {
		dbName := getEnv("DB_NAME", botName+"_db")
		initDB(dbName)
	} else {
		log.Warn("Not using database because USE_DB is not set to 1.")
	}

	ADMIN_UID, _ = strconv.ParseInt(os.Getenv("ADMIN_UID"), 10, 64)
	wpConvertWebm, _ = ants.NewPoolWithFunc(4, wConvertWebm)

	if runtime.GOOS == "linux" {
		BSDTAR_BIN = "bsdtar"
		CONVERT_BIN = "convert"
	} else {
		BSDTAR_BIN = "tar"
		CONVERT_BIN = "magick"
		CONVERT_ARGS = []string{"convert"}
	}
}

func initUserDirGCTimer(dataDir string) {
	ticker := time.NewTicker(48 * time.Hour)
	for t := range ticker.C {
		log.Infoln("UserDir GC Timter ticked at:", t)
		log.Info("Purging outdated user dir...")
		purgeOutdatedUserData(dataDir)
	}
}

func initLogrus() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		DisableLevelTruncation: true,
	})

	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARNING":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.TraceLevel)
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

func onError(err error, c tele.Context) {
	sendFatalError(err, c)
	cleanUserDataAndDir(c.Sender().ID)
}
