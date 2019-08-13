package main

import (
	"context"
	"fmt"
	"net/url"
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

func timeline(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	timeline, err := api.GetHomeTimeline(url.Values{})
	if err != nil {
		return err
	}
	for i := len(timeline); i > 0; i-- {
		t := timeline[i-1]

		fmt.Printf("From:\t%s <@%s>\n", t.User.Name, t.User.ScreenName)
		if t.InReplyToScreenName != "" {
			fmt.Printf("To:\t@%s\n", t.InReplyToScreenName)
		}
		fmt.Printf("Date:\t%s\n", globalTimeToLocal(t.CreatedAt))
		fmt.Println()
		fmt.Println(t.FullText)
		fmt.Println(".")
	}
	return nil
}
