package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zetamatta/experimental/mytwitter"
)

var sleepSecond = flag.Int64("ss", 1, "sleep seconds")

func rmTweets(api *mytwitter.Api, r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		idStr := strings.TrimSpace(sc.Text())
		idStr = strings.Trim(idStr, `"`)
		idNum, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", idStr, err.Error())
			continue
		}
		_, err = api.DeleteTweet(idNum, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", idStr, err.Error())
		} else {
			fmt.Printf("id %d done\n", idNum)
		}
		if *sleepSecond > 0 {
			time.Sleep(time.Duration(*sleepSecond) * time.Second)
		}
	}
}

func main1(args []string) error {
	api, _, err := mytwitter.Login()
	if err != nil {
		return err
	}
	defer api.Close()

	if len(args) <= 0 {
		rmTweets(api, os.Stdin)
		return nil
	}
	for _, fname := range args {
		fd, err := os.Open(fname)
		if err != nil {
			return err
		}
		rmTweets(api, fd)
		fd.Close()
	}
	return nil
}

func main() {
	flag.Parse()
	if err := main1(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
