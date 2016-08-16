package tasks

import (
	"fmt"
	"reflect"
	"time"
)

const (
	Added int = iota
	Removed
	Changed
)

// A Project is the container for an expanded Checklist, loaded elsewhere in
// this package. In this tool, a Project is the unit of a Trip.
type Project struct {
	Name string

	// Tasks for this project. Order matters.
	Tasks []Task

	// Some sort of external data that you might wish to attach. E.g;
	// a separate (perhaps Todoist-specific) representation.
	External interface{}
}

func (p Project) Empty() bool { return len(p.Tasks) == 0 }

type Task struct {
	// The content of the task, e.g; "Buy milk"
	Content string

	// Due date for the task in UTC
	DueDateUTC time.Time

	// Indentation level for the task
	// TODO(srees): remove this? This is essentially nesting projects, maybe just do that?
	Indent int

	// Position of the task
	Position int
}

type Diff struct {
	Type int
	Task Task
}

func (d Diff) String() string {
	t := ""
	switch d.Type {
	case Added:
		t = "add"
	case Removed:
		t = "remove"
	case Changed:
		t = "change"
	}

	return fmt.Sprintf("{type=%s task=%v}", t, d.Task)
}

func (p Project) DiffTasks(other Project) []Diff {
	var ret []Diff

	pt := makeLookupTable(p.Tasks)
	ot := makeLookupTable(other.Tasks)

	// Added & Removed
	ret = append(ret, findDiffs(other.Tasks, pt, Added)...)
	ret = append(ret, findDiffs(p.Tasks, ot, Removed)...)

	// Changed
	for _, t := range p.Tasks {
		if task, ok := ot[t.Content]; ok {
			if !reflect.DeepEqual(t, task) {
				ret = append(ret, Diff{Type: Changed, Task: task})
			}
		}
	}

	return ret
}

func findDiffs(in []Task, table map[string]Task, typ int) []Diff {
	var ret []Diff
	for _, t := range in {
		if _, found := table[t.Content]; !found {
			ret = append(ret, Diff{Type: typ, Task: t})
		}
	}
	return ret
}

func makeLookupTable(tk []Task) map[string]Task {
	ret := make(map[string]Task)
	for _, t := range tk {
		ret[t.Content] = t
	}
	return ret
}
