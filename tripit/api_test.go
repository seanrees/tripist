package tripit

import (
	"testing"
	"time"
)

func makeDateTime(d, t, tz, utc string) DateTime {
	return DateTime{Date: d, Time: t, Timezone: tz, UtcOffset: utc}
}

func TestFixStartAndEndDates(t *testing.T) {
	trips := []Trip{
		{Id: "T0", DisplayName: "Trip 0", StartDate: "2016-08-16", EndDate: "2016-08-18"},
		{Id: "T1", DisplayName: "Trip 1", StartDate: "2016-08-16", EndDate: "2016-08-16"},
		{Id: "T2", DisplayName: "Trip 2", StartDate: "2016-08-17", EndDate: "2016-08-20"},
	}

	airObjects := []AirObject{{
		Id:     "A0",
		TripId: "T0",
		Segment: []Segment{{
			Id:            "A0S0",
			StartDateTime: makeDateTime("2016-08-16", "10:00:00", "Europe/Dublin", "+01:00"),
			EndDateTime:   makeDateTime("2016-08-16", "12:30:00", "Europe/Dublin", "+01:00"),
		}, {
			Id:            "A0S1",
			StartDateTime: makeDateTime("2016-08-16", "15:00:00", "Europe/Dublin", "+01:00"),
			EndDateTime:   makeDateTime("2016-08-16", "17:30:00", "Europe/Dublin", "+01:00"),
		}},
	}, {
		Id:     "A1",
		TripId: "T0",
		Segment: []Segment{{
			Id:            "A1S0",
			StartDateTime: makeDateTime("2016-08-18", "18:15:00", "Europe/Dublin", "+01:00"),
			EndDateTime:   makeDateTime("2016-08-18", "20:30:00", "Europe/Dublin", "+01:00"),
		}},
	}, {
		Id:     "A2",
		TripId: "T1",
		Segment: []Segment{{
			Id:            "A2S0",
			StartDateTime: makeDateTime("2016-08-16", "06:15:00", "Europe/Dublin", "+01:00"),
			EndDateTime:   makeDateTime("2016-08-16", "09:30:00", "Europe/Dublin", "+01:00"),
		}},
	}}

	loc, err := time.LoadLocation("Europe/Dublin")
	if err != nil {
		t.Fatalf("Could not load timezone: %v", err)
	}

	resp := TripitResponse{Trip: trips, AirObject: airObjects}
	err = fixStartAndEndDates(&resp)
	if err != nil {
		t.Errorf("fixStartAndEndDates() == error (%v), want no error", err)
	}

	want := []Trip{{
		Id:              "T0",
		ActualStartDate: time.Date(2016, 8, 16, 10, 00, 00, 00, loc),
		ActualEndDate:   time.Date(2016, 8, 18, 20, 30, 00, 00, loc),
	}, {
		Id:              "T1",
		ActualStartDate: time.Date(2016, 8, 16, 6, 15, 00, 00, loc),
		ActualEndDate:   time.Date(2016, 8, 16, 9, 30, 00, 00, loc),
	}, {
		Id:              "T2",
		ActualStartDate: time.Date(2016, 8, 17, 00, 00, 00, 00, time.UTC),
		ActualEndDate:   time.Date(2016, 8, 20, 00, 00, 00, 00, time.UTC),
	}}

	for _, tr := range resp.Trip {
		for _, wa := range want {
			if wa.Id == tr.Id {
				if g, w := tr.ActualStartDate, wa.ActualStartDate; !g.Equal(w) {
					t.Errorf("fixStartAndEndDates() %s.ActualStartDate == %v, want %v", tr.Id, g, w)
				}
				if g, w := tr.ActualEndDate, wa.ActualEndDate; !g.Equal(w) {
					t.Errorf("fixStartAndEndDates() %s.ActualEndDate == %v, want %v", tr.Id, g, w)
				}
			}
		}
	}
}
