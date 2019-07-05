package main

import (
	"time"

	"github.com/nlopes/slack"
)

const (
	// GetMessageCount is count of messages to get
	GetMessageCount int = 1000
	// SleepTime is wait time to avoid api limit
	SleepTime int = 840
)

func cleanMessage(api *slack.Client, channel string) (err error) {
	params := slack.HistoryParameters{
		Count: GetMessageCount,
	}
	history, err := api.GetChannelHistory(channel, params)
	if err != nil {
		return err
	}

	for _, message := range history.Messages {
		api.DeleteMessage(channel, message.Timestamp)
		time.Sleep(time.Duration(SleepTime) * time.Millisecond)
	}

	return nil
}
