package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
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

	db, err := sql.Open("mysql", "root:pass@tcp(localhost:3306)/slack")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ins, err := db.Prepare("INSERT INTO ng_words(word) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}

	del, err := db.Prepare("DELETE FROM ng_words WHERE word = ?")
	if err != nil {
		log.Fatal(err)
	}

	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			go func() {
				if ev.Msg.BotID == "" && ev.Msg.User != SlackbotID {
					if !(muteUserList[ev.Msg.User]) {
						space := regexp.MustCompile(`^ +`)
						rsvText := space.ReplaceAllString(ev.Msg.Text, "")

						sign := regexp.MustCompile(`[@#<>]`)

						mention := regexp.MustCompile(fmt.Sprintf(`^<@%s>`, BotID))

						if mention.MatchString(rsvText) {
							rsvText = mention.ReplaceAllString(rsvText, "")
							rsvText = space.ReplaceAllString(rsvText, "")

							cleanCommand := regexp.MustCompile(`^/clean`)
							muteCommand := regexp.MustCompile(`^/mute`)
							ngWordCommand := regexp.MustCompile(`^/ng`)

							if cleanCommand.MatchString(rsvText) {
								cleanMessage(api, ev.Msg.Channel)
							} else if muteCommand.MatchString(rsvText) {
								rsvText = muteCommand.ReplaceAllString(rsvText, "")
								rsvText = space.ReplaceAllString(rsvText, "")
								muteTarget := sign.ReplaceAllString(rsvText, "")
								if _, ok := muteUserList[muteTarget]; ok {
									if muteUserList[muteTarget] {
										muteUserList[muteTarget] = false
										postMessage(botAPI, GeneralID, fmt.Sprintf("<@%s> が <@%s> のミュートを解除しました", ev.Msg.User, muteTarget))
									} else {
										muteUserList[muteTarget] = true
										postMessage(botAPI, GeneralID, fmt.Sprintf("<@%s> が <@%s> をミュートにしました", ev.Msg.User, muteTarget))
									}
								} else {
									postMessage(botAPI, ev.Msg.Channel, fmt.Sprintf("<@%s> は存在しないユーザーです！", muteTarget))
								}
							} else if ngWordCommand.MatchString(rsvText) {
								rsvText = ngWordCommand.ReplaceAllString(rsvText, "")
								rsvText = space.ReplaceAllString(rsvText, "")

								ngWordAdd := regexp.MustCompile(`^add`)
								ngWordRemove := regexp.MustCompile(`^remove`)
								ngWordList := regexp.MustCompile(`^list`)
								if ngWordAdd.MatchString(rsvText) {
									rsvText = ngWordAdd.ReplaceAllString(rsvText, "")
									rsvText = space.ReplaceAllString(rsvText, "")

									if rsvText != "" {
										var ngWords []string

										rows, err := db.Query("SELECT word FROM ng_words")
										if err != nil {
											log.Fatal(err)
										}

										for rows.Next() {
											var ngWord string
											if err := rows.Scan(&ngWord); err != nil {
												log.Fatal(err)
											}
											ngWords = append(ngWords, ngWord)
										}

										unique := true
										for _, w := range ngWords {
											if rsvText == w {
												unique = false
												break
											}
										}

										if unique {
											ins.Exec(rsvText)
											botAPI.PostMessage(ev.Msg.Channel, slack.MsgOptionText("「"+rsvText+"」をNGワードに登録しました。", false))
										}
									}
								} else if ngWordRemove.MatchString(rsvText) {
									rsvText = ngWordRemove.ReplaceAllString(rsvText, "")
									rsvText = space.ReplaceAllString(rsvText, "")

									del.Exec(rsvText)
									botAPI.PostMessage(ev.Msg.Channel, slack.MsgOptionText("「"+rsvText+"」をNGワードから除外しました。", false))
								} else if ngWordList.MatchString(rsvText) {
									_, ts, _ := botAPI.PostMessage(ev.Msg.Channel, slack.MsgOptionText("NGワード一覧を表示します。", false))
									var ngWords []string

									rows, err := db.Query("SELECT word FROM ng_words")
									if err != nil {
										log.Fatal(err)
									}

									for rows.Next() {
										var ngWord string
										if err := rows.Scan(&ngWord); err != nil {
											log.Fatal(err)
										}
										ngWords = append(ngWords, ngWord)
									}

									for _, w := range ngWords {
										botAPI.PostMessage(ev.Msg.Channel, slack.MsgOptionText(w, false), slack.MsgOptionTS(ts))
									}
								}
							}
						} else {
							var ngWords []string

							rows, err := db.Query("SELECT word FROM ng_words")
							if err != nil {
								log.Fatal(err)
							}

							for rows.Next() {
								var ngWord string
								if err := rows.Scan(&ngWord); err != nil {
									log.Fatal(err)
								}
								ngWords = append(ngWords, ngWord)
							}

							for _, w := range ngWords {
								if strings.Index(ev.Msg.Text, w) != -1 {
									api.DeleteMessage(ev.Msg.Channel, ev.Msg.Timestamp)
								}
							}
						}
					} else {
						api.DeleteMessage(ev.Msg.Channel, ev.Msg.Timestamp)
						postMessage(botAPI, MutedUsersMessagesChannelID, fmt.Sprintf("<@%s> が <#%s> で発言しました:\n%s", ev.Msg.User, ev.Msg.Channel, ev.Msg.Text))
					}
				}
			}()
		case *slack.ChannelCreatedEvent:
			fmt.Printf("%sが%sを作成しました\n", ev.Channel.Creator, ev.Channel.ID)
			postMessage(botAPI, ChannelCreateReportID, fmt.Sprintf("<@%s> が <#%s> を作成しました\n", ev.Channel.Creator, ev.Channel.ID))
		}
	}
}
