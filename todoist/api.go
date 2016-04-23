package todoist

import (
	"encoding/json"
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

func (s *SyncV6API) makeRequest(url string, obj interface{}) error {
	c := buildConfig().Client(oauth2.NoContext, s.token)
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(data, &obj)
	return nil
}

func (s *SyncV6API) ReadItems() ([]Item, error) {

	params := url.Values{}
	params.Add("token", s.token.AccessToken)
	params.Add("seq_no", "0")
	params.Add("resource_types", "[\"items\"]")

	sr := SyncResponse{}
	err := s.makeRequest(ApiPath+"?"+params.Encode(), &sr)
	if err != nil {
		return nil, err
	}
	return sr.Items, nil
}
