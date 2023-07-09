package tasks

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type ChecklistItem struct {
	Template string

	Indent int

	Due string
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

		ret = append(ret, ChecklistItem{
			Template: rec[0],
			Indent:   i,
			Due:      rec[2],
		})
	}

	if len(errors) > 0 {
		return ret, fmt.Errorf("unable to load: %s", strings.Join(errors, ", "))
	}
	return ret, nil
}
