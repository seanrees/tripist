package todoist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/seanrees/tripist/tasks"
	"github.com/twinj/uuid"
	"golang.org/x/oauth2"
)

const (
	ApiPath = "https://todoist.com/API/v9/sync"
)

func PTR(s string) *string { return &s }

// Removes special characters from the project name.
// Todoist only supports these characters in project names: - _ and .
// They specifically call out # " ( ) | & ! , as exclusions -- so we'll
// start with these.
//
// Src: https://todoist.com/help/articles/create-a-project
func rewriteProjectName(s string) string {
	ret := s
	chars := []string{"#", "\"", "(", ")", "|", "&", "!", ","}

	for _, c := range chars {
		ret = strings.ReplaceAll(ret, c, "")
	}

	return ret
}

type SyncV9API struct {
	token *oauth2.Token
}

func NewSyncV9API(t *oauth2.Token) *SyncV9API {
	return &SyncV9API{token: t}
}

func (s *SyncV9API) makeRequest(path string, data url.Values, obj interface{}) error {
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
func (s *SyncV9API) Read(types []string) (ReadResponse, error) {
	resp := ReadResponse{}
	params := url.Values{}
	params.Add("token", s.token.AccessToken)
	params.Add("sync_token", "*")

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

func (s *SyncV9API) Write(c Commands) (WriteResponse, error) {
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

func (s *SyncV9API) checkErrors(cmds *Commands, r *WriteResponse) []writeError {
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
			tag, tagok := m["error_tag"]
			if !tagok {
				tag = "(no error_tag)"
			}
			msg = fmt.Sprintf("sync %q error code %v: %v", tag, code, err)
			handled = true
		}

		if !handled {
			// If we got here, then we got an unknown response from Todoist. Sigh.
			msg = fmt.Sprintf("unknown response type %T from Todoist, data: %v", i, i)
		}

		c, ok := uuidTbl[uuid]
		if !ok {
			msg += ", error for a different UUID than asked"
		}

		if len(msg) > 0 {
			msg += fmt.Sprintf(" (uuid %s)", uuid)
			ret = append(ret, writeError{Message: msg, Item: c})
		}
	}

	return ret
}

func (s *SyncV9API) listItems(p *Project) ([]Item, error) {
	resp, err := s.Read([]string{Items})

	if err != nil {
		log.Printf("Could not read Todoist items: %v", err)
		return nil, err
	}

	var ret []Item
	for _, i := range resp.Items {
		if *i.ProjectId == *p.Id {
			ret = append(ret, i)
		}
	}

	log.Printf("Loaded %d items (%d for this project) from Todoist", len(resp.Items), len(ret))

	return ret, nil
}

func (s *SyncV9API) findProject(name string) (*Project, error) {
	resp, err := s.Read([]string{Projects})
	if err != nil {
		log.Printf("Could not read Todoist projects: %v", err)
		return nil, err
	}

	rp := rewriteProjectName(name)
	for _, p := range resp.Projects {
		if *p.Name == rp {
			log.Printf("Found existing project %q id=%v", *p.Name, *p.Id)
			return &p, nil
		}
	}
	return nil, nil
}

func (s *SyncV9API) createProject(name, tempId string) WriteItem {
	return WriteItem{
		Type:   PTR(ProjectAdd),
		TempId: PTR(tempId),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   Project{Name: PTR(rewriteProjectName(name))}}
}

func (s *SyncV9API) deleteProject(p *Project) WriteItem {
	return WriteItem{
		Type: PTR(ProjectDelete),
		UUID: PTR(uuid.NewV4().String()),
		Args: IdContainer{Id: *p.Id}}
}

func (s *SyncV9API) createItem(projId string, parent *string, t tasks.Task) WriteItem {
	log.Printf("Creating task %q (pos=%d) due %s", t.Content, t.Position, t.DueDateUTC.Format(time.RFC3339))

	// Todoist does not deal well with ItemOrder = 0; it won't honour order with
	// something at zero.
	pos := t.Position + 1

	return WriteItem{
		Type:   PTR(ItemAdd),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args: Item{
			Content:    &t.Content,
			ParentId:   parent,
			ChildOrder: &pos,
			Due: &Due{
				Date:     t.DueDateUTC.Format(time.RFC3339),
				Timezone: PTR("UTC"),
			},
			ProjectId: &projId}}
}

func (s *SyncV9API) updateItem(i Item, t tasks.Task) WriteItem {
	log.Printf("Updating task %q (pos=%d) due %s", t.Content, t.Position, t.DueDateUTC.Format(time.RFC3339))

	// Todoist does not deal well with ItemOrder = 0; it won't honour order with
	// something at zero.
	pos := t.Position + 1

	i.Due.Date = t.DueDateUTC.Format(time.RFC3339)
	i.Due.Timezone = PTR("UTC")
	i.ChildOrder = &pos

	return WriteItem{
		Type:   PTR(ItemUpdate),
		TempId: PTR(uuid.NewV4().String()),
		UUID:   PTR(uuid.NewV4().String()),
		Args:   i}
}

func (s *SyncV9API) deleteItem(i Item) WriteItem {
	return WriteItem{
		Type: PTR(ItemDelete),
		UUID: PTR(uuid.NewV4().String()),
		Args: IdContainer{Id: *i.Id}}
}

// Returns a tasks.Project, whether or not it was found, and any error.
func (s *SyncV9API) LoadProject(name string) (tasks.Project, bool, error) {
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

	parents := []string{"0"}

	for _, i := range li {
		if !i.Valid() {
			log.Printf("Ignoring invalid item: %s", i)
			continue
		}

		var due time.Time
		if i.Due == nil {
			log.Printf("No due date for %q, using empty value", *i.Content)
		} else {
			due, err = time.ParseInLocation(time.RFC3339, i.Due.Date, time.UTC)
			if err != nil {
				log.Printf("Could not parse %q: %v (ignoring, may generate diffs)", i.Due.Date, err)
			}
		}

		// Remapping task hierarchy onto indentation levels.
		indent := 0
		if i.ParentId != nil {
			for idx, id := range parents {
				if id == *i.ParentId {
					indent = idx + 1
				}
			}
			for len(parents) < indent+1 {
				parents = append(parents, "0")
			}
			parents[indent] = *i.Id
		} else {
			parents[0] = *i.Id
		}

		ret.Tasks = append(ret.Tasks, tasks.Task{
			Content:    *i.Content,
			DueDateUTC: due,
			// For historical reasons in Todoist, we're 1-based.
			Indent:    indent + 1,
			Completed: *i.Checked,
			Position:  (*i.ChildOrder) - 1})
	}

	ret.External = &projectItems{ProjectId: *p.Id, Items: li}

	return ret, found, nil
}

func (s *SyncV9API) CreateProject(p tasks.Project) error {
	tempId := uuid.NewV4().String()

	cmds := Commands{s.createProject(p.Name, tempId)}
	cmds = append(cmds, s.addTasks(tempId, p.Tasks)...)

	_, err := s.Write(cmds)
	return err
}

func (s *SyncV9API) addTasks(tempId string, ts []tasks.Task) Commands {
	var cmds Commands

	// Support indent -> parentId conversion.
	max := 0
	for _, t := range ts {
		if t.Indent > max {
			max = t.Indent
		}
	}

	parents := make([]*string, max+1)
	for _, t := range ts {
		i := s.createItem(tempId, parents[t.Indent-1], t)
		parents[t.Indent] = i.TempId

		cmds = append(cmds, i)
	}

	return cmds
}

func (s *SyncV9API) UpdateProject(p tasks.Project, diffs []tasks.Diff) error {
	tp, ok := p.External.(*projectItems)
	if !ok {
		return fmt.Errorf("missing or invalid external project pointer on %q", p.Name)
	}

	var cmds Commands
	var adds []tasks.Task

	for _, d := range diffs {
		switch d.Type {
		case tasks.Added:
			adds = append(adds, d.Task)

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

	cmds = append(cmds, s.addTasks(tp.ProjectId, adds)...)

	if len(cmds) > 0 {
		_, err := s.Write(cmds)
		return err
	} else {
		log.Printf("No commands to run to update project %q", p.Name)
	}

	return nil
}
