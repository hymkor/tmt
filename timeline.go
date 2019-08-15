package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"regexp"
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

var rxDotsLine = regexp.MustCompile(`(?m)^\.+$`)

func catTweet(t *anaconda.Tweet, bon, boff string, w io.Writer) {
	if t.RetweetedStatus != nil {
		fmt.Fprintf(w, "%sRetweeted-By%s: %s <@%s>\n",
			bon, boff, t.User.Name, t.User.ScreenName)
		t = t.RetweetedStatus
	}
	fmt.Fprintf(w, "%sFrom:%s %s <@%s>\n", bon, boff, t.User.Name, t.User.ScreenName)
	fmt.Fprintf(w, "%sMessage-ID:%s twitter.com/%s/status/%s\n", bon, boff, t.User.ScreenName, t.IdStr)
	if t.InReplyToScreenName != "" {
		fmt.Fprintf(w, "%sTo:%s @%s\n", bon, boff, t.InReplyToScreenName)
		if t.InReplyToStatusIdStr != "" {
			fmt.Fprintf(w,
				"%sIn-Reply-To:%s twitter.com/%s/status/%s\n",
				bon,
				boff,
				t.InReplyToScreenName,
				t.InReplyToStatusIdStr)
		}
	}
	fmt.Fprintf(w, "%sDate:%s %s\n", bon, boff, globalTimeToLocal(t.CreatedAt))
	fmt.Fprintln(w)
	fmt.Fprintln(w, rxDotsLine.ReplaceAllStringFunc(t.FullText, func(s string) string {
		return s + "."
	}))
	fmt.Fprintln(w, ".")
}

func callPager() (io.Writer, func(), error) {
	callPager := os.Getenv("Pager")
	if callPager == "" {
		return colorable.NewColorableStdout(), func() {}, nil
	}
	cmd := exec.Command(callPager)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	out, err := cmd.StdinPipe()
	if err != nil {
		return colorable.NewColorableStdout(), func() {}, err
	}
	cmd.Start()
	return out, func() { out.Close(); cmd.Wait() }, nil
}

func showTimeline(timeline []anaconda.Tweet) {
	w, closer, _ := callPager()
	for i := len(timeline); i > 0; i-- {
		catTweet(&timeline[i-1], "\x1B[0;32;1m", "\x1B[0m", w)
	}
	closer()
}

func timeline(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := api.GetHomeTimeline(url.Values{})
	if err != nil {
		return err
	}
	showTimeline(timeline)
	return nil
}

func mention(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := api.GetMentionsTimeline(url.Values{})
	if err != nil {
		return err
	}
	showTimeline(timeline)
	return nil
}

func said(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := myTimeline(api)
	if err != nil {
		return err
	}
	showTimeline(timeline)
	return nil
}
