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
	taskCutoffDays   = flag.Int("task_cutoff_days", 7, "Create tasks upto this many days in advance of their due date.")
	checklistCSV     = flag.String("checklist_csv", "checklist.csv", "Travel checklist CSV file.")
	verifyTodoist    = flag.Bool("verify_todoist", false, "Perform Todoist API validation. This is an exclusive flag.")
)

type userConfig struct {
	// tripitToken is the user's TripIt oAuth token (oauth.AccessToken.Token).
	TripitToken string

	// tripitSecret is the user's TripIt oAuth secret (oauth.AccessToken.Secret).
	TripitSecret string

	// todoistToken is the user's Todoist oauth2 AccessToken (oauth2.Token.AccessToken).
	TodoistToken string
}

type tripistConfig struct {
	TripitAPIKey        string
	TripitAPISecret     string
	TodoistClientID     string
	TodoistClientSecret string
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
	const tripistConfigFilename = "tripist.json"

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

	if err := readTripistConfig(tripistConfigFilename); err != nil {
		log.Println("Using built-in keys.")
	} else {
		log.Printf("Loaded API keys from %s", tripistConfigFilename)
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

	if *verifyTodoist {
		api := todoist.NewSyncV7API(conf.TodoistOAuth2Token())
		if err := todoist.Verify(api); err != nil {
			log.Printf("Todoist validation failed: %v", err)
		} else {
			log.Printf("Todoist validation success.")
		}
		return
	}

	checklist, err := tasks.Load(*checklistCSV)
	if err != nil {
		log.Fatalf("Unable to load travel checklist (%s): %v", *checklistCSV, err)
	}

	log.Printf("Loaded %s with %d tasks\n", *checklistCSV, len(checklist))

	trips := listTrips(conf)
	window := time.Now().AddDate(0, 0, *taskCutoffDays)

	log.Printf("Creating tasks up to cutoff %s", window)

	for _, t := range trips {
		createProject(conf, t, checklist, window)
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

func readTripistConfig(filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	tc := &tripistConfig{}
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
		log.Printf("Could not list trips: %v", err)
	}

	return trips
}

func createProject(uc *userConfig, trip tripit.Trip, cl []tasks.ChecklistItem, taskCutoff time.Time) {
	// Fill this in from todoist.Authorize().
	todoapi := todoist.NewSyncV7API(uc.TodoistOAuth2Token())

	start, err := trip.Start()
	if err != nil {
		log.Printf("Unable to get start date from trip: %v\n", err)
		return
	}

	end, err := trip.End()
	if err != nil {
		log.Printf("Unable to get end date from trip: %v\n", err)
		return
	}

	name := fmt.Sprintf("Trip: %s", trip.DisplayName)
	log.Printf("Processing %s", name)

	p := tasks.Project{
		Name:  name,
		Tasks: tasks.Expand(cl, start, end, time.Now(), taskCutoff)}

	if p.Empty() {
		log.Println("No tasks within cutoff window, skipping.")
		return
	}

	rp, found, err := todoapi.LoadProject(name)
	if err != nil {
		log.Printf("Could not load remote project: %v", err)
	}
	if found {
		diffs := rp.DiffTasks(p)

		err = todoapi.UpdateProject(rp, diffs)
		if err != nil {
			log.Printf("Unable to update project: %v", err)
		}
	} else {
		err = todoapi.CreateProject(p)
		if err != nil {
			log.Printf("Unable to create project: %v", err)
		}
	}
}
