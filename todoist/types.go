package todoist

const (
	True  = 1
	False = 0
)

type SyncResponse struct {
	Items []Item
}

type Item struct {
	Id             int64   `json:"id"`
	UserId         int64   `json:"user_id"`
	ProjectId      int64   `json:"project_id"`
	Content        string  `json:"content"`
	DateString     string  `json:"date_string"`
	DateLang       string  `json:"date_lang"`
	DueDateUTC     string  `json:"due_date_utc"`
	Priority       int64   `json:"priority"`
	Indent         int64   `json:"indent"`
	ItemOrder      int64   `json:"item_order"`
	DayOrder       int64   `json:"day_order"`
	Collapsed      int64   `json:"collapsed"`
	Labels         []int64 `json:"labels"`
	AssignedByUid  int64   `json:"assigned_by_uid"`
	ResponsibleUid int64   `json:"responsible_uid"`
	Checked        int64   `json:"checked"`
	InHistory      int64   `json:"in_history"`
	IsDeleted      int64   `json:"is_deleted"`
	IsArchived     int64   `json:"is_archived"`
	SyncId         int64   `json:"sync_id"`
	DateAdded      string  `json:"date_added"`
}
