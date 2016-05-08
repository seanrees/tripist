package tasks

import (
	"fmt"
	"github.com/seanrees/tripist/todoist"
	"github.com/seanrees/tripist/tripit"
	"github.com/twinj/uuid"
	"log"
	"strconv"
	"time"
)

func PTR(s string) *string { return &s }

func UpdateTrip(api *todoist.SyncV6API, tt tripit.Trip, tasks []Task) error {
	name := fmt.Sprintf("Trip: %s", tt.DisplayName)

	var commands todoist.Commands
	var tempId string
	var itemsPresent []todoist.Item

	start, err := tt.Start()
	if err != nil {
		log.Printf("Could not get start date from trip: %v", err)
		return nil
	}

	p, err := findExistingProject(api, name)
	if err != nil {
		return err
	}

	if p == nil {
		tempId = uuid.NewV4().String()
		commands = append(commands, createProject(name, tempId))
	} else {
		// Gosh, I really hate the Todoist API. Parameters with differential
		// types can bite me.
		tempId = strconv.Itoa(*p.Id)
		itemsPresent, err = listItems(api, p)
		if err != nil {
			return err
		}
	}

	checklist := make(map[string]bool)
	present := make(map[string]bool)
	for _, t := range tasks {
		checklist[t.Template] = true
	}
	for _, i := range itemsPresent {
		present[*i.Content] = true
	}
	missing := diff(checklist, present)

	for _, t := range tasks {
		if v, _ := missing[t.Template]; !v {
			due := t.Due(start)
			commands = append(commands, createItem(tempId, t, due))
		}
	}

	for _, i := range itemsPresent {
		for _, t := range tasks {
			if t.Template == *i.Content {
				due := t.Due(start)
				i.DateString = PTR(due.Format(todoist.DateFormat))
				i.DueDateUTC = PTR(due.UTC().Format(todoist.DueDateFormat))
				commands = append(commands, updateItem(i))
				break
			}
		}
	}

	log.Printf("Queued %d commands to Todoist", len(commands))
	_, err = api.Write(commands)
	if err != nil {
		log.Printf("Write() failed: %v", err)
	}
	return err
}

func findExistingProject(api *todoist.SyncV6API, name string) (*todoist.Project, error) {
	resp, err := api.Read([]string{todoist.Projects}, 0)
	if err != nil {
		log.Printf("Could not read Todoist projects: %v", err)
		return nil, err
	}
	for _, p := range resp.Projects {
		if *p.Name == name {
			log.Printf("Found existing project %q id=%d", *p.Name, *p.Id)
			return &p, nil
		}
	}
	return nil, nil
}

func createProject(name, tempId string) todoist.WriteItem {
	return todoist.WriteItem{
		Type:   PTR(todoist.ProjectAdd),
		TempId: PTR(tempId),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   todoist.Project{Name: PTR(name)}}
}

func listItems(api *todoist.SyncV6API, p *todoist.Project) ([]todoist.Item, error) {
	resp, err := api.Read([]string{todoist.Items}, 0)
	if err != nil {
		log.Printf("Could not read Todoist items: %v", err)
		return nil, err
	}

	log.Printf("Loaded %d items from Todoist", len(resp.Items))

	var ret []todoist.Item
	for _, i := range resp.Items {
		if i.ProjectIdInt() == *p.Id {
			ret = append(ret, i)
		}
	}

	log.Printf("Found %d items for project %q (id=%d)", len(ret), *p.Name, *p.Id)
	return ret, nil
}

func createItem(projId string, t Task, due time.Time) todoist.WriteItem {
	log.Printf("Creating task %q due %s", t.Template, due.Format(todoist.DueDateFormat))

	return todoist.WriteItem{
		Type:   PTR(todoist.ItemAdd),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args: todoist.Item{
			Content:    PTR(t.Template),
			Indent:     &t.Indent,
			DateString: PTR(due.Format(todoist.DateFormat)),
			DueDateUTC: PTR(due.UTC().Format(todoist.DueDateFormat)),
			ProjectId:  PTR(projId)}}
}

func updateItem(i todoist.Item) todoist.WriteItem {
	return todoist.WriteItem{
		Type:   PTR(todoist.ItemUpdate),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   i}
}

func deleteItems(ids []int) todoist.WriteItem {
	log.Printf("Deleting %d items: %v", len(ids), ids)

	return todoist.WriteItem{
		Type: PTR(todoist.ItemDelete),
		UUID: PTR(uuid.NewV4().String()),
		Args: todoist.DeleteItems{Ids: ids}}
}

func diff(lhs, rhs map[string]bool) map[string]bool {
	ret := make(map[string]bool)
	for k, _ := range lhs {
		_, ok := rhs[k]
		ret[k] = ok
	}

	return ret
}
