package tasks

import (
	"reflect"
	"testing"
	"time"
)

func TestExpand(t *testing.T) {
	due := time.Date(2016, 07, 15, 00, 00, 00, 00, time.UTC)

	cases := []struct {
		in   []ChecklistItem
		want []Task
	}{{
		in:   []ChecklistItem{},
		want: []Task{},
	}, {
		in: []ChecklistItem{{Template: "foo", Indent: 1, Days: -14}},
		want: []Task{{
			Content: "foo",
			Indent:  1,
			DueDate: time.Date(2016, 07, 01, 20, 00, 00, 00, time.UTC),
		}},
	}}

	for _, c := range cases {
		got := Expand(c.in, due)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Expand(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}
