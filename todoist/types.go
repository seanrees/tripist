package todoist

import (
	"fmt"
	"reflect"
)

const (
	True  = 1
	False = 0

	// Types that can be read.
	Items    = "items"
	Projects = "projects"

	// Types that can be written.
	ItemAdd       = "item_add"
	ItemDelete    = "item_delete"
	ItemUpdate    = "item_update"
	ProjectAdd    = "project_add"
	ProjectDelete = "project_delete"

	DateFormat = "01/02/2006"

	// Todoist's API is pretty inconsistent. This is another one: you create
	// tasks with DueDateUTC in DueDateFormatForWrite. These get reported back
	// in DueDateFormatForRead, which is just like Go's time.RFC1123Z except
	// missing a comma (wtf.)
	DueDateFormatForRead  = "Mon 02 Jan 2006 15:04:05 -0700"
	DueDateFormatForWrite = "2006-01-02T15:04"
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
	Ids []int `json:"ids"`
}

type Commands []WriteItem

type WriteResponse struct {
	SequenceNumber *int           `json:"seq_no"`
	TempIdMapping  map[string]int `json:"temp_id_mapping"`

	// Map of the Command UUID to one of either a string ("ok") or to another map[string]interface{}.
	SyncStatus map[string]interface{} `json:"sync_status"`
}

func (i WriteResponse) String() string {
	return stringify(i)
}

type Item struct {
	Id     *int `json:"id"`
	UserId *int `json:"user_id"`
	// This field is normally an Integer, but in a Write, can be a *string-ified UUID.
	ProjectId      interface{} `json:"project_id"`
	Content        *string     `json:"content"`
	DateString     *string     `json:"date_string"`
	DateLang       *string     `json:"date_lang"`
	DueDateUTC     *string     `json:"due_date_utc"`
	Priority       *int        `json:"priority"`
	Indent         *int        `json:"indent"`
	ItemOrder      *int        `json:"item_order"`
	DayOrder       *int        `json:"day_order"`
	Collapsed      *int        `json:"collapsed"`
	Labels         []int       `json:"labels"`
	AssignedByUid  *int        `json:"assigned_by_uid"`
	ResponsibleUid *int        `json:"responsible_uid"`
	Checked        *int        `json:"checked"`
	InHistory      *int        `json:"in_history"`
	IsDeleted      *int        `json:"is_deleted"`
	IsArchived     *int        `json:"is_archived"`
	SyncId         *int        `json:"sync_id"`
	DateAdded      *string     `json:"date_added"`
}

func (i Item) String() string {
	return stringify(i)
}
func (i Item) ProjectIdInt() int {
	v, ok := i.ProjectId.(float64) // Sigh, this is what Go marshals it as.
	if !ok {
		return -1
	}
	return int(v)
}

type Project struct {
	Id                *int    `json:"id"`
	UserId            *int    `json:"user_id"`
	Name              *string `json:"name"`
	Color             *int    `json:"color"`
	Indent            *int    `json:"indent"`
	ItemOrder         *int    `json:"item_order"`
	Collapsed         *int    `json:"collapsed"`
	Shared            *bool   `json:"shared,*string"`
	IsDeleted         *int    `json:"is_deleted"`
	IsArchived        *int    `json:"is_archived"`
	ArchivedDate      *string `json:"archived_date"`
	ArchivedTimestamp *int    `json:"archived_timestamp"`
	InboxProject      *bool   `json:"inbox_project,*string"`
	TeamInbox         *bool   `json:"team_inbox,*string"`
}

func (i Project) String() string {
	return stringify(i)
}

// For riding along when interfacing with the local tasks API.
type projectItems struct {
	ProjectId int
	Items     []Item
}

type writeError struct {
	Message string
	Item    WriteItem
}
