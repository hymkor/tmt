package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zetamatta/go-tmaint"
	"github.com/zetamatta/go-tmaint/secret"
)

func main1(args []string) error {
	if len(args) <= 0 {
		return nil
	}
	api, err := tmaint.Login(secret.ConsumerKey, secret.ConsumerSecret)
	if err != nil {
		return err
	}
	defer api.Close()

	switch args[0] {
	case "lsfollow":
		return lsfollow(api, args[1:])
	case "dofollow":
		return dofollow(api, args[1:])
	case "cat":
		return cat(api, args[1:])
	default:
		return fmt.Errorf("%s: no such sub-command", args[0])
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
