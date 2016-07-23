package tasks

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type due struct {
	duration time.Duration

	// end is true if the duration should be counted from the end of
	// the trip.
	end bool
}

func abs(t time.Duration) time.Duration {
	if t < 0 {
		return -t
	}
	return t
}

// Expand expands a travel checklist into a list of Tasks. If a task has a due date
// after cutoff, it is ignored.
func Expand(cl []ChecklistItem, start, end, cutoff time.Time) []Task {
	ret := []Task{}

	for pos, i := range cl {
		var dd time.Time
		d, err := parseDue(i.Due)
		if err != nil {
			log.Printf("Could not process due for task %q: %v (ignored)", i.Template, err)
			continue
		}
		if d.end {
			dd = end.Add(d.duration)
		} else {
			dd = start.Add(d.duration)
		}
		if abs(d.duration) >= 24*time.Hour {
			dd = time.Date(dd.Year(), dd.Month(), dd.Day(), 20, 00, 00, 00, time.UTC)
		}

		if dd.Before(cutoff) {
			ret = append(ret, Task{
				Content:  i.Template,
				Indent:   i.Indent,
				DueDate:  dd,
				Position: pos,
			})
		}
	}

	return ret
}

// parseDue expands a humanized due string into a due structure. A due
// string looks like: "16 hours before start" or "1 day after end"
func parseDue(s string) (due, error) {
	var ret due
	parts := strings.SplitN(strings.ToLower(s), " ", 4)
	if len(parts) < 4 {
		return ret, fmt.Errorf("due date not fully specified %q", s)
	}

	t, err := strconv.Atoi(parts[0])
	if err != nil {
		return ret, err
	}

	switch d := parts[1]; d {
	case "minute":
		fallthrough
	case "minutes":
		ret.duration = time.Duration(t) * time.Minute

	case "hour":
		fallthrough
	case "hours":
		ret.duration = time.Duration(t) * time.Hour

	case "day":
		fallthrough
	case "days":
		ret.duration = time.Duration(t*24) * time.Hour

	case "week":
		fallthrough
	case "weeks":
		ret.duration = time.Duration(t*24*7) * time.Hour

	default:
		// Just try it.
		ret.duration, err = time.ParseDuration(d)
		if err != nil {
			return ret, fmt.Errorf("unknown unit %q", d)
		}
	}

	switch parts[2] {
	case "after":
		// Nothing.

	case "from":
		fallthrough
	case "before":
		ret.duration = -ret.duration

	default:
		return ret, fmt.Errorf("unknown relation %q", parts[2])
	}

	switch parts[3] {
	case "departure":
		fallthrough
	case "start":
		// Nothing.

	case "return":
		fallthrough
	case "end":
		ret.end = true

	default:
		return ret, fmt.Errorf("unknown reference %q", parts[3])
	}

	return ret, nil
}
