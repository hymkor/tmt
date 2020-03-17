package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ChimeraCoder/anaconda"
	"github.com/urfave/cli/v2"

	"github.com/zetamatta/tmt/ctrlc"
	tmaint "github.com/zetamatta/tmt/oauth"
	"github.com/zetamatta/tmt/secret"
)

func mains(args []string) error {
	if len(args) <= 1 {
		if cfgPath, err := tmaint.ConfigurationPath(); err != nil {
			return err
		} else {
			defer fmt.Fprintf(os.Stderr, "Your configuration is saved on\n  %s\n", cfgPath)
		}
	}

	ctx, closer := ctrlc.Setup(context.Background())

	action1 := func(f func(c context.Context, api *anaconda.TwitterApi, args []string) error) func(*cli.Context) error {
		return func(c *cli.Context) error {
			api, err := tmaint.Login(c, secret.ConsumerKey, secret.ConsumerSecret)
			if err != nil {
				return err
			}
			defer api.Close()
			return f(ctx, api, c.Args().Slice())
		}
	}

	action2 := func(f func(c StringFlag, api *anaconda.TwitterApi, args []string) error) func(*cli.Context) error {
		return func(c *cli.Context) error {
			api, err := tmaint.Login(c, secret.ConsumerKey, secret.ConsumerSecret)
			if err != nil {
				return err
			}
			defer api.Close()
			return f(c, api, c.Args().Slice())
		}
	}

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "editor",
				Usage: "editor to post",
			},
			&cli.StringFlag{
				Name:  "a",
				Usage: "configuration file path",
			},
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:   "followers",
				Usage:  "list members you are followed",
				Action: action1(followings),
			},
			&cli.Command{
				Name:   "followings",
				Usage:  "list members you follows",
				Action: action1(followings),
			},
			&cli.Command{
				Name:   "follow",
				Usage:  "follow people listed in STDIN\n\t  (Write like @ScreenName, ignore others)",
				Action: action1(follow),
			},
			&cli.Command{
				Name:   "unfollow",
				Usage:  "unfollow people listed in STDIN\n\t  (Write like @ScreenName, ignore others)",
				Action: action1(unfollow),
			},
			&cli.Command{
				Name:   "dump",
				Usage:  "IDNum ... dump JSON for the tweet",
				Action: action1(dump),
			},
			&cli.Command{
				Name:   "post",
				Usage:  "post tweet from STDIN",
				Action: action2(post),
			},
			&cli.Command{
				Name:   "cont",
				Usage:  "post continued tweet from STDIN",
				Action: action2(cont),
			},
			&cli.Command{
				Name:   "whoami",
				Usage:  "show who are you",
				Action: action1(whoami),
			},
			&cli.Command{
				Name:   "timeline",
				Usage:  "get home timeline",
				Action: action1(timeline),
			},
			&cli.Command{
				Name:   "mention",
				Usage:  "get mention timeline",
				Action: action1(mention),
			},
			&cli.Command{
				Name:   "reply",
				Usage:  "IDNum ... reply to IDNum",
				Action: action2(reply),
			},
			&cli.Command{
				Name:   "said",
				Usage:  "show what I said",
				Action: action1(said),
			},
			&cli.Command{
				Name:   "retweet",
				Usage:  "IDNum",
				Action: action1(retweet),
			},
			&cli.Command{
				Name:   "view",
				Usage:  "start viewer",
				Action: action2(view),
			},
		},
	}
	defer closer()
	return app.Run(args)
}

func main() {
	if err := mains(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
