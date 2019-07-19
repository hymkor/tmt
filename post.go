package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/mattn/go-isatty"

	tw "github.com/zetamatta/tmt/oauth"
)

var ByteOrderMark = []byte{0xEF, 0xBB, 0xBF}

func post(ctx context.Context, api *tw.Api, args []string) error {
	return postWithValue(ctx, api, nil)
}

var flagEditor = flag.String("editor", "", "editor to use")

func postWithValue(ctx context.Context, api *tw.Api, values url.Values) error {
	var text []byte
	editor := *flagEditor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if isatty.IsTerminal(os.Stdin.Fd()) && editor != "" {
		fname := filepath.Join(os.TempDir(), "post.txt")
		if err := ioutil.WriteFile(fname, ByteOrderMark, 066); err != nil {
			return err
		}
		cmd := exec.Command(editor, fname)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Run()
		if !cmd.ProcessState.Success() {
			return errors.New("canceled.")
		}
		var buffer bytes.Buffer
		fd, err := os.Open(fname)
		if err != nil {
			return err
		}
		defer fd.Close()
		sc := bufio.NewScanner(fd)
		for sc.Scan() {
			line := sc.Bytes()
			line = bytes.Replace(line, ByteOrderMark, []byte{}, -1)
			buffer.Write(line)
		}
		if err := sc.Err(); err != nil {
			return err
		}
		text = buffer.Bytes()
	} else {
		var err error
		text, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	}
	_, err := api.PostTweet(string(text), values)
	return err
}

func cont(ctx context.Context, api *tw.Api, args []string) error {
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

	return postWithValue(ctx, api, values)
}
