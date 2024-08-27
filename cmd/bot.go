package main

import (
	"context"
	"gr0/statusbot/status"
	"gr0/statusbot/status/util"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	token := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	debugMode := os.Getenv("DEBUG")
	ignoredUsers := os.Getenv("IGNORED_USERS")
	debug, err := strconv.ParseBool(debugMode)
	if err != nil {
		debug = false
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot := status.NewStatusBot(token, appToken, channelID, debug, util.ReadIgnored(ignoredUsers))
	bot.Run(ctx)
}
