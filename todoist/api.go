package todoist

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/url"
)

const (
	ApiPath = "https://todoist.com/API/v6/sync"
)

type SyncV6API struct {
	token *oauth2.Token
}

func NewSyncV6API(t *oauth2.Token) *SyncV6API {
	return &SyncV6API{token: t}
}

func (s *SyncV6API) makeRequest(path string, data url.Values, obj interface{}) error {
	c := buildConfig().Client(oauth2.NoContext, s.token)
	resp, err := c.PostForm(path, data)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	j, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(j, &obj); err != nil {
		return err
	}

	return nil
}

// Reads specific types and returns a ReadResponse. Possible types are in constants:
// Projects, Items.
//
// sequenceNumber should be zero or a ReadResponse.SequenceNumber if you desire an
// incremental read.
func (s *SyncV6API) Read(types []string, sequenceNumber int) (ReadResponse, error) {
	resp := ReadResponse{}
	params := url.Values{}
	params.Add("token", s.token.AccessToken)
	params.Add("seq_no", fmt.Sprintf("%d", sequenceNumber))

	t, err := json.Marshal(types)
	if err != nil {
		return resp, nil
	}
	params.Add("resource_types", string(t))

	if err := s.makeRequest(ApiPath, params, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *SyncV6API) Write(c Commands) (WriteResponse, error) {
	resp := WriteResponse{}
	params := url.Values{}
	params.Add("token", s.token.AccessToken)

	cmds, err := json.Marshal(c)
	if err != nil {
		return resp, nil
	}
	params.Add("commands", string(cmds))

	if err := s.makeRequest(ApiPath, params, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}
