package main

import (
	"context"
	"io/ioutil"
	"os"

	tw "github.com/zetamatta/tmt/oauth"
)

func post(ctx context.Context, api *tw.Api, args []string) error {
	text, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	_, err = api.PostTweet(string(text), nil)
	return err
}
