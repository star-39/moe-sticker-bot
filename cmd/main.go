package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/config"
	"github.com/star-39/moe-sticker-bot/pkg/core"
)

// main.go handles states and basic response,
// complex operations are done in other files.

func main() {
	parseCmdLine()
	core.Init()
}

func parseCmdLine() {
	var help = flag.Bool("help", false, "Show help")
	var adminUid = flag.Int64("admin_uid", -1, "Admin's UID")
	var botToken = flag.String("bot_token", "", "Telegram Bot Token")
	var webapp = flag.Bool("webapp", false, "Enable WebApp support")
	var webappUrl = flag.String("webapp_url", "", "URL to WebApp, HTTPS only")
	var webappApiUrl = flag.String("webapp_api_url", "", "URL to WebApp API server, HTTPS only")
	var WebappListenAddr = flag.String("webapp_listen_addr", "", "API listen addr(IP:PORT)")
	var webappRootDir = flag.String("webapp_root_dir", "", "Relative to CWD or absolute")
	var useDB = flag.Bool("use_db", false, "Use MariaDB")
	var dbAddr = flag.String("db_addr", "", "mariadb address")
	var dbUser = flag.String("db_user", "", "mariadb usernmae")
	var dbPass = flag.String("db_pass", "", "mariadb password")
	var logLevel = flag.String("log_level", "debug", "Log level")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	conf := config.ConfigTemplate{}

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
	conf.WebappRootDir = *webappRootDir
	conf.WebappListenAddr = *WebappListenAddr

	conf.LogLevel = *logLevel

	conf.AdminUid = *adminUid

	config.Config = conf
}
