package main

import (
	"github.com/nlopes/slack"
)

func postMessage(api *slack.Client, channel, text string) (err error) {
	msgText := slack.MsgOptionText(text, false)
	if _, _, err := api.PostMessage(channel, msgText); err != nil {
		return err
	}

	return nil
}
