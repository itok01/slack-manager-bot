package main

import (
	"log"
	"os"

	"github.com/nlopes/slack"
)

const (
	// SlackToken is OAuth Access Token
	SlackToken string = ""
	// SlackBotToken is Bot User OAuth Access Token
	SlackBotToken string = ""
	// BotID is Bot ID
	BotID string = ""
	// SlackbotID is slackbot ID
	SlackbotID string = ""
	// GeneralID is general channel ID
	GeneralID string = ""
	// ChannelCreateReportID is channel create report channel id
	ChannelCreateReportID string = ""
	// MutedUsersMessagesChannelID receive muted user's message
	MutedUsersMessagesChannelID string = ""
)

func main() {
	api := slack.New(SlackToken)
	if _, err := api.AuthTest(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	botAPI := slack.New(SlackBotToken)
	if _, err := botAPI.AuthTest(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	rtm(api, botAPI)
}
