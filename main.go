package tmaint

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ChimeraCoder/anaconda"
)

type Access struct {
	AccessToken       string
	AccessTokenSecret string
}

type Setting struct {
	ScreenName string
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

func GetSetting() (*Access, error) {
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
	var access Access
	tokenText, err := ioutil.ReadFile(cfgname)
	if err != nil {
		access.AccessToken, access.AccessTokenSecret, err = PinOAuth(
			consumerKey,
			consumerSecret,
			url2pin)
		if err != nil {
			return nil, err
		}

		bin, err := json.Marshal(&access)
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(cfgname, bin, 0666)
		if err != nil {
			return nil, err
		}
	} else if err = json.Unmarshal(tokenText, &access); err != nil {
		return nil, err
	}
	return &access, nil
}

type Api = anaconda.TwitterApi

func Login() (*Api, *Setting, error) {
	access, err := GetSetting()
	if err != nil {
		return nil, nil, err
	}
	return anaconda.NewTwitterApiWithCredentials(
			access.AccessToken,
			access.AccessTokenSecret,
			consumerKey,
			consumerSecret),
		&Setting{}, nil
}
