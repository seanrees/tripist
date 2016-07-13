package tasks

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Make it so we can compare DiffTasks output consistently without worrying about order.
type byPosTypeContent []Diff

func (t byPosTypeContent) Len() int      { return len(t) }
func (t byPosTypeContent) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t byPosTypeContent) Less(i, j int) bool {
	return t[i].Position < t[j].Position || t[i].Type < t[j].Type || strings.Compare(t[i].Task.Content, t[j].Task.Content) < 0
}

func TestDiffTasks(t *testing.T) {
	p1 := Project{
		Name:  "Project 1",
		Tasks: []Task{{Content: "p1t1"}, {Content: "p1t2"}},
	}
	p1b := Project{
		Name:  "Project 1",
		Tasks: []Task{{Content: "p1t1"}, {Content: "p1t2", Indent: 1}},
	}
	p2 := Project{
		Name:  "Project 2",
		Tasks: []Task{{Content: "p2t1"}, {Content: "p2t2"}},
	}
	empty := Project{
		Name:  "Project 3",
		Tasks: []Task{},
	}

	cases := []struct {
		p     Project
		other Project
		want  []Diff
	}{{
		p:     p1,
		other: p2,
		want: []Diff{
			{Position: 0, Type: Added, Task: p2.Tasks[0]},
			{Position: 1, Type: Added, Task: p2.Tasks[1]},
			{Position: 0, Type: Removed, Task: p1.Tasks[0]},
			{Position: 1, Type: Removed, Task: p1.Tasks[1]},
		},
	}, {
		p:     p1,
		other: empty,
		want: []Diff{
			{Position: 0, Type: Removed, Task: p1.Tasks[0]},
			{Position: 1, Type: Removed, Task: p1.Tasks[1]},
		},
	}, {
		p:     empty,
		other: p2,
		want: []Diff{
			{Position: 0, Type: Added, Task: p2.Tasks[0]},
			{Position: 1, Type: Added, Task: p2.Tasks[1]},
		},
	}, {
		p:     p1,
		other: p1b,
		want:  []Diff{{Position: 1, Type: Changed, Task: p1b.Tasks[1]}},
	}}

	for _, c := range cases {
		got := c.p.DiffTasks(c.other)
		sort.Sort(byPosTypeContent(got))
		sort.Sort(byPosTypeContent(c.want))

		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("DiffTasks(%v) == %v, want %v", c.other, got, c.want)
		}
	}
}

func TestFindDiffs(t *testing.T) {
	cases := []struct {
		tasks []Task
		table map[string]Task
		typ   int
		want  []Diff
	}{{
		tasks: []Task{},
		table: map[string]Task{},
		typ:   Added,
	}, {
		tasks: []Task{{Content: "present"}, {Content: "not present"}},
		table: map[string]Task{
			"present": Task{Content: "present"},
			"extra":   Task{Content: "extra"},
		},
		typ:  Added,
		want: []Diff{{Position: 1, Type: Added, Task: Task{Content: "not present"}}},
	}, {
		// Testing pass through of typ and correct position.
		tasks: []Task{{Content: "not present"}, {Content: "present"}, {Content: "also not present"}},
		table: map[string]Task{"present": Task{Content: "present"}},
		typ:   Removed,
		want: []Diff{
			{Position: 0, Type: Removed, Task: Task{Content: "not present"}},
			{Position: 2, Type: Removed, Task: Task{Content: "also not present"}}},
	}}

	for _, c := range cases {
		got := findDiffs(c.tasks, c.table, c.typ)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("findDiffs(%v, %v, %v) == %v, want %v", c.tasks, c.table, c.typ, got, c.want)
		}
	}
}

func TestMakeLookupTable(t *testing.T) {
	cases := []struct {
		in   []Task
		want map[string]Task
	}{{
		in:   []Task{},
		want: map[string]Task{},
	}, {
		in: []Task{{Content: "One"}, {Content: "Two"}},
		want: map[string]Task{
			"One": Task{Content: "One"},
			"Two": Task{Content: "Two"},
		},
	}}

	for _, c := range cases {
		got := makeLookupTable(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("makeLookupTable(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}
