package tripit

import (
	"fmt"
	"github.com/mrjones/oauth"
	"log"
)

func buildConsumer() *oauth.Consumer {
	c := oauth.NewConsumer(
		consumerKey, consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.tripit.com/oauth/request_token",
			AuthorizeTokenUrl: "https://www.tripit.com/oauth/authorize",
			AccessTokenUrl:    "https://api.tripit.com/oauth/access_token",
		})

	// Required by TripIt.
	c.AdditionalAuthorizationUrlParams["oauth_callback"] = "https://freyr.erifax.org/tripist/"
	return c
}

func Authorize() *oauth.AccessToken {
	c := buildConsumer()
	//c.Debug(true)

	requestToken, url, err := c.GetRequestTokenAndUrl("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("1. Login to TripIt in your browser.")
	fmt.Println("2. After login, browse to: " + url)
	fmt.Println("2. Grant access and copy the 'oauth_token' parameter displayed.")
	fmt.Print("\nEnter oauth_token: ")
	verifyCode := ""
	fmt.Scanln(&verifyCode)

	at, err := c.AuthorizeToken(requestToken, verifyCode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Access token: " + at.Token)
	fmt.Println("Secret:       " + at.Secret)
	fmt.Println("Additional data:")
	for k, v := range at.AdditionalData {
		fmt.Printf("   %s = %s", k, v)
	}
	if len(at.AdditionalData) == 0 {
		fmt.Println("  none")
	}

	return at
}
