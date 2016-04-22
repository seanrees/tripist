package main

import (
	"fmt"
	"github.com/mrjones/oauth"
	"github.com/seanrees/tripist/tripit"
)

func main() {
  // TODO(seanrees): wire in some functionality ot call tripit.Authorize and
  // store results.
	// TODO(seanrees): pull in a configuration file.
	at := &oauth.AccessToken{
		Token:  "",
		Secret: "",
	}

	api := tripit.NewTripitV1(at)
	trips, err := api.List(&tripit.ListParameters{Traveler: "true"})
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	fmt.Printf("%v\n", trips)
}
