package tasks

import (
	"reflect"
	"testing"
	"time"
)

func TestExpand(t *testing.T) {
	tripStart := time.Date(2016, 07, 15, 00, 00, 00, 00, time.UTC)
	cutoff := time.Date(2016, 07, 10, 00, 00, 00, 00, time.UTC)

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
	}, {
		in: []ChecklistItem{
			{Template: "before cutoff", Indent: 1, Days: -8},
			{Template: "after cutoff", Indent: 1, Days: -4},
		},
		want: []Task{{
			Content: "before cutoff",
			Indent:  1,
			DueDate: time.Date(2016, 07, 07, 20, 00, 00, 00, time.UTC),
		}},
	}}

	for _, c := range cases {
		got := Expand(c.in, tripStart, cutoff)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Expand(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}
