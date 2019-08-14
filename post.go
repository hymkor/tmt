package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	// "reflect"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mattn/go-isatty"

	tw "github.com/zetamatta/tmt/oauth"
)

var ByteOrderMark = []byte{0xEF, 0xBB, 0xBF}

func post(ctx context.Context, api *anaconda.TwitterApi, args []string) error {
	return postWithValue(api, nil)
}

var flagEditor = flag.String("editor", "", "editor to use")

func makeDraft() (string, error) {
	fname := filepath.Join(os.TempDir(), "post.txt")
	if err := ioutil.WriteFile(fname, ByteOrderMark, 0600); err != nil {
		return "", err
	}
	return fname, nil
}

func callEditor(editor, fname string) ([]byte, error) {
	cmd := exec.Command(editor, fname)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Run()
	if !cmd.ProcessState.Success() {
		return nil, errors.New("canceled.")
	}
	text, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return bytes.Replace(text, ByteOrderMark, []byte{}, -1), nil
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

func postWithValue(api *tw.Api, values url.Values) error {
	editor := *flagEditor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if isatty.IsTerminal(os.Stdin.Fd()) && editor != "" {
		fname, err := makeDraft()
		if err != nil {
			return err
		}
		text, err := callEditor(editor, fname)
		if err != nil {
			return err
		}
		for {
			_, err := api.PostTweet(string(text), values)
			if err == nil {
				return nil
			}
			dumpTwitterError(err, os.Stderr)
			fmt.Fprintln(os.Stderr, "Hit [Enter] to retry.")
			var dummy [100]byte
			os.Stdin.Read(dummy[:])
			text, err = callEditor(editor, fname)
			if err != nil {
				return err
			}
		}
	} else {
		text, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		_, err = api.PostTweet(string(text), values)
		return err
	}
}

func cont(ctx context.Context, api *anaconda.TwitterApi, args []string) error {
	u, err := api.GetSelf(nil)
	if err != nil {
		return err
	}
	values := url.Values{}
	values.Add("screen_name", u.ScreenName)

	timeline, err := api.GetUserTimeline(values)
	if err != nil {
		return err
	}

	if len(timeline) <= 0 {
		return errors.New("too few timelins")
	}

	values = url.Values{}
	values.Add("in_reply_to_status_id", strconv.FormatInt(timeline[0].Id, 10))

	return postWithValue(api, values)
}

var rxSuffixID = regexp.MustCompile(`\d+$`)

func reply(ctx context.Context, api *anaconda.TwitterApi, args []string) error {
	if len(args) <= 0 {
		return errors.New("required tweet ID")
	}
	m := rxSuffixID.FindString(args[0])
	if m == "" {
		return errors.New("required string contains tweet ID")
	}
	values := url.Values{}
	values.Add("in_reply_to_status_id", m)
	return postWithValue(api, values)
}
