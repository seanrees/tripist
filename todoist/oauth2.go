package todoist

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"log"
)

func buildConfig() *oauth2.Config {
	// todoist.com requires ClientID and ClientSecret to be set as parameters
	// in the POST.
	oauth2.RegisterBrokenAuthHeaderProvider("https://todoist.com")

	return &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{"data:read_write,data:delete,project:delete"},
		RedirectURL:  "http://www.google.com",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://todoist.com/oauth/authorize",
			TokenURL: "https://todoist.com/oauth/access_token",
		},
	}
}

func Authorize() []byte {
	conf := buildConfig()

	// state=txs -- totally a random string.
	url := conf.AuthCodeURL("txs", oauth2.AccessTypeOffline)

	fmt.Println("1. Browse to: " + url)
	fmt.Println("2. Copy the code= parameter from your URL bar.")
	fmt.Print("\nEnter verification code: ")
	code := ""
	fmt.Scanln(&code)

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}
