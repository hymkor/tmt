package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	tmaint "github.com/hymkor/tmt/oauth"
	"github.com/hymkor/tmt/secret"
)

var from = flag.String("f", "", "ID that seek tweet from")
var session = flag.String("s", "", "Filename to keep session")

func main1(args []string) error {
	api, err := tmaint.Login(secret.ConsumerKey, secret.ConsumerSecret)
	if err != nil {
		return err
	}
	defer api.Close()

	v := url.Values{}

	if *session != "" {
		data, err := ioutil.ReadFile(*session)
		if err == nil {
			v["max_id"] = []string{strings.TrimSpace(string(data))}
		}
	}

	if *from != "" {
		v["max_id"] = []string{*from}
	}

	result, err := api.GetUserTimeline(v)
	if err != nil {
		return err
	}
	var lastId int64
	for _, t := range result {
		if t.Retweeted {
			fmt.Println("-------")
			fmt.Println(t.CreatedAt)
			fmt.Println(t.FullText)
			api.DeleteTweet(t.Id, false)
		}
		lastId = t.Id
	}
	fmt.Printf("\nlastID=%d\n", lastId)
	if *session != "" {
		return ioutil.WriteFile(*session, []byte(fmt.Sprint(lastId)), 0666)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := main1(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
