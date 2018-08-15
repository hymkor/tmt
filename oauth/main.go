package tmaint

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ChimeraCoder/anaconda"
)

type accessT struct {
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

func ConfigurationPath() (string, error) {
	exename, err := os.Executable()
	if err != nil {
		return "", err
	}
	return FilePathChangeExtension(exename, ".json"), nil
}

func getAccess(consumerKey, consumerSecret string) (*accessT, error) {
	var cfgname string
	if *account != "" {
		cfgname = *account
	} else {
		var err error
		cfgname, err = ConfigurationPath()
		if err != nil {
			return nil, err
		}
	}
	var access accessT
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

func Login(consumerKey, consumerSecret string) (*Api, error) {
	access, err := getAccess(consumerKey, consumerSecret)
	if err != nil {
		return nil, err
	}
	return anaconda.NewTwitterApiWithCredentials(
		access.AccessToken,
		access.AccessTokenSecret,
		consumerKey,
		consumerSecret), nil
}
