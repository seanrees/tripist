package tasks

import (
	"time"
)

// Expand expands a travel checklist into a list of Tasks. If a task has a due date
// after cutoff, it is ignored.
func Expand(cl []ChecklistItem, start, cutoff time.Time) []Task {
	ret := []Task{}

	for pos, i := range cl {
		due := i.Due(start)
		if due.Before(cutoff) {
			ret = append(ret, Task{
				Content:  i.Template,
				Indent:   i.Indent,
				DueDate:  due,
				Position: pos,
			})
		}
	}

	return ret
}
