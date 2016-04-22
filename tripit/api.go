package tripit

import (
	"encoding/json"
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"net/http"
)

const (
	ApiPath = "https://api.tripit.com/v1"
)

type TripitV1 struct {
	accessToken *oauth.AccessToken
}

func NewTripitV1(at *oauth.AccessToken) *TripitV1 {
	return &TripitV1{accessToken: at}
}

func (t *TripitV1) makeClient() (*http.Client, error) {
	c, err := buildConsumer().MakeHttpClient(t.accessToken)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (t *TripitV1) makeRequest(url string, obj interface{}) error {
	c, err := t.makeClient()
	if err != nil {
		return err
	}

	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(data, obj)
	return nil
}

type ListParameters struct {
	Traveler       string
	Past           bool
	ModifiedSince  int64
	IncludeObjects bool
}

// Lists trips.
func (t *TripitV1) List(p *ListParameters) ([]Trip, error) {
	path := ApiPath + "/list/trip"

	if p != nil {
		path += fmt.Sprintf(
			"/traveler/%s/past/%v/modified_since/%v/include_objects/%v",
			p.Traveler, p.Past, p.ModifiedSince, p.IncludeObjects)
	}

	path += "/format/json"

	tr := TripitResponse{}
	err := t.makeRequest(path, &tr)
	if err != nil {
		return nil, err
	}

	return tr.Trip, nil
}
