package tmaint

import (
	"flag"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ChimeraCoder/anaconda"
)

type Setting struct {
	ScreenName        string
	AccessToken       string
	AccessTokenSecret string
	ConsumerKey       string
	ConsumerSecret    string
}

func FilePathChangeExtension(path, newext string) string {
	basename := path[:len(path)-len(filepath.Ext(path))]
	if newext[0] == '.' {
		return basename + newext
	} else {
		return basename + "." + newext
	}
}

var account = flag.String("a", "", "account json")

func GetSetting() (*Setting, error) {
	var cfgname string
	if *account != "" {
		cfgname = *account
	} else {
		exename, err := os.Executable()
		if err != nil {
			return nil, err
		}
		cfgname = FilePathChangeExtension(exename, ".json")
	}

	tokenText, err := ioutil.ReadFile(cfgname)
	if err != nil {
		return nil, err
	}

	var tk Setting
	err = json.Unmarshal(tokenText, &tk)
	return &tk, err
}

type Api = anaconda.TwitterApi

func Login() (*Api, *Setting, error) {
	tk, err := GetSetting()
	if err != nil {
		return nil, nil, err
	}
	return anaconda.NewTwitterApiWithCredentials(
			tk.AccessToken, tk.AccessTokenSecret,
			tk.ConsumerKey, tk.ConsumerSecret),
		tk, nil
}
