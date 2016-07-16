package todoist

import (
	"fmt"
	"github.com/seanrees/tripist/tasks"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Verify runs a series of compatability tests on Todoist, returning error if incompatible.
//
// This code verifies that Todoist and this are compatible. It is done in lieu of
// a unit test or mock, as any local implementation would diverge from the ultimate
// source of truth.
//
// This code fails fast and should not be expected to return all compatibility errors in
// one call.
//
// This could be rewritten with Go's testing package but would require a TestMain to setup
// the API and user keys.
func Verify(api *SyncV7API) error {
	step := 0

	var err error
	var p *Project
	var tp *tasks.Project
	var items []Item

	name := randomProjectName()
	if _, err = verifyProjectPresence(name, &step, false, api); err != nil {
		return err
	}

	l(&step, "Creating project %q", name)
	err = api.CreateProject(tasks.Project{Name: name})
	if err != nil {
		return err
	}

	l(&step, "Verifying %q created successfully", name)
	p, err = verifyProjectPresence(name, &step, true, api)
	if err != nil {
		return err
	}

	l(&step, "Adding items to project %q", name)
	due := time.Date(2016, 07, 15, 12, 00, 00, 00, time.UTC)
	testTasks := []tasks.Task{
		{Content: "one", Position: 1, DueDate: due},
		{Content: "two", Position: 2, DueDate: due},
	}
	cmds := Commands{}
	id := strconv.Itoa(*p.Id)
	for _, t := range testTasks {
		cmds = append(cmds, api.createItem(id, t))
	}
	if _, err := api.Write(cmds); err != nil {
		return err
	}

	tp, err = verifyTasksInProject(name, &step, testTasks, api)

	l(&step, "Updating an item in project %q", name)
	testTasks[0].Position = 3
	testTasks[0].DueDate.Add(24 * time.Hour)
	d := tasks.Diff{tasks.Changed, testTasks[0]}
	err = api.UpdateProject(*tp, []tasks.Diff{d})
	if err != nil {
		return err
	}

	_, err = verifyTasksInProject(name, &step, testTasks, api)

	l(&step, "Deleting an item")
	items, err = api.listItems(p)
	del := items[0]
	if err != nil {
		return err
	}
	if _, err = api.Write(Commands{api.deleteItem(del)}); err != nil {
		return err
	}
	items, err = api.listItems(p)
	for _, i := range items {
		if *i.Content == *del.Content {
			return fmt.Errorf("delete item failed on item %q", *i.Content)
		}
	}

	l(&step, "Deleting project %q", name)
	if _, err = api.Write(Commands{api.deleteProject(p)}); err != nil {
		return err
	}

	if _, err = verifyProjectPresence(name, &step, false, api); err != nil {
		return err
	}

	return nil
}

func l(step *int, s string, v ...interface{}) {
	msg := fmt.Sprintf("[step %02d] %s", *step, s)
	log.Printf(msg, v...)
	*step++
}

func randomProjectName() string {
	chars := []rune("abcdefABCDEF0123456789")
	name := make([]rune, 5)
	for i := range name {
		name[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("Todoist Verification %s", string(name))
}

func verifyProjectPresence(name string, step *int, expected bool, api *SyncV7API) (*Project, error) {
	str := ""
	if !expected {
		str = "not "
	}

	l(step, "Verifying %q %spresent", name, str)

	p, err := api.findProject(name)
	if err != nil {
		return nil, err
	}
	found := p != nil
	if found != expected {
		return nil, fmt.Errorf("found %q which should %sexist", name, str)
	}

	return p, nil
}

func verifyTasksInProject(name string, step *int, expected []tasks.Task, api *SyncV7API) (*tasks.Project, error) {
	l(step, "Verifying items in project %q", name)
	tp, found, err := api.LoadProject(name)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("unable to find project %q with LoadProject", name)
	}
	if d := tp.DiffTasks(tasks.Project{Tasks: expected}); len(d) > 0 {
		return nil, fmt.Errorf("unexpected diffs in write / read cycle: %v", d)
	}

	return &tp, nil
}
