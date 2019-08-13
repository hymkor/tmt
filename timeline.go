package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mattn/go-colorable"
)

const _TIME_LAYOUT = "Mon Jan 02 15:04:05 -0700 2006"

func globalTimeToLocal(org string) string {
	dt, err := time.Parse(_TIME_LAYOUT, org)
	if err != err {
		return org
	}
	return dt.Local().Format(_TIME_LAYOUT)
}

func catTweet(t anaconda.Tweet, bon, boff string, w io.Writer) {
	fmt.Fprintf(w, "%sFrom:%s\t%s <@%s>\n", bon, boff, t.User.Name, t.User.ScreenName)
	if t.InReplyToScreenName != "" {
		fmt.Fprintf(w, "%sTo:%s\t@%s\n", bon, boff, t.InReplyToScreenName)
		if t.InReplyToStatusIdStr != "" {
			fmt.Fprintf(w,
				"%sIn-Reply-To:%s https://twitter.com/%s/status/%s\n",
				bon,
				boff,
				t.InReplyToScreenName,
				t.InReplyToStatusIdStr)
		}
	}
	fmt.Fprintf(w, "%sDate:%s\t%s\n", bon, boff, globalTimeToLocal(t.CreatedAt))
	fmt.Fprintln(w)
	fmt.Fprintln(w, t.FullText)
	fmt.Fprintln(w, ".")
}

func timeline(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := api.GetHomeTimeline(url.Values{})
	if err != nil {
		return err
	}
	w := colorable.NewColorableStdout()
	for i := len(timeline); i > 0; i-- {
		catTweet(timeline[i-1], "\x1B[0;32;1m", "\x1B[0m", w)
	}
	return nil
}
