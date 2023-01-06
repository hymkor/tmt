package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mattn/go-isatty"

	tw "github.com/hymkor/tmt/oauth"
)

var byteOrderMark = "\xEF\xBB\xBF"

func post(c StringFlag, api *anaconda.TwitterApi, args []string) error {
	_, err := doPost(c, api, "", nil)
	return err
}

func makeDraft(text string) (string, error) {
	fname := filepath.Join(os.TempDir(), "post.txt")
	if err := ioutil.WriteFile(fname, []byte(byteOrderMark+text), 0600); err != nil {
		return "", err
	}
	return fname, nil
}

func callEditor(editor, fname string) ([]byte, error) {
	if editor == "" {
		editor = DEFAULT_EDITOR
	}
	cmd := exec.Command(editor, fname)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Run()
	if cmd.ProcessState == nil || !cmd.ProcessState.Success() {
		return nil, errors.New("canceled.")
	}
	text, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return bytes.Replace(text, []byte(byteOrderMark), []byte{}, -1), nil
}

func dumpTwitterError(err error, w io.Writer) {
	if apierr, ok := err.(*anaconda.ApiError); ok {
		for _, e := range apierr.Decoded.Errors {
			fmt.Fprintf(w, "%d: %s\n", e.Code, e.Message)
		}
	} else {
		//fmt.Fprintln(os.Stderr, reflect.TypeOf(err).String())
		fmt.Fprintln(w, err.Error())
	}
}

func _postWithEditor(api *tw.Api, editor string, draft string, values url.Values) (*anaconda.Tweet, error) {
	fname, err := makeDraft(draft)
	if err != nil {
		return nil, err
	}
	text, err := callEditor(editor, fname)
	if err != nil {
		return nil, err
	}
	for {
		textStr := string(text)
		if draft == textStr || strings.TrimSpace(textStr) == "" {
			return nil, errors.New("cancel post")
		}
		post, err := api.PostTweet(textStr, values)
		if err == nil {
			return &post, nil
		}
		dumpTwitterError(err, os.Stderr)
		fmt.Fprintln(os.Stderr, "Hit [Enter] to retry.")
		var dummy [100]byte
		os.Stdin.Read(dummy[:])
		text, err = callEditor(editor, fname)
		if err != nil {
			return nil, err
		}
	}
}

func doPost(flags StringFlag, api *tw.Api, draft string, values url.Values) (*anaconda.Tweet, error) {
	editor := flags.String("editor")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if isatty.IsTerminal(os.Stdin.Fd()) && editor != "" {
		return _postWithEditor(api, editor, draft, values)
	} else {
		text, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		post, err := api.PostTweet(string(text), values)
		return &post, err
	}
}

func myTimeline(api *anaconda.TwitterApi) ([]anaconda.Tweet, error) {
	u, err := api.GetSelf(nil)
	if err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Add("screen_name", u.ScreenName)

	return api.GetUserTimeline(values)
}

func cont(c StringFlag, api *anaconda.TwitterApi, args []string) error {
	timeline, err := myTimeline(api)
	if err != nil {
		return err
	}
	if len(timeline) <= 0 {
		return errors.New("too few timelins")
	}
	values := url.Values{}
	values.Add("in_reply_to_status_id", strconv.FormatInt(timeline[0].Id, 10))

	_, err = doPost(c, api, "", values)
	return err
}

var rxSuffixID = regexp.MustCompile(`\d+$`)

func reply(flags StringFlag, api *anaconda.TwitterApi, args []string) error {
	if len(args) <= 0 {
		return errors.New("required tweet ID")
	}
	m := rxSuffixID.FindString(args[0])
	if m == "" {
		return errors.New("required string contains tweet ID")
	}

	id, err := strconv.ParseInt(m, 10, 64)
	tweet, err := api.GetTweet(id, nil)
	if err != nil {
		return err
	}
	var draft strings.Builder
	var t *anaconda.Tweet
	if tweet.RetweetedStatus != nil {
		t = tweet.RetweetedStatus
	} else {
		t = &tweet
	}
	fmt.Fprintf(&draft, "@%s ", t.User.ScreenName)
	if t.InReplyToScreenName != "" {
		fmt.Fprintf(&draft, "@%s ", t.InReplyToScreenName)
	}
	values := url.Values{}
	values.Add("in_reply_to_status_id", m)
	_, err = doPost(flags, api, draft.String(), values)
	return err
}

func retweet(_ context.Context, api *anaconda.TwitterApi, args []string) error {
	for _, idStr := range args {
		m := rxSuffixID.FindString(idStr)
		if m == "" {
			return fmt.Errorf("%s: invalid ID", idStr)
		}
		id, err := strconv.ParseInt(m, 10, 64)
		if err != nil {
			return fmt.Errorf("%s: %s", idStr, err.Error())
		}
		api.Retweet(id, false)
	}
	return nil
}
