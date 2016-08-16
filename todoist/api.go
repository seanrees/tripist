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

type SyncV7API struct {
	token *oauth2.Token
}

func NewSyncV7API(t *oauth2.Token) *SyncV7API {
	return &SyncV7API{token: t}
}

func (s *SyncV7API) makeRequest(path string, data url.Values, obj interface{}) error {
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
func (s *SyncV7API) Read(types []string, sequenceNumber int) (ReadResponse, error) {
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

func (s *SyncV7API) Write(c Commands) (WriteResponse, error) {
	resp := WriteResponse{}
	params := url.Values{}
	params.Add("token", s.token.AccessToken)

	cmds, err := json.Marshal(c)
	if err != nil {
		return resp, nil
	}
	params.Add("commands", string(cmds))

	log.Printf("Writing %d commands to Todoist", len(c))

	if err := s.makeRequest(ApiPath, params, &resp); err != nil {
		return resp, err
	}

	if errs := s.checkErrors(&c, &resp); len(errs) > 0 {
		for _, e := range errs {
			log.Printf("Write error from Todoist: %s", e.Message)
		}
		return resp, fmt.Errorf("write failed for %d/%d commands, call checkErrors", len(errs), len(c))
	}

	return resp, nil
}

func (s *SyncV7API) checkErrors(cmds *Commands, r *WriteResponse) []writeError {
	var ret []writeError
	uuidTbl := make(map[string]WriteItem)
	for _, c := range *cmds {
		uuidTbl[*c.UUID] = c
	}

	for uuid, i := range r.SyncStatus {
		msg := ""
		handled := false

		// If it's a string, it should be just "ok" -- if not, that's an
		// unexpected state from Todoist.
		if state, ok := i.(string); ok {
			if state != "ok" {
				msg = fmt.Sprintf("unexpected error code %q (should be map or 'ok')", state)
			}
			handled = true
		}

		// If it's an error, it should be a map with numerous properties (like:
		// "error", "error_code", etc.
		if m, ok := i.(map[string]interface{}); ok {
			code, cok := m["error_code"]
			if !cok {
				code = "(no error_code)"
			}
			err, eok := m["error"]
			if !eok {
				err = "(no error message)"
			}
			cmd, cmdok := m["command_type"]
			if !cmdok {
				cmd = "(no command_type)"
			}
			msg = fmt.Sprintf("sync %q error code %v: %v", cmd, code, err)
			handled = true
		}

		if !handled {
			// If we got here, then we got an unknown response from Todoist. Sigh.
			msg = fmt.Sprintf("unknown response type %T from Todoist, data: %v", i, i)
		}

		c, ok := uuidTbl[uuid]
		if !ok {
			msg += fmt.Sprintf(", error for a different UUID than asked")
		}

		if len(msg) > 0 {
			msg += fmt.Sprintf(" (uuid %s)", uuid)
			ret = append(ret, writeError{Message: msg, Item: c})
		}
	}

	return ret
}

func (s *SyncV7API) listItems(p *Project) ([]Item, error) {
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

func (s *SyncV7API) findProject(name string) (*Project, error) {
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

func (s *SyncV7API) createProject(name, tempId string) WriteItem {
	return WriteItem{
		Type:   PTR(ProjectAdd),
		TempId: PTR(tempId),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   Project{Name: PTR(name)}}
}

func (s *SyncV7API) deleteProject(p *Project) WriteItem {
	return WriteItem{
		Type: PTR(ProjectDelete),
		UUID: PTR(uuid.NewV4().String()),
		Args: IdContainer{Ids: []int{*p.Id}}}
}

func (s *SyncV7API) createItem(projId string, t tasks.Task) WriteItem {
	log.Printf("Creating task %q (pos=%d) due %s", t.Content, t.Position, t.DueDateUTC.Format(DueDateFormatForWrite))

	// Todoist does not deal well with ItemOrder = 0; it won't honour order with
	// something at zero.
	pos := t.Position + 1

	return WriteItem{
		Type:   PTR(ItemAdd),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args: Item{
			Content:    &t.Content,
			Indent:     &t.Indent,
			ItemOrder:  &pos,
			DateString: PTR(t.DueDateUTC.Format(DateFormat)),
			DueDateUTC: PTR(t.DueDateUTC.Format(DueDateFormatForWrite)),
			ProjectId:  &projId}}
}

func (s *SyncV7API) updateItem(i Item, t tasks.Task) WriteItem {
	log.Printf("Updating task %q (pos=%d) due %s", t.Content, t.Position, t.DueDateUTC.Format(DueDateFormatForWrite))

	// Todoist does not deal well with ItemOrder = 0; it won't honour order with
	// something at zero.
	pos := t.Position + 1

	i.DateString = PTR(t.DueDateUTC.Format(DateFormat))
	i.DueDateUTC = PTR(t.DueDateUTC.Format(DueDateFormatForWrite))
	i.Indent = &t.Indent
	i.ItemOrder = &pos

	return WriteItem{
		Type:   PTR(ItemUpdate),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   i}
}

func (s *SyncV7API) deleteItem(i Item) WriteItem {
	return WriteItem{
		Type: PTR(ItemDelete),
		UUID: PTR(uuid.NewV4().String()),
		Args: IdContainer{Ids: []int{*i.Id}}}
}

// Returns a tasks.Project, whether or not it was found, and any error.
func (s *SyncV7API) LoadProject(name string) (tasks.Project, bool, error) {
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
		if !i.Valid() {
			log.Printf("Ignoring invalid item: %s", i)
			continue
		}

		var due time.Time
		if i.DueDateUTC == nil {
			log.Printf("No due date for %q, using empty value", *i.Content)
		} else {
			due, err = time.ParseInLocation(DueDateFormatForRead, *i.DueDateUTC, time.UTC)
			if err != nil {
				log.Printf("Could not parse %q: %v (ignoring, may generate diffs)", *i.DueDateUTC, err)
			}
		}

		ret.Tasks = append(ret.Tasks, tasks.Task{
			Content:    *i.Content,
			DueDateUTC: due,
			Indent:     *i.Indent,
			Position:   (*i.ItemOrder) - 1})
	}

	ret.External = &projectItems{ProjectId: *p.Id, Items: li}

	return ret, found, nil
}

func (s *SyncV7API) CreateProject(p tasks.Project) error {
	tempId := uuid.NewV4().String()

	cmds := Commands{s.createProject(p.Name, tempId)}
	for _, t := range p.Tasks {
		cmds = append(cmds, s.createItem(tempId, t))
	}

	_, err := s.Write(cmds)
	return err
}

func (s *SyncV7API) UpdateProject(p tasks.Project, diffs []tasks.Diff) error {
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
					cmds = append(cmds, s.updateItem(i, d.Task))
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
