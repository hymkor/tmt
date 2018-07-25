package main

import (
	"fmt"
	tw "github.com/zetamatta/go-tmaint"
	"time"
)

func follow(api *tw.Api, args []string) error {
	pageCh := api.GetFollowersListAll(nil)
	i := 0
	for p := range pageCh {
		if p.Error != nil {
			return p.Error
		}
		for _, u := range p.Followers {
			i++
			fmt.Printf("%6d %s %s %s\n", i, u.IdStr, u.ScreenName, u.Name)
			time.Sleep(time.Second)
		}
		fmt.Println()
		time.Sleep(time.Minute)
	}
	return nil
}
