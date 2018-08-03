package main

import (
	"flag"
	"fmt"
	"os"

	tmaint "github.com/zetamatta/tmt/oauth"
	"github.com/zetamatta/tmt/secret"
)

type subCommandT struct {
	F func(*tmaint.Api, []string) error
	U string
}

var subcommands = map[string]*subCommandT{
	"followers":  {followers, " ... list members you are followed"},
	"followings": {followings, " ... list members you follows"},
	"follow":     {follow, "... follow person listed in STDIN"},
	"dump":       {dump, "IDNum ... dump JSON for the tweet"},
}

func main1(args []string) error {
	if len(args) <= 0 {
		exename, err := os.Executable()
		if err == nil {
			fmt.Fprintf(os.Stderr, "Usage:\n %s SUBCOMMAND ...\n", exename)
		}
		for name, value := range subcommands {
			fmt.Fprintf(os.Stderr, "\t%s %s\n", name, value.U)
		}
		return nil
	}
	api, err := tmaint.Login(secret.ConsumerKey, secret.ConsumerSecret)
	if err != nil {
		return err
	}
	defer api.Close()

	subcommand1, ok := subcommands[args[0]]
	if !ok {
		return fmt.Errorf("%s: no such sub-command", args[0])
	}
	return subcommand1.F(api, args[1:])
}

func main() {
	flag.Parse()
	if err := main1(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
