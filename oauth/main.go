package tmaint

import (
	"encoding/json"
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

func ConfigurationPath() (string, error) {
	exename, err := os.Executable()
	if err != nil {
		return "", err
	}
	return FilePathChangeExtension(exename, ".json"), nil
}

type Flags interface {
	String(name string) string
}

func getAccess(flags Flags, consumerKey, consumerSecret string) (*accessT, error) {
	cfgname := flags.String("a")
	if cfgname == "" {
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

func Login(flags Flags, consumerKey, consumerSecret string) (*Api, error) {
	access, err := getAccess(flags, consumerKey, consumerSecret)
	if err != nil {
		return nil, err
	}
	return anaconda.NewTwitterApiWithCredentials(
		access.AccessToken,
		access.AccessTokenSecret,
		consumerKey,
		consumerSecret), nil
}
