package todoist

import (
	"fmt"
	"golang.org/x/oauth2"
	"log"
)

func buildConfig() *oauth2.Config {
	// todoist.com requires ClientID and ClientSecret to be set as parameters
	// in the POST.
	oauth2.RegisterBrokenAuthHeaderProvider("https://todoist.com")

	return &oauth2.Config{
		ClientID:     oauth2ClientID,
		ClientSecret: oauth2ClientSecret,
		Scopes:       []string{"data:read_write,data:delete,project:delete"},
		// TODO(seanrees): point this at something other than google.com since it
		// will eat the code= param if redirected to the country-specific site.
		RedirectURL: "http://www.google.com",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://todoist.com/oauth/authorize",
			TokenURL: "https://todoist.com/oauth/access_token",
		},
	}
}

func Authorize() *oauth2.Token {
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

	return token
}
