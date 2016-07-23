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
		csv:  "foo,1,1 day before start",
		want: []ChecklistItem{{"foo", 1, "1 day before start"}},
		err:  false,
	}, {
		csv:  "foo,1,bizzle\nbar,1,wizzle",
		want: []ChecklistItem{{"foo", 1, "bizzle"}, {"bar", 1, "wizzle"}},
		err:  false,
	}, {
		csv:  "foo\nbar,1,error",
		want: []ChecklistItem{},
		err:  true,
	}, {
		csv:  "foo,1,e\nnot-enough-fields\nbar,3,f",
		want: []ChecklistItem{{"foo", 1, "e"}, {"bar", 3, "f"}},
		err:  true,
	}, {
		csv:  "foo,0,oof\nbar,5,rab\nbaz,1,zab", // Indent out of range.
		want: []ChecklistItem{{"baz", 1, "zab"}},
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
