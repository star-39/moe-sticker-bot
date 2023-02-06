package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/core"
)

// Common abbr. in this project:
// S : Sticker
// SS : StickerSet

func main() {
	parseCmdLine()
	core.Init()
}

func parseCmdLine() {
	var help = flag.Bool("help", false, "Show help")
	var adminUid = flag.Int64("admin_uid", -1, "Admin's UID")
	var botToken = flag.String("bot_token", "", "Telegram Bot Token")
	var dataDir = flag.String("data_dir", "", "temp data working dir")
	var webapp = flag.Bool("webapp", false, "Enable WebApp support")
	var webappUrl = flag.String("webapp_url", "", "URL to WebApp, HTTPS only")
	var webappApiUrl = flag.String("webapp_api_url", "", "URL to WebApp API server, HTTPS only")
	var WebappListenAddr = flag.String("webapp_listen_addr", "", "API listen addr(IP:PORT)")
	var webappDataDir = flag.String("webapp_data_dir", "", "Relative to CWD or absolute")
	var useDB = flag.Bool("use_db", false, "Use MariaDB")
	var dbAddr = flag.String("db_addr", "", "mariadb address")
	var dbUser = flag.String("db_user", "", "mariadb usernmae")
	var dbPass = flag.String("db_pass", "", "mariadb password")
	var logLevel = flag.String("log_level", "debug", "Log level")
	var localBotApiAddr = flag.String("local_botapi_addr", "", "Local Bot API Server Address")
	var localBotApiDir = flag.String("local_botapi_dir", "", "Local Bot API Working directory(Absolute)")
	var webhookPublicAddr = flag.String("webhook_public_addr", "", "Webhook public address(WebhookEndpoint).")
	var webhookListenAddr = flag.String("webhook_listen_addr", "", "Webhook listen address.")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	conf := core.ConfigTemplate{}

	conf.BotToken = *botToken
	if conf.BotToken == "" {
		log.Error("Please use --help.")
		log.Error("Please note that specifing BOT_TOKEN env var is no longer supported.")
		log.Fatal("No bot token provided!")
	}

	conf.UseDB = *useDB
	conf.DbAddr = *dbAddr
	conf.DbUser = *dbUser
	conf.DbPass = *dbPass

	conf.WebApp = *webapp
	conf.WebappUrl = *webappUrl
	// Defaults apiUrl to webappUrl
	if conf.WebappApiUrl == "" {
		conf.WebappApiUrl = *webappUrl
	} else {
		conf.WebappApiUrl = *webappApiUrl
	}
	conf.WebappDataDir = *webappDataDir
	conf.WebappListenAddr = *WebappListenAddr

	conf.LocalBotApiAddr = *localBotApiAddr
	conf.LocalBotApiDir = *localBotApiDir
	conf.WebhookPublicAddr = *webhookPublicAddr
	conf.WebhookListenAddr = *webhookListenAddr

	conf.LogLevel = *logLevel

	conf.AdminUid = *adminUid
	conf.DataDir = *dataDir

	core.Config = conf
}
