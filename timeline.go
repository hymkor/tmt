package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

const _TIME_LAYOUT = "Mon Jan 02 15:04:05 -0700 2006"

func globalTimeToLocal(org string) string {
	dt, err := time.Parse(_TIME_LAYOUT, org)
	if err != err {
		return org
	}
	return dt.Local().Format(_TIME_LAYOUT)
}

func catTweet(t anaconda.Tweet, w io.Writer) {
	fmt.Fprintf(w, "From:\t%s <@%s>\n", t.User.Name, t.User.ScreenName)
	if t.InReplyToScreenName != "" {
		fmt.Fprintf(w, "To:\t@%s\n", t.InReplyToScreenName)
	}
	fmt.Fprintf(w, "Date:\t%s\n", globalTimeToLocal(t.CreatedAt))
	fmt.Fprintln(w)
	fmt.Fprintln(w, t.FullText)
	fmt.Fprintln(w, ".")
}

func timeline(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := api.GetHomeTimeline(url.Values{})
	if err != nil {
		return err
	}
	for i := len(timeline); i > 0; i-- {
		catTweet(timeline[i-1], os.Stdout)
	}
	return nil
}
