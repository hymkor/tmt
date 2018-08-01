package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/ChimeraCoder/anaconda"

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

func listUsersSlowly(i *int, users []anaconda.User) {
	for _, u := range users {
		(*i)++
		fmt.Printf("%6d %s @%s %s\n", *i, u.IdStr, u.ScreenName, u.Name)
		os.Stdout.Sync()
		time.Sleep(time.Second * time.Duration(3))
	}
}

func followers(api *tw.Api, args []string) error {
	pageCh := api.GetFollowersListAll(nil)
	i := 0
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		listUsersSlowly(&i, p.Followers)
		fmt.Println()
	}
	return nil
}

func followings(api *tw.Api, args []string) error {
	pageCh := api.GetFriendsListAll(nil)
	i := 0
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		listUsersSlowly(&i, p.Friends)
	}
	return nil
}
