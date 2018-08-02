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

func showUser(u *anaconda.User) {
	fmt.Printf("ID:%s\tScreenName:@%s\tName:%s\n", u.IdStr, u.ScreenName, u.Name)
}

func follow(api *tw.Api, args []string) error {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		match1 := rxScreenName.FindString(sc.Text())
		if match1 != "" {
			screenName := match1[1:]
			u, err := api.FollowUser(screenName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", screenName, err.Error())
			} else {
				showUser(&u)
				os.Stdout.Sync()
			}
			time.Sleep(time.Second * time.Duration(3))
		}
	}
	return nil
}

func listUsersSlowly(users []anaconda.User) {
	for _, u := range users {
		showUser(&u)
		os.Stdout.Sync()
		time.Sleep(time.Second * time.Duration(3))
	}
}

func followers(api *tw.Api, args []string) error {
	pageCh := api.GetFollowersListAll(nil)
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		listUsersSlowly(p.Followers)
		fmt.Println()
	}
	return nil
}

func followings(api *tw.Api, args []string) error {
	pageCh := api.GetFriendsListAll(nil)
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		listUsersSlowly(p.Friends)
	}
	return nil
}
