package tasks

import (
	"time"
)

func Expand(cl []ChecklistItem, due time.Time) []Task {
	ret := make([]Task, 0, len(cl))

	for _, i := range cl {
		ret = append(ret, Task{
			Content: i.Template,
			Indent:  i.Indent,
			DueDate: i.Due(due),
		})
	}

	return ret
}
