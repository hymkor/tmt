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
	*anaconda.Tweet
	contents []string
}

func (row *rowT) Title() string {
	return fmt.Sprintf("@%s %s", row.Tweet.User.ScreenName, row.Tweet.FullText)
}

func (row *rowT) Contents() []string {
	if row.contents == nil {
		var buffer strings.Builder
		catTweet(row.Tweet, "", "", &buffer)
		row.contents = strings.Split(buffer.String(), "\n")
	}
	return row.contents
}

func viewTimeline(timeline []anaconda.Tweet, handler func(*twopane.View, string) bool) error {
	rows := make([]twopane.Row, 0, len(timeline))
	for i := range timeline {
		rows = append(rows, &rowT{Tweet: &timeline[i]})
	}
	return twopane.View{Rows: rows, Clear: true, Handler: handler}.Run()
}

func view(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := api.GetHomeTimeline(url.Values{})
	if err != nil {
		return err
	}
	return viewTimeline(timeline, nil)
}
