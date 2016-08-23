# Tripist

This tool creates Todoist projects for upcoming trips in Tripit. It uses a configurable checklist (described below) to create tasks for each upcoming trip.

To be most useful, this program should be run once daily to create/update any tasks for upcoming trips. By default, the Tripist only creates tasks that are due within the next week. This is changeable with ```-task_cutoff_days```.

## Usage

### Flags
```
Usage of bin/tripist:
  -authorize_todoist
       	Perform Todoist Authorization. This is an exclusive flag.
  -authorize_tripit
       	Perform Tripit Authorization. This is an exclusive flag.
  -checklist_csv string
       	Travel checklist CSV file. (default "checklist.csv")
  -task_cutoff_days int
       	Create tasks upto this many days in advance of their due date. (default 7)
  -verify_todoist
       	Perform Todoist API validation. This is an exclusive flag.
```

### Checklist

For each trip, Tripist expands a trip checklist into Todoist tasks. The checklist is a CSV file with the following headers:
1. Action / Task to do (text)
2. Indentation level (1 to 4)
3. Due Date (humanised string, e.g; 1 day before start, 2 days after end)

A sample checklist looks like this:
```
# Action / Text to Display, Indentation Level, Days Before Trip
Pre-trip, 1, 1 hour before start
Charge Headphones, 2, 2 days before start
Checkin to flight, 2, 1 day before start
Hail taxi to Airport, 2, 3 hours before start
Packing List, 1, 1 day brefore start
Toiletries, 2, 1 day before start
Passport, 2, 1 day before start
Clothes for DAYS, 2, 1 day before start
Post-trip, 1, 1 day after end
Order groceries, 2, 1 day before end
```

This will produce a Todoist project like this:
```
. Pre-trip
`--- Charge Headphones (2 days before)
`--- Checkin to flight (1 day before)
`--- Hail taxi to Airport (3 hours before)
. Packing List
`--- Toiletries (1 day before)
`--- Passport (1 day before)
`--- Clothes for 2 days (1 day before)
. Post-trip
`--- Order groceries (1 day before you return)
```

Note, Tripist will expand the special keyword ```DAYS``` in a checklist with the number of days in the trip.

### API Keys
To use this, you'll need API keys. If I know you, just ask and I'll give
you the ones I'm using. If I don't know you, you'll need to create them with Tripit and Todoist independently. It's free and easy (at the time of this writing).

#### API Key Location
Store the API keys in ```tripist.json``` in Tripist's runtime working directory.

This looks like:
```
% cat tripist.json
{
    "TripitAPIKey": "",
    "TripitAPISecret": "",
    "TodoistClientID": "",
    "TodoistClientSecret": ""
}
```

Once you have the API keys, you need to authorize the application to read
your Tripit data and generate Todoist projects. To do this:
```
% go build github.com/seanrees/tripist

% ./tripist -authorize_tripit
1. Login to TripIt in your browser.
2. After login, browse to: <URL>
3. Grant access and copy the 'oauth_token' parameter displayed.

Enter oauth_token: <CODE>

% ./tripist -authorize_todoist
1. Browse to: <URL>
2. Grant access and copy the 'code' parameter displayed.

Enter verification code: <CODE>
```

Once configured and a checklist is in checklist.csv, just run it like so:
```
% go install github.com/seanrees/tripist
% bin/tripist
```
