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
	c.AdditionalAuthorizationUrlParams["oauth_callback"] = "out-of-band"
	return c
}

func Authorize() {
	c := buildConsumer()
	c.Debug(true)

	requestToken, url, err := c.GetRequestTokenAndUrl("")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("1. Login to TripIt in your browser.")
	fmt.Println("2. After login, browse to: " + url)
	fmt.Println("3. Grant access & copy the verification code.")
	fmt.Print("\nEnter verification code: ")
	verifyCode := ""
	fmt.Scanln(&verifyCode)

	accessToken, err := c.AuthorizeToken(requestToken, verifyCode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Access token: " + accessToken.Token)
	fmt.Println("Secret:       " + accessToken.Secret)
	fmt.Println("Additional data:")
	for k, v := range accessToken.AdditionalData {
		fmt.Printf("   %s = %s", k, v)
	}
	if len(accessToken.AdditionalData) == 0 {
		fmt.Println("  none")
	}
}
