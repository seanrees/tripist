package tasks

import (
	"reflect"
	"testing"
	"time"
)

func TestExpand(t *testing.T) {
	tripStart := time.Date(2016, 07, 15, 00, 00, 00, 00, time.UTC)
	tripEnd := time.Date(2016, 07, 20, 12, 30, 00, 00, time.UTC)
	now := time.Date(2016, 07, 01, 10, 00, 00, 00, time.UTC)
	stdCutoff := time.Date(2016, 07, 10, 00, 00, 00, 00, time.UTC)

	cases := []struct {
		in     []ChecklistItem
		cutoff time.Time
		want   []Task
	}{{
		in:   []ChecklistItem{},
		want: []Task{},
	}, {
		in:     []ChecklistItem{{Template: "foo", Indent: 1, Due: "14 days before start"}},
		cutoff: stdCutoff,
		want: []Task{{
			Content: "foo",
			Indent:  1,
			DueDate: time.Date(2016, 07, 01, 20, 00, 00, 00, time.UTC),
		}},
	}, {
		in: []ChecklistItem{
			{Template: "no day adjustment", Indent: 1, Due: "1 hour before start"},
			{Template: "end", Indent: 1, Due: "2 days before end"},
			{Template: "after", Indent: 1, Due: "4 hours after start"},
		},
		cutoff: tripEnd,
		want: []Task{
			{Content: "no day adjustment",
				Indent:  1,
				DueDate: time.Date(2016, 07, 14, 23, 00, 00, 00, time.UTC)},
			{Content: "end",
				Indent:   1,
				Position: 1,
				DueDate:  time.Date(2016, 07, 18, 20, 00, 00, 00, time.UTC)},
			{Content: "after",
				Indent:   1,
				Position: 2,
				DueDate:  time.Date(2016, 07, 15, 04, 00, 00, 00, time.UTC)},
		},
	}, {
		in:   []ChecklistItem{{Template: "due date already passed", Indent: 1, Due: "15 days before start"}},
		want: []Task{},
	}, {
		in: []ChecklistItem{
			{Template: "before cutoff", Indent: 1, Due: "8 days before start"},
			{Template: "after cutoff", Indent: 1, Due: "4 days before start"},
		},
		cutoff: stdCutoff,
		want: []Task{{
			Content: "before cutoff",
			Indent:  1,
			DueDate: time.Date(2016, 07, 07, 20, 00, 00, 00, time.UTC),
		}},
	}}

	for _, c := range cases {
		got := Expand(c.in, tripStart, tripEnd, now, c.cutoff)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Expand(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseDue(t *testing.T) {
	cases := []struct {
		in        string
		want      due
		wantError bool
	}{{
		in:   "1 hour from start",
		want: due{duration: -time.Hour},
	}, {
		in:   "3 hours after end",
		want: due{duration: 3 * time.Hour, end: true},
	}, {
		in:   "1 day before end",
		want: due{duration: -24 * time.Hour, end: true},
	}, {
		in:   "2 weeks before start",
		want: due{duration: -2 * 7 * 24 * time.Hour},
	}, {
		in:        "",
		wantError: true,
	}, {
		in:        "A days before start",
		wantError: true,
	}, {
		in:        "1 month before start", // month not supported.
		wantError: true,
	}, {
		in:        "1 day prior to start", // prior to not supported.
		wantError: true,
	}, {
		in:        "1 day before commencement",
		wantError: true,
	}}

	for _, c := range cases {
		got, err := parseDue(c.in)
		if err != nil {
			if !c.wantError {
				t.Errorf("parseDue(%q) error %v want no error", c.in, err)
			}
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseDue(%q) == %v want %v", c.in, got, c.want)
		}
	}

}
