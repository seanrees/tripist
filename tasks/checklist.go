package tasks

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type ChecklistItem struct {
	Template string

	Indent int

	// This may not be expressive enough. E.g; some tasks might want to be quantised to units
	// of weekends rather than days. I suspect anything sub-day is not useful.
	Days int
}

func (t ChecklistItem) Due(due time.Time) time.Time {
	// Add (probably subtract) the relevant t.Days from the given due time, then add back
	// 20 hours to actually complete the task. This works best with tasks due at the *start*
	// of some day, e.g; at 00:00.
	return due.Add(time.Duration(t.Days) * 24 * time.Hour).Add(20 * time.Hour)
}

func Load(templateFilename string) ([]ChecklistItem, error) {
	f, err := os.Open(templateFilename)
	if err != nil {
		return nil, err
	}
	return load(f)
}

func load(ior io.Reader) ([]ChecklistItem, error) {
	var errors []string
	var ret []ChecklistItem
	r := csv.NewReader(ior)
	for l := 1; ; l++ {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			// package csv includes line numbers in errors.
			errors = append(errors, err.Error())
			continue
		}

		if len(rec) < 3 {
			errors = append(errors, fmt.Sprintf("line %d: not enough fields", l))
			continue
		}

		i, err := strconv.Atoi(rec[1])
		if err != nil {
			errors = append(errors, fmt.Sprintf("line %d: %v", l, err))
			continue
		}
		if i < 1 || i > 4 {
			errors = append(errors, fmt.Sprintf("line %d: indent out of range %d [1-4]", l, i))
			continue
		}

		d, err := strconv.Atoi(rec[2])
		if err != nil {
			errors = append(errors, fmt.Sprintf("line %d: %v", l, err))
			continue
		}

		ret = append(ret, ChecklistItem{
			Template: rec[0],
			Indent:   i,
			Days:     d,
		})
	}

	if len(errors) > 0 {
		return ret, fmt.Errorf("unable to load: %s", strings.Join(errors, ", "))
	}
	return ret, nil
}
