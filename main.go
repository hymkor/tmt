package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/zetamatta/tmt/ctrlc"
	tmaint "github.com/zetamatta/tmt/oauth"
	"github.com/zetamatta/tmt/secret"
)

type subCommandT struct {
	F func(context.Context, *tmaint.Api, []string) error
	U string
}

var subcommands = map[string]*subCommandT{
	"followers":  {followers, " ... list members you are followed"},
	"followings": {followings, " ... list members you follows"},
	"follow":     {follow, "... follow people listed in STDIN\n\t  (Write like @ScreenName, ignore others)"},
	"unfollow":   {unfollow, "... unfollow people listed in STDIN\n\t  (Write like @ScreenName, ignore others)"},
	"dump":       {dump, "IDNum ... dump JSON for the tweet"},
	"post":       {post, "... post tweet from STDIN"},
	"cont":       {cont, "... post continued tweet from STDIN"},
	"whoami":     {whoami, "... show who are you"},
	"timeline":   {timeline, "... get home timeline"},
	"mention":    {mention, "... get mention timeline"},
	"reply":      {reply, "IDNum ... reply to IDNum"},
}

func main1(args []string) error {
	if len(args) <= 0 {
		exename, err := os.Executable()
		if err == nil {
			fmt.Fprintf(os.Stderr, "Usage:\n  %s COMMAND...\n", exename)
		}
		for name, value := range subcommands {
			fmt.Fprintf(os.Stderr, "\t%s %s\n", name, value.U)
		}
		if cfgPath, err := tmaint.ConfigurationPath(); err != nil {
			return err
		} else {
			fmt.Fprintf(os.Stderr, "Your configuration is saved on\n  %s\n", cfgPath)
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
	ctx, closer := ctrlc.Setup(context.Background())
	rc := subcommand1.F(ctx, api, args[1:])
	closer()
	return rc
}

func main() {
	flag.Parse()
	if err := main1(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
