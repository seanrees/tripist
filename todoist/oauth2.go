package todoist

import (
	"fmt"
	"golang.org/x/oauth2"
	"log"
)

func buildConfig() *oauth2.Config {
	// todoist.com requires ClientID and ClientSecret to be set as parameters
	// in the POST.
	//oauth2.RegisterBrokenAuthHeaderProvider("https://todoist.com")

	return &oauth2.Config{
		ClientID:     Oauth2ClientID,
		ClientSecret: Oauth2ClientSecret,
		Scopes:       []string{"data:read_write,data:delete,project:delete"},
		RedirectURL:  "https://freyr.erifax.org/tripist/",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://todoist.com/oauth/authorize",
			TokenURL: "https://todoist.com/oauth/access_token",
		},
	}
}

func Authorize() *oauth2.Token {
	conf := buildConfig()

	// state=erifax -- totally a random string.
	url := conf.AuthCodeURL("erifax", oauth2.AccessTypeOffline)

	fmt.Println("1. Browse to: " + url)
	fmt.Println("2. Grant access and copy the 'code' parameter displayed.")
	fmt.Print("\nEnter code: ")
	code := ""
	fmt.Scanln(&code)

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}

	return token
}
