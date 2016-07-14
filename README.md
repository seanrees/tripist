# Tripist

This tool creates Todoist projects for upcoming trips in Tripit. It scans your
upcoming trips, evaluates your travel checklist, and creates the project and
tasks as appropriate.

The current default is to create new tasks when they are 1 week away. This can be
overriden with -task_cutoff_days.

This program reads a simple checklist in CSV format like this:
```
# Action / Text to Display, Indentation Level, Days Before Trip
Pre-trip, 1, 0
Charge Headphones, 2, -2
Packing List, 1, 0
Toiletries, 2, -1
Passport, 2, -1
Clothes, 2, -1
```

This will produce a Todoist project like this:
```
. Pre-trip
`--- Charge Headphones (2 days before)
. Packing List
`--- Toiletries (1 day before)
`--- Passport (1 day before)
`--- Clothes (1 day before)
```

To use this, you'll need API keys. If I know you, just ask and I'll give
you the ones I'm using.

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
% tripist
```
