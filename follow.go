package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/ChimeraCoder/anaconda"

	"github.com/zetamatta/tmt/ctrlc"
	tw "github.com/zetamatta/tmt/oauth"
)

var rxScreenName = regexp.MustCompile(`@\w+`)

func showUser(u *anaconda.User) {
	fmt.Printf("ID:%s\tScreenName:@%s\tName:%s\n", u.IdStr, u.ScreenName, u.Name)
}

func follow(ctx context.Context, api *tw.Api, args []string) error {
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
			if ctrlc.Sleep(ctx, time.Second*time.Duration(rand.Intn(100))/10) {
				return ctx.Err()
			}
		}
	}
	return nil
}

func listUsersSlowly(ctx context.Context, users []anaconda.User) error {
	for _, u := range users {
		showUser(&u)
		os.Stdout.Sync()
		if ctrlc.Sleep(ctx, time.Second*time.Duration(3)) {
			return ctx.Err()
		}
	}
	return nil
}

func followers(ctx context.Context, api *tw.Api, args []string) error {
	pageCh := api.GetFollowersListAll(nil)
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		if err := listUsersSlowly(ctx, p.Followers); err != nil {
			return err
		}
		fmt.Println()
	}
	return nil
}

func followings(ctx context.Context, api *tw.Api, args []string) error {
	pageCh := api.GetFriendsListAll(nil)
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		if err := listUsersSlowly(ctx, p.Friends); err != nil {
			return err
		}
	}
	return nil
}
