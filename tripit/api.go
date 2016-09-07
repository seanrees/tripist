package tripit

import (
	"encoding/json"
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"log"
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

type callbackFn func([]byte) error

func (t *TripitV1API) makeRequest(url string, cb callbackFn) error {
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

	return cb(data)
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
	cb := func(data []byte) error {
		err := json.Unmarshal(data, &tr)
		if err != nil {
			// Workaround some broken Tripit behaviour: if there is only one
			// trip, we'll not be able to parse the JSON as a list-of-Trips. So
			// we parse out just the single Trip with a special message and then
			// copy it in.
			if len(tr.Trip) == 0 {
				log.Printf("No trips loaded and unmarshal error, trying single trip variant")
				sr := TripitSingleTripResponse{}
				if err := json.Unmarshal(data, &sr); err == nil {
					log.Printf("Loaded trip via single-trip variant")
					tr.Trip = append(tr.Trip, sr.Trip)
					return nil
				} else {
					log.Printf("Single-trip variant failed with: %v", err)
				}
			}
			return err
		}
		return nil
	}

	err := t.makeRequest(path, cb)
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
				for _, s := range a.Segments() {
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
