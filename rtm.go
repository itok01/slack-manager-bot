package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/nlopes/slack"
)

func rtm(api, botAPI *slack.Client) {
	rtm := botAPI.NewRTM()
	go rtm.ManageConnection()

	muteUserList := map[string]bool{}
	users, err := api.GetUsers()
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		muteUserList[user.ID] = false
	}

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			go func() {
				if ev.Msg.BotID == "" && ev.Msg.User != SlackbotID {
					space := regexp.MustCompile(`^ +`)
					rsvText := space.ReplaceAllString(ev.Msg.Text, "")

					sign := regexp.MustCompile(`[@#<>]`)

					mention := regexp.MustCompile(fmt.Sprintf(`^<@%s>`, BotID))

					if mention.MatchString(rsvText) && !(muteUserList[ev.Msg.User]) {
						if _, _, err := api.DeleteMessage(ev.Msg.Channel, ev.Msg.Timestamp); err != nil {
							fmt.Println(err)
						}
						rsvText = mention.ReplaceAllString(rsvText, "")
						rsvText = space.ReplaceAllString(rsvText, "")

						cleanCommand := regexp.MustCompile(`^/clean`)
						muteCommand := regexp.MustCompile(`^/mute`)

						if cleanCommand.MatchString(rsvText) {
							cleanMessage(api, ev.Msg.Channel)
						} else if muteCommand.MatchString(rsvText) {
							rsvText = muteCommand.ReplaceAllString(rsvText, "")
							rsvText = space.ReplaceAllString(rsvText, "")
							muteTarget := sign.ReplaceAllString(rsvText, "")
							if muteUserList[muteTarget] {
								muteUserList[muteTarget] = false
								postMessage(botAPI, GeneralID, fmt.Sprintf("<@%s> のミュートを解除しました", muteTarget))
							} else {
								muteUserList[muteTarget] = true
								postMessage(botAPI, GeneralID, fmt.Sprintf("<@%s> をミュートにしました", muteTarget))
							}
						}
					} else {
						api.DeleteMessage(ev.Msg.Channel, ev.Msg.Timestamp)
					}
				}
			}()
		case *slack.ChannelCreatedEvent:
			fmt.Printf("%sが%sを作成しました\n", ev.Channel.Creator, ev.Channel.ID)
			postMessage(botAPI, ChannelCreateReportID, fmt.Sprintf("<@%s> が <#%s> を作成しました\n", ev.Channel.Creator, ev.Channel.ID))
		}
	}
}
