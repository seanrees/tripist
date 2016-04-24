// Binary tripist access your trips in TripIt and creates a Todoist project.
package main

import (
	"fmt"
	"github.com/mrjones/oauth"
	"github.com/seanrees/tripist/todoist"
	"github.com/seanrees/tripist/tripit"
	"github.com/twinj/uuid"
	"golang.org/x/oauth2"
	"time"
)

func strptr(s string) *string { return &s }

func main() {
	// Authorization bits.
	//tripit.Authorize()
	//token := todoist.Authorize()

	trips := listTrips()
	for _, t := range trips {
		//fmt.Printf("%#v\n", t)
		start, err := time.Parse(time.RFC3339, t.StartDate+"T00:00:00Z")
		if err != nil {
			fmt.Printf("Could not parse %s: %v", t.StartDate, err)
		}

		window := time.Now().AddDate(0, 0, 45)
		if start.After(window) {
			continue
		}

		fmt.Printf("Trip within window: %s\n", t.DisplayName)

		createProject(t.DisplayName)
	}
}

func listTrips() []tripit.Trip {
	at := &oauth.AccessToken{
        // Fill this in from tripit.Authorize() results.
		Token:  "",
		Secret: "",
	}
	api := tripit.NewTripitV1(at)
	trips, err := api.List(&tripit.ListParameters{Traveler: "true"})
	if err != nil {
		fmt.Printf("Could not list trips: %v", err)
	}

	return trips
}

func createProject(tripName string) {
    // Fill this in from todoist.Authorize().
	t := &oauth2.Token{AccessToken: ""}
	s := todoist.NewSyncV6API(t)

	name := "Trip: " + tripName
	items := []string{
		"Pack clothes", // TODO: add how many days, wx suggestions?
		"Pack electronics (Chromecast, chargers)",
		"Pack toiletries",
		"Clean litter box",
		"Feed cats",
		"Set heater appropriately",
		"Pack essential documents",
	}

	// Check for pre-existing projects.
	resp, err := s.Read([]string{todoist.Projects}, 0)
	if err != nil {
		fmt.Printf("Could not read Todoist projects: %v", err)
	}
	for _, p := range resp.Projects {
		if *p.Name == name {
			fmt.Printf("Project already exists, no work needed.\n")
			return
		}
	}

	fmt.Printf("Creating project: %q\n", name)

	// Create a new project.
	tempId := uuid.NewV4().String()
	commands := todoist.Commands{todoist.WriteItem{
		Type:   strptr(todoist.ProjectAdd),
		TempId: strptr(tempId),
		UUID:   strptr(uuid.NewV4().String()),
		Args:   todoist.Project{Name: strptr(name)}}}

	for _, i := range items {
		commands = append(commands, todoist.WriteItem{
			Type:   strptr(todoist.ItemAdd),
			TempId: strptr(uuid.NewV4().String()),
			UUID:   strptr(uuid.NewV4().String()),
			Args: todoist.Item{
				Content:   strptr(i),
				ProjectId: strptr(tempId)}})
	}

	_, err = s.Write(commands)
	if err != nil {
		fmt.Printf("Could not create project: %v\n", err)
	}
}
