package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"time"

	tw "github.com/zetamatta/go-tmaint"
)

var rxNumber = regexp.MustCompile(`\d+`)

func dump(api *tw.Api, args []string) error {
	for i, arg1 := range args {
		if i > 0 {
			time.Sleep(time.Second * time.Duration(3))
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
