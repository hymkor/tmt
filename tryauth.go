package tmaint

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/garyburd/go-oauth/oauth"
)

const outOfBound = "oob"

func PinOAuth(consumerKey string, consumerSecret string, pinInput func(url string) string) (accessToken string, accessTecret string, err error) {
	oauthClient := oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  consumerKey,
			Secret: consumerSecret,
		},
	}

	tempCred, err := oauthClient.RequestTemporaryCredentials(nil, outOfBound, nil)
	if err != nil {
		return "", "", err
	}

	url := oauthClient.AuthorizationURL(tempCred, nil)
	pin := pinInput(url)
	if pin == "" {
		return "", "", errors.New("pin input cancel")
	}
	tokenCred, _, err := oauthClient.RequestToken(nil, tempCred, pin)
	return tokenCred.Token, tokenCred.Secret, err
}

func openUrl(url string) {
	cmd := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url)
	cmd.Run()
}

func url2pin(url string) (pin string) {
	openUrl(url)
	fmt.Print("PIN:")
	fmt.Scan(&pin)
	return
}
