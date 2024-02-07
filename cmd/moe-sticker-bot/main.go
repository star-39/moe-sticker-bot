package main

import (
	"flag"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/core"
)

// Common abbr. in this project:
// S : Sticker
// SS : StickerSet

func main() {
	conf := parseCmdLine()
	core.Init(conf)
}

func parseCmdLine() core.ConfigTemplate {
	var help = flag.Bool("help", false, "Show help")
	var adminUid = flag.Int64("admin_uid", -1, "Admin's UID(optional)")
	var botToken = flag.String("bot_token", "", "Telegram Bot Token")
	var dataDir = flag.String("data_dir", "", "Overwrites the working directory where msb puts data.")
	var webappUrl = flag.String("webapp_url", "", "Public HTTPS URL to WebApp, in unset, webapp will be disabled.")
	var WebappApiListenAddr = flag.String("webapp_listen_addr", "", "Webapp API server listen address(IP:PORT)")
	var webappDataDir = flag.String("webapp_data_dir", "", "Where to put webapp data to share with ReactApp ")
	var dbAddr = flag.String("db_addr", "", "mariadb(mysql) address, if unset, database will be disabled.")
	var dbUser = flag.String("db_user", "", "mariadb(mysql) usernmae")
	var dbPass = flag.String("db_pass", "", "mariadb(mysql) password")
	var logLevel = flag.String("log_level", "debug", "Log level")
	// var botApiAddr = flag.String("botapi_addr", "", "Local Bot API Server Address")
	// var botApiDir = flag.String("botapi_dir", "", "Local Bot API Working directory")
	// var webhookPublicAddr = flag.String("webhook_public_addr", "", "Webhook public address(WebhookEndpoint).")
	// var webhookListenAddr = flag.String("webhook_listen_addr", "", "Webhook listen address(IP:PORT)")
	// var webhookCert = flag.String("webhook_cert", "", "Certificate for WebHook")
	flag.Parse()
	if *help {
		flag.Usage()
		println("Only --bot_token is required to run.")
		os.Exit(0)
	}

	conf := core.ConfigTemplate{}

	conf.BotToken = *botToken
	if conf.BotToken == "" {
		log.Error("Please set --bot_token")
		log.Error("Use --help to see options.")
		log.Fatal("No bot token provided!")
	}
	if !strings.Contains(conf.BotToken, ":") {
		log.Fatal("Bad bot token!")
	}

	conf.DbAddr = *dbAddr
	conf.DbUser = *dbUser
	conf.DbPass = *dbPass

	conf.WebappUrl = *webappUrl
	conf.WebappDataDir = *webappDataDir
	conf.WebappApiListenAddr = *WebappApiListenAddr

	// conf.BotApiAddr = *botApiAddr
	// conf.BotApiDir = *botApiDir
	// conf.WebhookPublicAddr = *webhookPublicAddr
	// conf.WebhookListenAddr = *webhookListenAddr
	// conf.WebhookCert = *webhookCert

	conf.LogLevel = *logLevel

	conf.AdminUid = *adminUid
	conf.DataDir = *dataDir

	return conf
	// core.Config = conf
}
