// Binary tripist access your trips in TripIt and creates a Todoist project.
package main

import (
	"flag"
	"fmt"
	"github.com/mrjones/oauth"
	"github.com/seanrees/tripist/config"
	"github.com/seanrees/tripist/tasks"
	"github.com/seanrees/tripist/todoist"
	"github.com/seanrees/tripist/tripit"
	"golang.org/x/oauth2"
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

func tripitOAuthAccessToken(u config.UserKeys) *oauth.AccessToken {
	return &oauth.AccessToken{
		Token:  u.TripitToken,
		Secret: u.TripitSecret,
	}
}

func todoistOAuth2Token(u config.UserKeys) *oauth2.Token {
	return &oauth2.Token{AccessToken: u.TodoistToken}
}

func main() {
	const configFilename = "user.json"
	const tripistConfigFilename = "tripist.json"

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stderr)

	var conf config.UserKeys
	if _, err := os.Stat(configFilename); err == nil {
		conf, err = config.Read(configFilename)
		if err != nil {
			log.Fatalf("Unable to read configuration: %v\n", err)
		}
	}

	if err := config.LoadAPIKeys(tripistConfigFilename); err != nil {
		log.Println("Using built-in keys.")
	} else {
		log.Printf("Loaded API keys from %s", tripistConfigFilename)
	}

	if *authorizeTripit {
		at := tripit.Authorize()
		conf.TripitToken = at.Token
		conf.TripitSecret = at.Secret
		config.Write(conf, configFilename)
		return
	}

	if *authorizeTodoist {
		t := todoist.Authorize()
		conf.TodoistToken = t.AccessToken
		config.Write(conf, configFilename)
		return
	}

	if *verifyTodoist {
		api := todoist.NewSyncV7API(todoistOAuth2Token(conf))
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

func listTrips(uc config.UserKeys) []tripit.Trip {
	api := tripit.NewTripitV1API(tripitOAuthAccessToken(uc))
	trips, err := api.List(&tripit.ListParameters{Traveler: "true", IncludeObjects: true})
	if err != nil {
		log.Printf("Could not list trips: %v", err)
	}

	return trips
}

func createProject(uc config.UserKeys, trip tripit.Trip, cl []tasks.ChecklistItem, taskCutoff time.Time) {
	// Fill this in from todoist.Authorize().
	todoapi := todoist.NewSyncV7API(todoistOAuth2Token(uc))

	name := fmt.Sprintf("Trip: %s", trip.DisplayName)
	log.Printf("Processing %s", name)

	p := tasks.Project{
		Name:  name,
		Tasks: tasks.Expand(cl, trip.ActualStartDate, trip.ActualEndDate, time.Now(), taskCutoff)}

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
