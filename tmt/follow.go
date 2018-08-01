package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"time"

	tw "github.com/zetamatta/go-tmaint"
)

var rxScreenName = regexp.MustCompile(`@\w+`)

func follow(api *tw.Api, args []string) error {
	sc := bufio.NewScanner(os.Stdin)
	i := 0
	for sc.Scan() {
		match1 := rxScreenName.FindString(sc.Text())
		if match1 != "" {
			screenName := match1[1:]
			u, err := api.FollowUser(screenName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", screenName, err.Error())
			} else {
				i++
				fmt.Printf("%6d %s @%s %s\n", i, u.IdStr, u.ScreenName, u.Name)
				os.Stdout.Sync()
			}
			time.Sleep(time.Second * time.Duration(3))
		}
	}
	return nil
}

func followers(api *tw.Api, args []string) error {
	pageCh := api.GetFollowersListAll(nil)
	i := 0
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		for _, u := range p.Followers {
			i++
			fmt.Printf("%6d %s @%s %s\n", i, u.IdStr, u.ScreenName, u.Name)
			os.Stdout.Sync()
			time.Sleep(time.Second * time.Duration(3))
		}
		fmt.Println()
	}
	return nil
}
