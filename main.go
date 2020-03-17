package main

import (
	"context"
	"fmt"
	"os"

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

	api, err := tmaint.Login(secret.ConsumerKey, secret.ConsumerSecret)
	if err != nil {
		return err
	}
	defer api.Close()

	ctx, closer := ctrlc.Setup(context.Background())

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "editor",
				Usage: "editor to post",
			},
		},
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "followers",
				Usage: "list members you are followed",
				Action: func(c *cli.Context) error {
					return followings(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "followings",
				Usage: "list members you follows",
				Action: func(c *cli.Context) error {
					return followings(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "follow",
				Usage: "follow people listed in STDIN\n\t  (Write like @ScreenName, ignore others)",
				Action: func(c *cli.Context) error {
					return follow(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "unfollow",
				Usage: "unfollow people listed in STDIN\n\t  (Write like @ScreenName, ignore others)",
				Action: func(c *cli.Context) error {
					return unfollow(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "dump",
				Usage: "IDNum ... dump JSON for the tweet",
				Action: func(c *cli.Context) error {
					return dump(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "post",
				Usage: "post tweet from STDIN",
				Action: func(c *cli.Context) error {
					return post(c, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "cont",
				Usage: "post continued tweet from STDIN",
				Action: func(c *cli.Context) error {
					return cont(c, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "whoami",
				Usage: "show who are you",
				Action: func(c *cli.Context) error {
					return whoami(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "timeline",
				Usage: "get home timeline",
				Action: func(c *cli.Context) error {
					return timeline(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "mention",
				Usage: "get mention timeline",
				Action: func(c *cli.Context) error {
					return mention(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "reply",
				Usage: "IDNum ... reply to IDNum",
				Action: func(c *cli.Context) error {
					return reply(c, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "said",
				Usage: "show what I said",
				Action: func(c *cli.Context) error {
					return said(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "retweet",
				Usage: "IDNum",
				Action: func(c *cli.Context) error {
					return retweet(ctx, api, c.Args().Slice())
				},
			},
			&cli.Command{
				Name:  "view",
				Usage: ".. start viewer",
				Action: func(c *cli.Context) error {
					return view(c, api, c.Args().Slice())
				},
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
