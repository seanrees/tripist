package tasks

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		csv  string
		want []ChecklistItem
		err  bool
	}{{
		csv:  "",
		want: []ChecklistItem{},
		err:  false,
	}, {
		csv:  "foo,1,1",
		want: []ChecklistItem{{"foo", 1, 1}},
		err:  false,
	}, {
		csv:  "foo,1,1\nbar,1,2",
		want: []ChecklistItem{{"foo", 1, 1}, {"bar", 1, 2}},
		err:  false,
	}, {
		csv:  "foo,1,not-a-num",
		want: []ChecklistItem{},
		err:  true,
	}, {
		csv:  "foo\nbar,1,1",
		want: []ChecklistItem{},
		err:  true,
	}, {
		csv:  "foo,1,1\nnot-enough-fields\nbar,3,2",
		want: []ChecklistItem{{"foo", 1, 1}, {"bar", 3, 2}},
		err:  true,
	}, {
		csv:  "foo,0,1\nbar,5,1\nbaz,1,1", // Indent out of range.
		want: []ChecklistItem{{"baz", 1, 1}},
		err:  true,
	}}

	for _, c := range cases {
		r := strings.NewReader(c.csv)
		got, err := load(r)

		if err != nil && !c.err {
			t.Errorf("load(%q) == error (%v), want no error", c.csv, err)
		}

		if lg, lw := len(got), len(c.want); lg != lw {
			t.Errorf("len(load(%q)) == %d, want %d", c.csv, lg, lw)
		}

		if len(c.want) > 0 && !reflect.DeepEqual(got, c.want) {
			t.Errorf("load(%q) == %v, want %v", c.csv, got, c.want)
		}
	}
}
