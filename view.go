package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ChimeraCoder/anaconda"

	"github.com/zetamatta/go-twopane"
)

type rowT struct {
	anaconda.Tweet
	contents []string
}

func (row *rowT) Title() string {
	return fmt.Sprintf("\x1B[32;1m%s\x1B[37;1m %s", row.Tweet.User.ScreenName, row.Tweet.FullText)
}

func (row *rowT) Contents() []string {
	if row.contents == nil {
		var buffer strings.Builder
		catTweet(&row.Tweet, "\x1B[0;32;1m", "\x1B[0m", &buffer)
		row.contents = strings.Split(buffer.String(), "\n")
	}
	return row.contents
}

const (
	CTRL_R = "\x12"
)

func viewTimeline(api *anaconda.TwitterApi, getTimeline func() ([]anaconda.Tweet, error)) error {
	timeline, err := getTimeline()
	if err != nil {
		return err
	}
	rows := make([]twopane.Row, 0, len(timeline))
	for i := len(timeline) - 1; i >= 0; i-- {
		rows = append(rows, &rowT{Tweet: timeline[i]})
	}
	for {
		var nextaction func() error
		err := twopane.View{
			Rows:    rows,
			Reverse: true,
			Handler: func(param *twopane.Param) bool {
				switch param.Key {
				case "t":
					if row, ok := rows[param.Cursor].(*rowT); ok {
						_, err := api.Retweet(row.Tweet.Id, false)
						if err == nil {
							param.Message("[Retweeted]")
						} else {
							param.Message(err.Error())
						}
						if ch, err := param.GetKey(); err == nil {
							param.UnGetKey(ch)
						}
					}
					return true
				case "T":
					if row, ok := rows[param.Cursor].(*rowT); ok {
						var buffer strings.Builder
						fmt.Fprintf(&buffer,
							"https://twitter.com/%s/status/%s\n%s",
							row.Tweet.User.ScreenName,
							row.Tweet.IdStr,
							row.Tweet.FullText)
						doPost(api, buffer.String(), nil)
					}
					return true
				case "n":
					post, err := doPost(api, "", nil)
					if err == nil {
						param.View.Rows = append(param.View.Rows, &rowT{Tweet: *post})
					}
					return true
				case ".", CTRL_R:
					nextaction = func() error {
						timeline, err := getTimeline()
						if err != nil {
							return err
						}
						lastId := rows[len(rows)-1].(*rowT).Tweet.Id
						for i := len(timeline) - 1; i >= 0; i-- {
							if timeline[i].Id > lastId {
								rows = append(rows, &rowT{Tweet: timeline[i]})
							}
						}
						return nil
					}
					return false
				default:
					return true
				}
			},
		}.Run()

		if err != nil {
			return err
		}
		if nextaction == nil {
			return nil
		}
		if err := nextaction(); err != nil {
			return err
		}
	}
}

func view(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	return viewTimeline(api, func() ([]anaconda.Tweet, error) { return api.GetHomeTimeline(url.Values{}) })
}
