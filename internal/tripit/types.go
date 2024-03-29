package tripit

import (
	"fmt"
	"time"
)

type TripitResponse struct {
	Timestamp int64 `json:"timestamp,string"`
	NumBytes  int64 `json:"num_bytes,string"`
	PageNum   int64 `json:"page_num,string"`
	PageSize  int64 `json:"page_size,string"`
	MaxPage   int64 `json:"max_page,string"`

	// If you have 1 Trip, the TripIt API does returns a single result and
	// not a list.
	Trip []Trip

	AirObject []AirObject
	// Other data includes: LodgingObject, WeatherObject, and Profile.
}

// Tripit brokenness: it will return a single Trip object (instead of
// list-of-Trips) if you have one trip upcoming. In some circumstances, it will
// return a single AirObject instead of a list. Sigh.
type TripitSingleTripResponse struct {
	Trip Trip
}
type TripitSingleAirObjectResponse struct {
	AirObject AirObject
}

type Trip struct {
	Id                     string `json:"id"`
	LastModified           int64  `json:"last_modified,string"`
	StartDate              string `json:"start_date"`
	EndDate                string `json:"end_date"`
	DisplayName            string `json:"display_name"`
	ImageUrl               string `json:"image_url"`
	IsPrivate              bool   `json:"is_private,string"`
	IsTraveler             bool   `json:"is_traveler,string"`
	PrimaryLocation        string `json:"primary_location_address"`
	PrimaryLocationAddress Address
	TripPurposes           TripPurpose

	// Set by fixStartAndEndDates.
	ActualStartDate time.Time `json:"-"`
	ActualEndDate   time.Time `json:"-"`
}

func (t *Trip) Start() (time.Time, error) {
	return time.Parse(time.RFC3339, t.StartDate+"T00:00:00Z")
}
func (t *Trip) End() (time.Time, error) {
	return time.Parse(time.RFC3339, t.EndDate+"T00:00:00Z")
}

type Address struct {
	Address   string
	City      string
	State     string
	Country   string
	Latitude  float32 `json:",string"`
	Longitude float32 `json:",string"`
}

type TripPurpose struct {
	IsAutoGenerated bool   `json:"is_auto_generated,string"`
	PurposeTypeCode string `json:"purpose_type_code"`
}

type AirObject struct {
	Id              string `json:"id"`
	TripId          string `json:"trip_id"`
	SupplierConfNum string `json:"supplier_conf_num"`

	// Tripit returns either a list of Segments or a single Segment inline.
	// The JSON unmarshaler will unmarshal this into a map[key]val or a [map[key]val]].
	Segment interface{}
}

func (ao *AirObject) Segments() []Segment {
	var ret []Segment
	switch ao.Segment.(type) {
	case map[string]interface{}:
		seg := ao.Segment.(map[string]interface{})
		ret = append(ret, newSegment(seg))

	case []interface{}:
		segs := ao.Segment.([]interface{})
		for _, kv := range segs {
			switch kv.(type) {
			case map[string]interface{}:
				seg := kv.(map[string]interface{})
				ret = append(ret, newSegment(seg))
			default:
				fmt.Printf("Unknown type in manyton: %T\n", kv)
			}
		}

	default:
		fmt.Printf("Unknown type for Segment: %T\n", ao.Segment)
		return nil
	}

	return ret
}

type Segment struct {
	StartDateTime    DateTime
	EndDateTime      DateTime
	StartAirportCode string
	EndAirportCode   string
}

func newSegment(kv map[string]interface{}) Segment {
	sd := newDateTime(kv["StartDateTime"].(map[string]interface{}))

	ed := DateTime{"", "", "", ""}
	if kv["EndDateTime"] != nil {
		ed = newDateTime(kv["EndDateTime"].(map[string]interface{}))
	}

	sac := ""
	eac := ""

	if kv["start_airport_code"] != nil {
		sac = kv["start_airport_code"].(string)
	}

	if kv["end_airport_code"] != nil {
		eac = kv["end_airport_code"].(string)
	}

	return Segment{
		StartDateTime:    sd,
		EndDateTime:      ed,
		StartAirportCode: sac,
		EndAirportCode:   eac,
	}
}

type DateTime struct {
	Date      string
	Time      string
	Timezone  string
	UtcOffset string `json:"utc_offset"`
}

func newDateTime(kv map[string]interface{}) DateTime {
	return DateTime{
		Date:      fmt.Sprintf("%v", kv["date"]),
		Time:      fmt.Sprintf("%v", kv["time"]),
		Timezone:  fmt.Sprintf("%v", kv["timezone"]),
		UtcOffset: fmt.Sprintf("%v", kv["utc_offset"]),
	}
}

func (dt *DateTime) Parse() (time.Time, error) {
	loc, err := time.LoadLocation(dt.Timezone)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("2006-01-02 15:04:05", dt.Date+" "+dt.Time, loc)
}
