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
	mine     bool
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
	return twopane.View{
		Rows:    rows,
		Reverse: true,
		Handler: func(param *twopane.Param) bool {
			switch param.Key {
			case "f":
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					tw, err := api.Favorite(row.Tweet.Id)
					if err == nil {
						param.Message("[Favorited]")
						row.Tweet = tw
						row.contents = nil
						row.mine = true
					} else {
						param.Message(err.Error())
					}
					if ch, err := param.GetKey(); err == nil {
						param.UnGetKey(ch)
					}
				}
			case "t":
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					tw, err := api.Retweet(row.Tweet.Id, false)
					if err == nil {
						param.Message("[Retweeted]")
						row.Tweet = tw
						row.contents = nil
						row.mine = true
					} else {
						param.Message(err.Error())
					}
					if ch, err := param.GetKey(); err == nil {
						param.UnGetKey(ch)
					}
				}
			case "T":
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					var buffer strings.Builder
					fmt.Fprintf(&buffer,
						"https://twitter.com/%s/status/%s\n%s",
						row.Tweet.User.ScreenName,
						row.Tweet.IdStr,
						row.Tweet.FullText)
					doPost(api, buffer.String(), nil)
				}
			case "n":
				post, err := doPost(api, "", nil)
				if err == nil {
					param.View.Rows = append(param.View.Rows, &rowT{Tweet: *post, mine: true})
				}
			case ".", CTRL_R:
				timeline, err := getTimeline()
				if err == nil {
					lastId := int64(0)
					for i := len(param.View.Rows) - 1; i >= 0; i-- {
						if !param.View.Rows[i].(*rowT).mine {
							lastId = param.View.Rows[i].(*rowT).Tweet.Id
							break
						}
					}
					for i := len(timeline) - 1; i >= 0; i-- {
						if timeline[i].Id > lastId {
							param.View.Rows = append(param.View.Rows, &rowT{Tweet: timeline[i]})
						}
					}
				}
			}
			return true
		},
	}.Run()
}

func view(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	return viewTimeline(api, func() ([]anaconda.Tweet, error) { return api.GetHomeTimeline(url.Values{}) })
}
