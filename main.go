// Binary tripist access your trips in TripIt and creates a Todoist project.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mrjones/oauth"
	"github.com/seanrees/tripist/tasks"
	"github.com/seanrees/tripist/todoist"
	"github.com/seanrees/tripist/tripit"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	authorizeTripit  = flag.Bool("authorize_tripit", false, "Perform Tripit Authorization. This is an exclusive flag.")
	authorizeTodoist = flag.Bool("authorize_todoist", false, "Perform Todoist Authorization. This is an exclusive flag.")
	travelWindowDays = flag.Int("travel_window_days", 45, "Number of days ahead to look for trips.")
	checklistCSV     = flag.String("checklist_csv", "checklist.csv", "Travel checklist CSV file.")
)

type userConfig struct {
	// tripitToken is the user's TripIt oAuth token (oauth.AccessToken.Token).
	TripitToken string

	// tripitSecret is the user's TripIt oAuth secret (oauth.AccessToken.Secret).
	TripitSecret string

	// todoistToken is the user's Todoist oauth2 AccessToken (oauth2.Token.AccessToken).
	TodoistToken string
}

func (u *userConfig) TripitOAuthAccessToken() *oauth.AccessToken {
	return &oauth.AccessToken{
		Token:  u.TripitToken,
		Secret: u.TripitSecret,
	}
}

func (u *userConfig) TodoistOAuth2Token() *oauth2.Token {
	return &oauth2.Token{AccessToken: u.TodoistToken}
}

func main() {
	const configFilename = "user.json"

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stderr)

	conf := &userConfig{}
	if _, err := os.Stat(configFilename); err == nil {
		conf, err = readConfig(configFilename)
		if err != nil {
			log.Fatalf("Unable to read configuration: %v\n", err)
		}
	}

	if *authorizeTripit {
		at := tripit.Authorize()
		conf.TripitToken = at.Token
		conf.TripitSecret = at.Secret
		writeConfig(conf, configFilename)
		return
	}

	if *authorizeTodoist {
		t := todoist.Authorize()
		conf.TodoistToken = t.AccessToken
		writeConfig(conf, configFilename)
		return
	}

	checklist, err := tasks.Load(*checklistCSV)
	if err != nil {
		log.Fatalf("Unable to load travel checklist (%s): %v", *checklistCSV, err)
	}

	log.Printf("Loaded %s with %d tasks\n", *checklistCSV, len(checklist))

	trips := listTrips(conf)
	for _, t := range trips {
		start, err := t.Start()
		if err != nil {
			fmt.Printf("Could not parse %s: %v", t.StartDate, err)
		}

		window := time.Now().AddDate(0, 0, *travelWindowDays)
		if start.After(window) {
			continue
		}

		log.Printf("Trip within window: %s\n", t.DisplayName)
		createProject(conf, t, checklist)
	}

}

func readConfig(filename string) (*userConfig, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	uc := &userConfig{}
	err = json.Unmarshal(b, &uc)
	return uc, err
}

func writeConfig(uc *userConfig, filename string) error {
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

func listTrips(uc *userConfig) []tripit.Trip {
	api := tripit.NewTripitV1API(uc.TripitOAuthAccessToken())
	trips, err := api.List(&tripit.ListParameters{Traveler: "true"})
	if err != nil {
		fmt.Printf("Could not list trips: %v", err)
	}

	return trips
}

func createProject(uc *userConfig, trip tripit.Trip, checklist []tasks.ChecklistItem) {
	// Fill this in from todoist.Authorize().
	s := todoist.NewSyncV6API(uc.TodoistOAuth2Token())

	tasks.UpdateTrip(s, trip, checklist)
}
