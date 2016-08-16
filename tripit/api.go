package tripit

import (
	"encoding/json"
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	ApiPath = "https://api.tripit.com/v1"
)

type TripitV1API struct {
	accessToken *oauth.AccessToken
}

func NewTripitV1API(at *oauth.AccessToken) *TripitV1API {
	return &TripitV1API{accessToken: at}
}

func (t *TripitV1API) makeClient() (*http.Client, error) {
	c, err := buildConsumer().MakeHttpClient(t.accessToken)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (t *TripitV1API) makeRequest(url string, obj interface{}) error {
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
func (t *TripitV1API) List(p *ListParameters) ([]Trip, error) {
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

	err = fixStartAndEndDates(&tr)
	if err != nil {
		return nil, err
	}

	return tr.Trip, nil
}

// Corrects the Start and End of a Trip using flight data.
func fixStartAndEndDates(tr *TripitResponse) error {
	for i := range tr.Trip {
		t := &tr.Trip[i]

		var min, max time.Time
		for _, a := range tr.AirObject {
			if a.TripId == t.Id {
				for _, s := range a.Segment {
					for _, d := range []DateTime{s.StartDateTime, s.EndDateTime} {
						ti, err := d.Parse()
						if err != nil {
							return err
						}
						if max.IsZero() || ti.After(max) {
							max = ti
						}
						if min.IsZero() || ti.Before(min) {
							min = ti
						}
					}
				}
			}
		}

        var err error
		if min.IsZero() {
			t.ActualStartDate, err = t.Start()
			if err != nil {
				return err
			}
		} else {
			t.ActualStartDate = min
		}

		if max.IsZero() {
			t.ActualEndDate, err = t.End()
			if err != nil {
				return err
			}
		} else {
			t.ActualEndDate = max
		}
	}

	return nil
}
