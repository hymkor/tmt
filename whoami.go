package main

import (
	"context"
	"fmt"

	tw "github.com/zetamatta/tmt/oauth"
)

func whoami(ctx context.Context, api *tw.Api, args []string) error {
	u, err := api.GetSelf(nil)
	if err != nil {
		return err
	}
	fmt.Println(u.ScreenName)
	return nil
}
