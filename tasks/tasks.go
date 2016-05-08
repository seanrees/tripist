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

type Task struct {
	Template string

	Indent int

	// This may not be expressive enough. E.g; some tasks might want to be quantised to units
	// of weekends rather than days. I suspect anything sub-day is not useful.
	Days int
}

func (t Task) Due(start time.Time) time.Time {
	// Make these due on the day, at 20:00.
	return start.Add(time.Duration(t.Days) * 24 * time.Hour).Add(20 * time.Hour)
}

func Load(templateFilename string) ([]Task, error) {
	f, err := os.Open(templateFilename)
	if err != nil {
		return nil, err
	}
	return load(f)
}

func load(ior io.Reader) ([]Task, error) {
	var errors []string
	var ret []Task
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

		ret = append(ret, Task{
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
