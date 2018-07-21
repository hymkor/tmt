package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

var from = flag.String("f", "", "ID that seek tweet from")
var session = flag.String("s", "", "Filename to keep session")

type Setting struct {
	ScreenName        string
	AccessToken       string
	AccessTokenSecret string
	ConsumerKey       string
	ConsumerSecret    string
}

func GetSetting() (*Setting, error) {
	exename, err := os.Executable()
	if err != nil {
		return nil, err
	}

	cfgname := exename[:len(exename)-len(filepath.Ext(exename))] + ".json"

	tokenText, err := ioutil.ReadFile(cfgname)
	if err != nil {
		return nil, err
	}

	var tk Setting
	err = json.Unmarshal(tokenText, &tk)
	return &tk, err
}

func main1(args []string) error {
	tk, err := GetSetting()
	if err != nil {
		return err
	}

	api := anaconda.NewTwitterApiWithCredentials(
		tk.AccessToken, tk.AccessTokenSecret,
		tk.ConsumerKey, tk.ConsumerSecret)
	defer api.Close()

	v := url.Values{"screen_name": {tk.ScreenName}}

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
