package todoist

import (
	"fmt"
	"reflect"
)

const (
	// Types that can be read.
	Items    = "items"
	Projects = "projects"

	// Types that can be written.
	ItemAdd       = "item_add"
	ItemDelete    = "item_delete"
	ItemUpdate    = "item_update"
	ProjectAdd    = "project_add"
	ProjectDelete = "project_delete"
)

// This is like the %+v verb in fmt, but dereferences pointers.
func stringify(obj interface{}) string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	if t.Kind() != reflect.Struct || !v.IsValid() {
		return "error"
	}

	result := "{"
	for i := 0; i < t.NumField(); i++ {
		ft := t.FieldByIndex([]int{i})
		fv := v.FieldByIndex([]int{i})

		name := ft.Name
		value := "<invalid>"

		// Deference the pointer if appropriate.
		if fv.Kind() == reflect.Ptr && !fv.IsNil() {
			fv = fv.Elem()
		}
		switch {
		case fv.Kind() == reflect.Ptr && fv.IsNil():
			value = "<nil>"
		case fv.Kind() == reflect.String:
			value = fmt.Sprintf("%q", fv.Interface())
		default:
			value = fmt.Sprintf("%v", fv.Interface())
		}

		result += fmt.Sprintf("%s:%s ", name, value)
	}
	result += "}"

	return result
}

type ReadResponse struct {
	SequenceNumber *int `json:"seq_no"`
	UserId         *int
	Items          []Item
	Projects       []Project
}

func (i ReadResponse) String() string {
	return stringify(i)
}

type WriteItem struct {
	// One of: WriteItem, WriteProject.
	Type   *string     `json:"type"`
	TempId *string     `json:"temp_id"`
	UUID   *string     `json:"uuid"`
	Args   interface{} `json:"args"`
}

func (i WriteItem) String() string {
	return stringify(i)
}

type IdContainer struct {
	Id string `json:"id"`
}

type Commands []WriteItem

type WriteResponse struct {
	SequenceNumber *int              `json:"seq_no"`
	TempIdMapping  map[string]string `json:"temp_id_mapping"`

	// Map of the Command UUID to one of either a string ("ok") or to another map[string]interface{}.
	SyncStatus map[string]interface{} `json:"sync_status"`
}

func (i WriteResponse) String() string {
	return stringify(i)
}

type Item struct {
	Id             *string `json:"id"`
	UserId         *string `json:"user_id"`
	ProjectId      *string `json:"project_id"`
	Content        *string `json:"content"`
	Due            *Due    `json:"due"`
	Priority       *int    `json:"priority"`
	ParentId       *string `json:"parent_id"`
	ChildOrder     *int    `json:"child_order"`
	DayOrder       *int    `json:"day_order"`
	Collapsed      *bool   `json:"collapsed"`
	Labels         []int   `json:"labels"`
	AssignedByUid  *string `json:"assigned_by_uid"`
	ResponsibleUid *string `json:"responsible_uid"`
	Checked        *bool   `json:"checked"`
	IsDeleted      *bool   `json:"is_deleted"`
	SyncId         *int    `json:"sync_id"`
}

func (i Item) Valid() bool {
	return i.Content != nil && i.Id != nil
}

func (i Item) String() string {
	return stringify(i)
}

type Due struct {
	// RFC3339 or YYYY-MM-DDTHH:MM:SS (no trailing Z.)
	Date        string  `json:"date"`
	Timezone    *string `json:"timezone"`
	IsRecurring bool    `json:"is_recurring"`
	String      string  `json:"string"`
	Language    string  `json:"lang"`
}

type Project struct {
	Id           *string `json:"id"`
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	ParentId     *string `json:"parent_id"`
	ChildOrder   *int    `json:"child_order"`
	Collapsed    *bool   `json:"collapsed"`
	Shared       *bool   `json:"shared,*string"`
	IsDeleted    *bool   `json:"is_deleted"`
	IsArchived   *bool   `json:"is_archived"`
	InboxProject *bool   `json:"inbox_project,*string"`
	TeamInbox    *bool   `json:"team_inbox,*string"`
}

func (i Project) String() string {
	return stringify(i)
}

// For riding along when interfacing with the local tasks API.
type projectItems struct {
	ProjectId string
	Items     []Item
}

type writeError struct {
	Message string
	Item    WriteItem
}
