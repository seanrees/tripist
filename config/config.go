// Package config includes config support libraries for the programme.
package config

import (
	"encoding/json"
	"fmt"
	"github.com/seanrees/tripist/todoist"
	"github.com/seanrees/tripist/tripit"
	"io/ioutil"
	"os"
)

type UserKeys struct {
	// tripitToken is the user's TripIt oAuth token (oauth.AccessToken.Token).
	TripitToken string

	// tripitSecret is the user's TripIt oAuth secret (oauth.AccessToken.Secret).
	TripitSecret string

	// todoistToken is the user's Todoist oauth2 AccessToken (oauth2.Token.AccessToken).
	TodoistToken string
}

type apiKeys struct {
	TripitAPIKey        string
	TripitAPISecret     string
	TodoistClientID     string
	TodoistClientSecret string
}

func Read(filename string) (UserKeys, error) {
	uc := UserKeys{}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return uc, err
	}

	err = json.Unmarshal(b, &uc)
	return uc, err
}

func Write(uc UserKeys, filename string) error {
	b, err := json.MarshalIndent(uc, "", "\t")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	defer f.Close()
	if err != nil {
		return err
	}
	n, err := f.Write(b)
	if err != nil {
		return err
	}
	if l := len(b); l != n {
		return fmt.Errorf("short write to %s, wrote %d of %d bytes", filename, n, l)
	}
	return nil
}

func LoadAPIKeys(filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	tc := &apiKeys{}
	err = json.Unmarshal(b, &tc)
	if err != nil {
		return err
	}

	// Copy in parameters.
	tripit.ConsumerKey = tc.TripitAPIKey
	tripit.ConsumerSecret = tc.TripitAPISecret
	todoist.Oauth2ClientID = tc.TodoistClientID
	todoist.Oauth2ClientSecret = tc.TodoistClientSecret

	return nil
}
