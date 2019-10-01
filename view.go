package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/atotto/clipboard"
	"github.com/toqueteos/webbrowser"

	"github.com/zetamatta/go-twopane"
)

func pasteUrl(buffer io.Writer, t *anaconda.Tweet) {
	fmt.Fprintf(buffer, "https://twitter.com/%s/status/%s",
		t.User.ScreenName,
		t.IdStr)
}

func toUrl(t *anaconda.Tweet) string {
	var buffer strings.Builder
	pasteUrl(&buffer, t)
	return buffer.String()
}

type rowT struct {
	anaconda.Tweet
	contents []string
	mine     bool
	title    string
}

func (row *rowT) Title(_ interface{}) string {
	if row.title == "" {
		row.title = fmt.Sprintf("\x1B[32m%s\x1B[37;1m %s",
			row.Tweet.User.ScreenName,
			strings.Replace(html.UnescapeString(row.Tweet.FullText), "\n", " ", -1))
	}
	return row.title
}

func (row *rowT) Contents(_ interface{}) []string {
	if row.contents == nil {
		var buffer strings.Builder
		catTweet(&row.Tweet, "\x1B[0;32m", "\x1B[0m", &buffer)
		row.contents = strings.Split(buffer.String(), "\n")
	}
	return row.contents
}

const (
	CTRL_C = "\x03"
	CTRL_M = "\x0D"
	CTRL_R = "\x12"
)

var rxUrl = regexp.MustCompile(`https?\:\/\/[[:graph:]]+`)

func findUrl(tw *anaconda.Tweet) string {
	var text string
	if tw.RetweetedStatus != nil {
		text = tw.RetweetedStatus.FullText
	} else {
		text = tw.FullText
	}
	if m := rxUrl.FindString(text); m != "" {
		return m
	} else {
		return toUrl(tw)
	}
}

func viewTimeline(api *anaconda.TwitterApi, getTimeline func() ([]anaconda.Tweet, error)) error {
	timeline, err := getTimeline()
	if err != nil {
		return err
	}
	rows := make([]twopane.Row, 0, len(timeline))
	for i := len(timeline) - 1; i >= 0; i-- {
		rows = append(rows, &rowT{Tweet: timeline[i]})
	}
	var me *anaconda.User

	return twopane.View{
		Rows:       rows,
		Reverse:    true,
		StatusLine: "[q]Quit [n]post [f]Like [t]Retweet [T]Comment [.]Reload [C-c]CopyURL [o]OpenURL [CR]MoveThread",
		Handler: func(param *twopane.Param) bool {
			switch param.Key {
			case CTRL_M:
				if row, ok := param.Rows[param.Cursor].(*rowT); ok {
					tw := &row.Tweet
					if tw.RetweetedStatus != nil {
						tw = tw.RetweetedStatus
					}
					if tw.InReplyToStatusID > 0 {
						if tw1, err := api.GetTweet(tw.InReplyToStatusID, nil); err == nil {
							param.Rows = append(param.Rows, nil)
							copy(param.Rows[param.Cursor+1:], param.Rows[param.Cursor:])
							param.Rows[param.Cursor] = &rowT{Tweet: tw1}
						}
					}
				}
			case "o":
				tw := &param.Rows[param.Cursor].(*rowT).Tweet
				url := findUrl(tw)
				param.Message("Open " + url + " ? [Y/N]")
				if ch, err := param.GetKey(); err == nil && strings.EqualFold(ch, "y") {
					webbrowser.Open(url)
				}
			case CTRL_C:
				tw := &param.Rows[param.Cursor].(*rowT).Tweet
				url := findUrl(tw)
				param.Message("[Copy] " + url)
				clipboard.WriteAll(url)
				if ch, err := param.GetKey(); err == nil {
					param.UnGetKey(ch)
				}
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
						param.View.Rows = append(param.View.Rows, &rowT{
							Tweet: tw,
							mine:  true,
						})
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
					pasteUrl(&buffer, &row.Tweet)
					fmt.Fprintf(&buffer, "\n%s", row.Tweet.FullText)

					if tw, err := doPost(api, buffer.String(), nil); err == nil {
						param.View.Rows = append(param.View.Rows, &rowT{
							Tweet: *tw,
							mine:  true,
						})
					}
				}
			case "n":
				post, err := doPost(api, "", nil)
				if err == nil {
					param.View.Rows = append(param.View.Rows, &rowT{Tweet: *post, mine: true})
					param.Cursor = len(param.View.Rows) - 1
				}
			case "r":
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					if me == nil {
						meTmp, err := api.GetSelf(nil)
						if err == nil {
							me = &meTmp
						}
					}

					draft := ""
					if me == nil || me.Id != row.User.Id {
						draft = fmt.Sprintf("@%s ", row.User.ScreenName)
					}
					values := url.Values{}
					values.Add("in_reply_to_status_id", row.IdStr)
					if tw, err := doPost(api, draft, values); err == nil {
						param.View.Rows = append(param.View.Rows, &rowT{
							Tweet: *tw,
							mine:  true,
						})
						param.Cursor = len(param.View.Rows) - 1
					}
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
					param.Cursor = len(param.View.Rows) - 1
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
