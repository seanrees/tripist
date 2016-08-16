package tripit

import "time"

type TripitResponse struct {
	Timestamp int64 `json:"timestamp,string"`
	NumBytes  int64 `json:"num_bytes,string"`
	Trip      []Trip
	AirObject []AirObject
	// Other data includes: LodgingObject, WeatherObject, and Profile.
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
	TripInvitees           []Invitee
	TripPurposes           TripPurpose

	// Set by fixStartAndEndDates.
	ActualStartDate time.Time
	ActualEndDate   time.Time
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

type Invitee struct {
	IsOwner    bool `json:"is_owner,string"`
	IsReadOnly bool `json:"is_read_only,string"`
	IsTraveler bool `json:"is_traveler,string"`
}

type TripPurpose struct {
	IsAutoGenerated bool   `json:"is_auto_generated,string"`
	PurposeTypeCode string `json:"purpose_type_code"`
}

type AirObject struct {
	Id              string `json:"id"`
	TripId          string `json:"trip_id"`
	SupplierConfNum string `json:"supplier_conf_num"`
	Segment         []Segment
}

type Segment struct {
	Id               string `json:"id"`
	StartAirportCode string `json:"start_airport_code"`
	StartCityName    string `json:"start_city_name"`
	StartTerminal    string `json:"start_terminal"`
	EndAirportCode   string `json:"end_airport_code"`
	EndCityName      string `json:"end_city_name"`
	EndTerminal      string `json:"end_terminal"`
	StartDateTime    DateTime
	EndDateTime      DateTime
}

type DateTime struct {
	Date      string `json:"date"`
	Time      string `json:"time"`
	Timezone  string `json:"timezone"`
	UtcOffset string `json:"utc_offset"`
}

func (dt *DateTime) Parse() (time.Time, error) {
	loc, err := time.LoadLocation(dt.Timezone)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("2006-01-02 15:04:05", dt.Date+" "+dt.Time, loc)
}
