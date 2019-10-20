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
	urls     []string
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

		row.urls = findUrlAll(&row.Tweet)
		for i, url1 := range row.urls {
			if i >= 10 {
				break
			}
			row.urls[i] = tco(url1)
			fmt.Fprintf(&buffer, "\n[%d] %s", i, row.urls[i])
		}
		row.contents = strings.Split(buffer.String(), "\n")
	}
	return row.contents
}

const (
	CTRL_M = "\x0D"
	CTRL_R = "\x12"
	CTRL_U = "\x15"
)

var rxUrl = regexp.MustCompile(`https?\:\/\/[[:graph:]]+`)

func findUrlAll(tw *anaconda.Tweet) []string {
	var text string
	if tw.RetweetedStatus != nil {
		text = tw.RetweetedStatus.FullText
	} else {
		text = tw.FullText
	}
	return rxUrl.FindAllString(text, -1)
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

func yesNo(p *twopane.Param, msg string) bool {
	p.Message("\x1B[33;1m" + msg + "\x1B[0m")
	ch, err := p.GetKey()
	return err == nil && strings.EqualFold(ch, "y")
}

func errorMessage(err error) string {
	var buffer strings.Builder
	buffer.WriteString("\x1B[35;1m")
	if e, ok := err.(*anaconda.ApiError); ok {
		for i, e1 := range e.Decoded.Errors {
			if i > 0 {
				fmt.Fprintln(&buffer)
			}
			fmt.Fprintf(&buffer, "[%d] %s", e1.Code, e1.Message)
		}
	} else {
		buffer.WriteString(err.Error())
	}
	buffer.WriteString("\x1B[0m")
	return buffer.String()
}

func fetch(getTimeline *Timeline, param *twopane.Param, already map[int64]struct{}) {
	timeline, err := getTimeline.Fetch()
	if err != nil {
		param.Message(errorMessage(err))
		peekKey(param)
		return
	}
	for i := len(timeline) - 1; i >= 0; i-- {
		if _, ok := already[timeline[i].Id]; !ok {
			param.View.Rows = append(param.View.Rows, &rowT{Tweet: timeline[i]})
			already[timeline[i].Id] = struct{}{}
		}
	}
	param.Cursor = len(param.View.Rows) - 1
}

func view(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timelines := map[string]*Timeline{
		"H": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return api.GetHomeTimeline(url.Values{})
			},
		},
		"R": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return api.GetMentionsTimeline(url.Values{})
			},
		},
		"L": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return api.GetFavorites(url.Values{})
			},
			Drop: func(id int64) error {
				_, err := api.Unfavorite(id)
				return err
			},
		},
		"U": &Timeline{
			Fetch: func() ([]anaconda.Tweet, error) {
				return myTimeline(api)
			},
		},
	}

	already := map[int64]struct{}{}
	getTimeline := timelines["H"]

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
		StatusLine: "[F1][?]Help [q]Quit [n]Post [r]Reply [l]Like [t]Retweet [.]Reload",
		Handler: func(param *twopane.Param) bool {
			switch param.Key {
			case "?", "F1":
				fmt.Fprint(param.Out, "\x1b[0J")
				fmt.Fprint(param.Out, `
[F1][?] This help
[Q] Quit
[J] Next Tweet
[K] Previous Tweet
[.] Load new Tweets
[Space] Page down
[Shift]+[H] Show Home Timeline
[Shift]+[R] Show Reply Timeline
[Shift]+[L] Show Favorites Timeline
[Shift]+[U] Show Your Tweets
[Ctrl]+[U] Show Current user's Timeline
[N] New Tweet
[L] Like
[R] Reply
[T] Retweet
[Shift]+[T] Retweet with comment
[Enter] Show this thread
[0]..[9] Open written URL`)
				param.GetKey()
			case "d":
				if getTimeline.Drop == nil {
					break
				}
				if row, ok := param.Rows[param.Cursor].(*rowT); ok {
					if !yesNo(param, "Remove this tweet ? [y/n]") {
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
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				if row1, ok := param.Rows[param.Cursor].(*rowT); ok {
					index := strings.Index("0123456789", param.Key)
					if index >= len(row1.urls) {
						break
					}
					urls := row1.urls
					if index == len(urls)-1 {
						// current tweet
						webbrowser.Open(urls[index])
						break
					}
					m := rxTweetStatusUrl.FindStringSubmatch(urls[index])
					if m != nil {
						if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
							insTweet(api, param, id)
							break
						}
					}
					webbrowser.Open(urls[index])
				}
			case "f", "l":
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
					if !yesNo(param, "Retweet ? [y/n]") {
						break
					}
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
						already[tw.Id] = struct{}{}
					}
				}
			case "n":
				post, err := doPost(api, "", nil)
				if err == nil {
					param.View.Rows = append(param.View.Rows, &rowT{Tweet: *post})
					param.Cursor = len(param.View.Rows) - 1
					already[post.Id] = struct{}{}
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
			case CTRL_U:
				if row, ok := param.View.Rows[param.Cursor].(*rowT); ok {
					getTimeline.Backup = param.Rows
					screenName := row.User.ScreenName
					getTimeline, ok = timelines[screenName]
					if !ok {
						getTimeline = &Timeline{
							Fetch: func() ([]anaconda.Tweet, error) {
								values := url.Values{}
								values.Add("screen_name", screenName)
								return api.GetUserTimeline(values)
							},
						}
						timelines[screenName] = getTimeline
					}
					param.Rows = getTimeline.Backup
					already = map[int64]struct{}{}
					fetch(getTimeline, param, already)

					param.Cursor = 0
					for i, newrow := range param.Rows {
						if newrow.(*rowT).Tweet.Id == row.Tweet.Id {
							param.Cursor = i
						}
					}
				}
			default: // change timeline
				if newTimline, ok := timelines[param.Key]; ok {
					getTimeline.Backup = param.Rows
					getTimeline = newTimline
					param.Rows = getTimeline.Backup
					already = map[int64]struct{}{}
					param.Cursor = len(param.View.Rows) - 1
				} else {
					break
				}
				fallthrough
			case ".", CTRL_R:
				fetch(getTimeline, param, already)
				break
			}
			return true
		},
	}.Run()
}
