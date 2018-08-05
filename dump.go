package main

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/zetamatta/tmt/ctrlc"
	tw "github.com/zetamatta/tmt/oauth"
)

var rxNumber = regexp.MustCompile(`\d+`)

func dump(ctx context.Context, api *tw.Api, args []string) error {
	for i, arg1 := range args {
		if i > 0 && ctrlc.Sleep(ctx, time.Second*time.Duration(3)) {
			return ctx.Err()
		}
		idStr := rxNumber.FindString(arg1)
		if idStr == "" {
			continue
		}

		idNum, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return err
		}

		tweet, err := api.GetTweet(idNum, nil)
		if err != nil {
			return err
		}
		bin, err := json.MarshalIndent(&tweet, "", "  ")
		if err != nil {
			return err
		}
		os.Stdout.Write(bin)
	}
	return nil
}
