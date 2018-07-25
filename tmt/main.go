package main

import (
	"fmt"
	"os"

	tw "github.com/zetamatta/go-tmaint"
)

func main1(args []string) error {
	if len(args) <= 0 {
		return nil
	}
	api, _, err := tw.Login()
	if err != nil {
		return err
	}
	defer api.Close()

	switch args[0] {
	case "lsfollow":
		return lsfollow(api, args[1:])
	case "dofollow":
		return dofollow(api, args[1:])
	default:
	}
	return nil
}

func main() {
	if err := main1(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
