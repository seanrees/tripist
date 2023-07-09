package todoist

import (
	"strings"
	"testing"
)

func TestRewriteProjectName(t *testing.T) {
	cases := []struct {
		s    string
		want string
	}{{
		s:    "San Francisco, CA (March, 2022)",
		want: "San Francisco CA March 2022",
	}, {
		s:    "",
		want: "",
	}, {
		s:    "No Rewrite",
		want: "No Rewrite",
	}, {
		s:    "#\"()|&!",
		want: "",
	}, {
		s:    "-_.",
		want: "-_.",
	}}

	for _, c := range cases {
		got := rewriteProjectName(c.s)
		if got != c.want {
			t.Errorf("rewriteProjectName(%q) == %q, want %q", c.s, got, c.want)
		}
	}
}
func TestCheckErrors(t *testing.T) {
	api := NewSyncV9API(nil)

	// The actual commands don't matter, so we choose createProject as it's the
	// simplest to create.
	cmds := Commands{
		api.createProject("test:ok", "ignore-me"), // no error.
		api.createProject("test:unexpected-state", "ignore-me"),
		api.createProject("test:error-map", "ignore-me"),
		api.createProject("test:error-map-missing-fields", "ignore-me"),
		api.createProject("test:unknown-type", "ignore-me"),
	}

	// Populate the error states.
	errmap := make(map[string]interface{})
	errmap["error_code"] = 150
	errmap["error"] = "you frobbed the wrong widget"
	errmap["error_tag"] = "FROB_WIDGET"

	wr := WriteResponse{}
	wr.SyncStatus = make(map[string]interface{})
	wr.SyncStatus[*cmds[0].UUID] = "ok"
	wr.SyncStatus[*cmds[1].UUID] = "unexpected-state"
	wr.SyncStatus[*cmds[2].UUID] = errmap
	wr.SyncStatus[*cmds[3].UUID] = make(map[string]interface{})
	wr.SyncStatus[*cmds[4].UUID] = 10
	wr.SyncStatus["not a uuid"] = "ok" // adds another error.

	errs := api.checkErrors(&cmds, &wr)
	if got, want := len(errs), len(cmds); got != want {
		t.Errorf("len(checkErrors()) == %d, want %d", got, want)
	}

	want := map[string]string{
		"test:unexpected-state":         "unexpected error code",
		"test:error-map":                "sync \"FROB_WIDGET\" error code 150: you frobbed the wrong widget",
		"test:error-map-missing-fields": "(no error_code): (no error message)",
		"test:unknown-type":             "unknown response type",
	}

	tcToUUID := make(map[string]string)
	for _, c := range cmds {
		p, ok := c.Args.(Project)
		if !ok {
			panic("Args should be a Project")
		}
		tcToUUID[*p.Name] = *c.UUID
	}

	for _, err := range errs {
		if err.Item.Type == nil {
			if !strings.Contains(err.Message, "different UUID") {
				t.Errorf("checkErrors(): unexpected error: %q", err.Message)
			}
			continue
		}

		p, ok := err.Item.Args.(Project)
		if !ok {
			panic("Args should be a Project")
		}

		wantSubstr, found := want[*p.Name]
		switch {
		case !found:
			t.Errorf("checkErrors(%s): no error expected, got %q", *p.Name, err.Message)

		case !strings.Contains(err.Message, wantSubstr):
			t.Errorf("checkErrors(%s): got %q missing %q", *p.Name, err.Message, wantSubstr)
		}

		if got, want := *err.Item.UUID, tcToUUID[*p.Name]; got != want {
			t.Errorf("checkErrors(%s): wrong UUID got %s want %s", *p.Name, got, want)
		}
	}
}
