package todoist

import (
	"encoding/json"
	"fmt"
	"github.com/seanrees/tripist/tasks"
	"github.com/twinj/uuid"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"time"
)

const (
	ApiPath = "https://todoist.com/API/v7/sync"
)

func PTR(s string) *string { return &s }

type SyncV6API struct {
	token *oauth2.Token
}

func NewSyncV6API(t *oauth2.Token) *SyncV6API {
	return &SyncV6API{token: t}
}

func (s *SyncV6API) makeRequest(path string, data url.Values, obj interface{}) error {
	c := buildConfig().Client(oauth2.NoContext, s.token)
	resp, err := c.PostForm(path, data)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	j, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(j, &obj); err != nil {
		return err
	}

	return nil
}

// Reads specific types and returns a ReadResponse. Possible types are in constants:
// Projects, Items.
//
// sequenceNumber should be zero or a ReadResponse.SequenceNumber if you desire an
// incremental read.
func (s *SyncV6API) Read(types []string, sequenceNumber int) (ReadResponse, error) {
	resp := ReadResponse{}
	params := url.Values{}
	params.Add("token", s.token.AccessToken)
	params.Add("seq_no", fmt.Sprintf("%d", sequenceNumber))

	t, err := json.Marshal(types)
	if err != nil {
		return resp, nil
	}
	params.Add("resource_types", string(t))

	if err := s.makeRequest(ApiPath, params, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *SyncV6API) Write(c Commands) (WriteResponse, error) {
	resp := WriteResponse{}
	params := url.Values{}
	params.Add("token", s.token.AccessToken)

	cmds, err := json.Marshal(c)
	if err != nil {
		return resp, nil
	}
	params.Add("commands", string(cmds))

	if err := s.makeRequest(ApiPath, params, &resp); err != nil {
		return resp, err
	}

	log.Printf("Response from Todoist: %v", resp)

	uuidToCmd := make(map[string]*WriteItem)
	errors := 0
	for _, cmd := range c {
		uuidToCmd[*cmd.UUID] = &cmd
	}

	for uuid, state := range resp.SyncStatus {
		code, ok := state.(string)
		if ok {
			if code != "ok" {
				log.Printf("Unexpected error %q from Todoist for command UUID %s", state, uuid)
				errors++
			}
			continue
		}

		e, ok := state.(map[string]interface{})
		if ok {
			cmd, ok := uuidToCmd[uuid]
			if ok {
				log.Printf("Error syncing %s (UUID %s): %v (%v)", *cmd.Type, uuid, e["error"], cmd.Args)
			} else {
				log.Printf("Error for unsent UUID %s: %v", uuid, e)
			}
			errors++
			continue
		}

		log.Printf("Unknown response type from Todoist: %T", state)
	}

	if errors > 0 {
		return resp, fmt.Errorf("%d sync errors writing to Todoist", errors)
	}

	return resp, nil
}

func (s *SyncV6API) listItems(p *Project) ([]Item, error) {
	resp, err := s.Read([]string{Items}, 0)
	if err != nil {
		log.Printf("Could not read Todoist items: %v", err)
		return nil, err
	}

	var ret []Item
	for _, i := range resp.Items {
		if i.ProjectIdInt() == *p.Id {
			ret = append(ret, i)
		}
	}

	log.Printf("Loaded %d items (%d for this project) from Todoist", len(resp.Items), len(ret))

	return ret, nil
}

func (s *SyncV6API) findProject(name string) (*Project, error) {
	resp, err := s.Read([]string{Projects}, 0)
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

func (s *SyncV6API) createProject(name, tempId string) WriteItem {
	return WriteItem{
		Type:   PTR(ProjectAdd),
		TempId: PTR(tempId),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   Project{Name: PTR(name)}}
}

func (s *SyncV6API) createItem(projId string, t tasks.Task) WriteItem {
	log.Printf("Creating task %q due %s", t.Content, t.DueDate.Format(DueDateFormatForWrite))

	return WriteItem{
		Type:   PTR(ItemAdd),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args: Item{
			Content:    &t.Content,
			Indent:     &t.Indent,
			DateString: PTR(t.DueDate.Format(DateFormat)),
			DueDateUTC: PTR(t.DueDate.Format(DueDateFormatForWrite)),
			ProjectId:  &projId}}
}

func (s *SyncV6API) updateItem(i Item) WriteItem {
	return WriteItem{
		Type:   PTR(ItemUpdate),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   i}
}

// Returns a tasks.Project, whether or not it was found, and any error.
func (s *SyncV6API) LoadProject(name string) (tasks.Project, bool, error) {
	ret := tasks.Project{Name: name}
	found := false

	p, err := s.findProject(name)
	switch {
	case err != nil:
		return ret, found, err
	case p == nil:
		return ret, found, nil
	}
	found = true

	li, err := s.listItems(p)
	if err != nil {
		return ret, found, err
	}

	for _, i := range li {
		due, err := time.ParseInLocation(DueDateFormatForRead, *i.DueDateUTC, time.UTC)
		if err != nil {
			log.Printf("Could not parse %q: %v (ignoring, may generate diffs)", *i.DueDateUTC, err)
		}

		ret.Tasks = append(ret.Tasks, tasks.Task{
			Content: *i.Content,
			DueDate: due,
			Indent:  *i.Indent})
	}

	ret.External = &projectItems{ProjectId: *p.Id, Items: li}

	return ret, found, nil
}

func (s *SyncV6API) CreateProject(p tasks.Project) error {
	tempId := uuid.NewV4().String()

	cmds := Commands{s.createProject(p.Name, tempId)}
	for _, t := range p.Tasks {
		cmds = append(cmds, s.createItem(tempId, t))
	}

	_, err := s.Write(cmds)
	return err
}

func (s *SyncV6API) UpdateProject(p tasks.Project, diffs []tasks.Diff) error {
	tp, ok := p.External.(*projectItems)
	if !ok {
		return fmt.Errorf("missing or invalid external project pointer on %q", p.Name)
	}

	// Sigh. Todoist's API is inconsistent here; in Create, we make a string but get
	// back an int. So we need to make it a string again.
	projectId := strconv.Itoa(tp.ProjectId)

	var cmds Commands

	for _, d := range diffs {
		switch d.Type {
		case tasks.Added:
			cmds = append(cmds, s.createItem(projectId, d.Task))

		case tasks.Changed:
			for _, i := range tp.Items {
				if *i.Content == d.Task.Content {
					i.DateString = PTR(d.Task.DueDate.Format(DateFormat))
					i.DueDateUTC = PTR(d.Task.DueDate.Format(DueDateFormatForWrite))
					i.Indent = &d.Task.Indent
					cmds = append(cmds, s.updateItem(i))
					log.Printf("New due date = %s", *i.DueDateUTC)
					break
				}
			}

		case tasks.Removed:
			log.Printf("Not removing missing task: %q", d.Task.Content)
		}
	}

	if len(cmds) > 0 {
		_, err := s.Write(cmds)
		return err
	} else {
		log.Printf("No commands to run to update project %q", p.Name)
	}

	return nil
}
