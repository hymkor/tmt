package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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

func findUrlAll(tw *anaconda.Tweet) []string {
	var text string
	if tw.RetweetedStatus != nil {
		text = tw.RetweetedStatus.FullText
	} else {
		text = tw.FullText
	}
	list := rxUrl.FindAllString(text, -1)
	return append(list, toUrl(tw))
}

func findUrl(tw *anaconda.Tweet) string {
	return findUrlAll(tw)[0]
}

type Timeline struct {
	Fetch  func() ([]anaconda.Tweet, error)
	Backup []twopane.Row
	Drop   func(id int64) error
}

func tco(url string) string {

	limit := 10
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for {
		limit--
		if limit <= 0 {
			return url
		}

		var res *http.Response
		var err error

		if strings.HasPrefix(url, "https://t.co") ||
			strings.HasPrefix(url, "https://amzn.to") ||
			strings.HasPrefix(url, "http://amzn.to/") ||
			strings.HasPrefix(url, "https://youtu.be") ||
			strings.HasPrefix(url, "https://bit.ly") {
			res, err = client.Head(url)
		} else if strings.HasPrefix(url, "https://htn.to") ||
			strings.HasPrefix(url, "https://b.hatena.ne.jp/-/redirect") {
			res, err = client.Get(url)
		} else {
			break
		}

		if err != nil {
			return url
		}

		loc, err := res.Location()
		if !res.Close && res.Body != nil {
			res.Body.Close()
		}
		if err != nil {
			return url
		}
		url = loc.String()
	}
	return url
}

func peekKey(param *twopane.Param) {
	if ch, err := param.GetKey(); err == nil {
		param.UnGetKey(ch)
	}
}

func insTweet(api *anaconda.TwitterApi, param *twopane.Param, id int64) error {
	tw1, err := api.GetTweet(id, nil)
	if err != nil {
		return err
	}
	param.Rows = append(param.Rows, nil)
	copy(param.Rows[param.Cursor+1:], param.Rows[param.Cursor:])
	param.Rows[param.Cursor] = &rowT{Tweet: tw1}
	return nil
}

var rxTweetStatusUrl = regexp.MustCompile(`^https://twitter.com/\w+/status/(\d+)$`)

func errorMessage(err error) string {
	if e, ok := err.(*anaconda.ApiError); ok {
		var buffer strings.Builder
		for _, e1 := range e.Decoded.Errors {
			if buffer.Len() > 0 {
				fmt.Fprintln(&buffer)
			}
			fmt.Fprintf(&buffer, "[%d] %s", e1.Code, e1.Message)
		}
		return buffer.String()
	} else {
		return err.Error()
	}
}

func view(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timelines := map[string]*Timeline{
		"h": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return api.GetHomeTimeline(url.Values{})
			},
		},
		"n": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return api.GetMentionsTimeline(url.Values{})
			},
		},
		"f": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return api.GetFavorites(url.Values{})
			},
			Drop: func(id int64) error {
				_, err := api.Unfavorite(id)
				return err
			},
		},
		"u": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return myTimeline(api)
			},
		},
	}

	already := map[int64]struct{}{}
	getTimeline := timelines["h"]

	timeline, err := getTimeline.Fetch()
	if err != nil {
		return err
	}
	rows := make([]twopane.Row, 0, len(timeline))
	for i := len(timeline) - 1; i >= 0; i-- {
		if _, ok := already[timeline[i].Id]; !ok {
			rows = append(rows, &rowT{Tweet: timeline[i]})
			already[timeline[i].Id] = struct{}{}
		}
	}
	var me *anaconda.User

	return twopane.View{
		Rows:       rows,
		Reverse:    true,
		StatusLine: "[q]Quit [n]post [f]Like [t]Retweet [T]Comment [.]Reload [C-c]CopyURL [o]OpenURL [CR]MoveThread",
		Handler: func(param *twopane.Param) bool {
			switch param.Key {
			case "d":
				if getTimeline.Drop == nil {
					break
				}
				if row, ok := param.Rows[param.Cursor].(*rowT); ok {
					param.Message("Remove this tweet ? [y/n]")
					if ch, err := param.GetKey(); err != nil || ch != "y" {
						break
					}
					if err := getTimeline.Drop(row.Id); err != nil {
						param.Message(errorMessage(err))
						peekKey(param)
						break
					}
					copy(param.Rows[param.Cursor:], param.Rows[param.Cursor+1:])
					param.Rows = param.Rows[:len(param.Rows)-1]
					if param.Cursor > 0 {
						param.Cursor--
					}
				}
			case CTRL_M:
				if row, ok := param.Rows[param.Cursor].(*rowT); ok {
					tw := &row.Tweet
					if tw.RetweetedStatus != nil {
						tw = tw.RetweetedStatus
					}
					if tw.InReplyToStatusID > 0 {
						insTweet(api, param, tw.InReplyToStatusID)
					}
				}
			case "o":
				tw := &param.Rows[param.Cursor].(*rowT).Tweet
				url := findUrlAll(tw)
				var msg strings.Builder
				msg.WriteString("Open")
				for i, url1 := range url {
					if i >= 10 {
						break
					}
					url[i] = tco(url1)
					fmt.Fprintf(&msg, "\n[%d] %s", i, url[i])
				}
				msg.WriteString(" ?")
				param.Message(msg.String())
				if ch, err := param.GetKey(); err == nil {
					if index := strings.Index("0123456789", ch); index >= 0 {
						if index == len(url)-1 {
							// current tweet
							webbrowser.Open(url[index])
							break
						}
						m := rxTweetStatusUrl.FindStringSubmatch(url[index])
						if m != nil {
							if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
								insTweet(api, param, id)
								break
							}
						}
						webbrowser.Open(url[index])
					}
				}
			case CTRL_C:
				tw := &param.Rows[param.Cursor].(*rowT).Tweet
				url := findUrl(tw)
				param.Message("[Copy] " + url)
				clipboard.WriteAll(url)
				peekKey(param)
			case "f":
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					tw, err := api.Favorite(row.Tweet.Id)
					if err == nil {
						param.Message("[Favorited]")
						row.Tweet = tw
						row.contents = nil
					} else {
						param.Message(errorMessage(err))
					}
					peekKey(param)
				}
			case "t":
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					tw, err := api.Retweet(row.Tweet.Id, false)
					if err == nil {
						param.Message("[Retweeted]")
						param.View.Rows = append(param.View.Rows, &rowT{
							Tweet: tw,
						})
					} else {
						param.Message(errorMessage(err))
					}
					peekKey(param)
				} else {
					break
				}
			case "T":
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					var buffer strings.Builder
					pasteUrl(&buffer, &row.Tweet)
					fmt.Fprintf(&buffer, "\n%s", row.Tweet.FullText)

					if tw, err := doPost(api, buffer.String(), nil); err == nil {
						param.View.Rows = append(param.View.Rows, &rowT{
							Tweet: *tw,
						})
					}
				}
			case "n":
				post, err := doPost(api, "", nil)
				if err == nil {
					param.View.Rows = append(param.View.Rows, &rowT{Tweet: *post})
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
						})
						param.Cursor = len(param.View.Rows) - 1
					}
				}
			case "g":
				param.Message("[h]Home [n]Mention [f]Like")
				if ch, err := param.GetKey(); err == nil {
					if newTimline, ok := timelines[ch]; ok {
						getTimeline.Backup = param.Rows
						getTimeline = newTimline
						param.Rows = getTimeline.Backup
						param.Cursor = len(param.View.Rows) - 1
					}
				}
				fallthrough
			case ".", CTRL_R:
				timeline, err := getTimeline.Fetch()
				if err != nil {
					param.Message(errorMessage(err))
					peekKey(param)
					break
				}
				for i := len(timeline) - 1; i >= 0; i-- {
					if _, ok := already[timeline[i].Id]; !ok {
						param.View.Rows = append(param.View.Rows, &rowT{Tweet: timeline[i]})
						already[timeline[i].Id] = struct{}{}
					}
				}
				param.Cursor = len(param.View.Rows) - 1
			}
			return true
		},
	}.Run()
}
