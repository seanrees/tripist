// Binary gcmap produces gcmap.com URL for your historical and upcoming travel.
package main

import (
	"flag"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/mrjones/oauth"
	"github.com/seanrees/tripist/internal/config"
	"github.com/seanrees/tripist/internal/tripit"
)

var (
	startYear = flag.Int("start_year", 0, "Only consider trips after (incusive) of start_year.")
	endYear   = flag.Int("end_year", 9999, "Only consider trips before (inclusive) of end_year.")
)

type byStartDate []tripit.Segment

func (a byStartDate) Len() int      { return len(a) }
func (a byStartDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byStartDate) Less(i, j int) bool {
	s, err := a[i].StartDateTime.Parse()
	if err != nil {
		return true
	}
	e, err := a[j].StartDateTime.Parse()
	if err != nil {
		return false
	}
	return s.Before(e)
}

type airportCounter map[string]int
type airportCounterSort struct {
	ac   airportCounter
	keys []string
}

func (a airportCounter) increment(code string) {
	a[code] = a[code] + 1
}

// Returns sorted keys (descending).
func (a airportCounter) sortedKeys() []string {
	var keys []string

	for k := range a {
		keys = append(keys, k)
	}

	s := airportCounterSort{a, keys}
	sort.Sort(s)

	return s.keys
}

func (a airportCounterSort) Len() int      { return len(a.keys) }
func (a airportCounterSort) Swap(i, j int) { a.keys[i], a.keys[j] = a.keys[j], a.keys[i] }
func (a airportCounterSort) Less(i, j int) bool {
	return a.ac[a.keys[i]] > a.ac[a.keys[j]]
}

func main() {
	const configFilename = "user.json"
	const tripistConfigFilename = "tripist.json"

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stderr)

	var conf config.UserKeys
	if _, err := os.Stat(configFilename); err == nil {
		conf, err = config.Read(configFilename)
		if err != nil {
			log.Fatalf("Unable to read configuration: %v\n", err)
		}
	} else {
		log.Fatalf("Configuration file %q does not exist.\n", configFilename)
	}

	if err := config.LoadAPIKeys(tripistConfigFilename); err != nil {
		log.Println("Using built-in keys.")
	} else {
		log.Printf("Loaded API keys from %s", tripistConfigFilename)
	}

	token := &oauth.AccessToken{
		Token:  conf.TripitToken,
		Secret: conf.TripitSecret,
	}

	api := tripit.NewTripitV1API(token)
	lp := tripit.ListParameters{
		Traveler:       "true",
		Past:           true,
		ModifiedSince:  0,
		IncludeObjects: true,
	}

	var page int64 = 1
	var segmentsByTrip [][]tripit.Segment

	for {
		lp.PageNum = page
		tr, err := api.ListRaw(&lp)

		if err != nil {
			log.Printf("Got error: %v\n", err)
		}

		log.Printf("Loaded %d (page=%d) trips\n", len(tr.Trip), page)
		for i, a := range tr.Trip {
			log.Printf("Trip %3d: %s\n", (page-1)*tr.PageSize+int64(i), a.DisplayName)
		}

		for _, trip := range tr.Trip {
			if sy, ey := trip.ActualStartDate.Year(), trip.ActualEndDate.Year(); sy < *startYear || ey > *endYear {
				log.Printf("Ignoring %q because starts/ends in %d/%d, want in [%d-%d]", trip.DisplayName, sy, ey, *startYear, *endYear)
				continue
			}

			var segs []tripit.Segment
			for _, ao := range tr.AirObject {
				if ao.TripId == trip.Id {
					segs = append(segs, ao.Segments()...)
				}
			}
			segmentsByTrip = append(segmentsByTrip, segs)
		}

		page += 1
		if page > tr.MaxPage {
			break
		}
	}

	var paths []string
	airports := make(airportCounter)

	for _, segs := range segmentsByTrip {
		sort.Sort(byStartDate(segs))
		last := ""
		path := ""
		for _, s := range segs {
			if last != s.StartAirportCode {
				if len(path) > 0 {
					paths = append(paths, path)
				}
				path = s.StartAirportCode
				airports.increment(s.StartAirportCode)
			}
			path += "-" + s.EndAirportCode
			last = s.EndAirportCode

			airports.increment(s.EndAirportCode)
		}
		if len(path) > 0 {
			paths = append(paths, path)
		}
	}

	log.Printf("Top airports:\n")
	for i, k := range airports.sortedKeys() {
		log.Printf("  #%02d: %s (%d visits)\n", i, k, airports[k])
	}

	log.Printf("http://www.gcmap.com/mapui?P=%s", strings.Join(paths, ","))
}
