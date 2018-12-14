package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	tw "github.com/zetamatta/tmt/oauth"
)

var ByteOrderMark = []byte{0xEF, 0xBB, 0xBF}

func post(ctx context.Context, api *tw.Api, args []string) error {
	var text []byte
	editor := os.Getenv("EDITOR")
	if editor != "" {
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
	_, err := api.PostTweet(string(text), nil)
	return err
}
