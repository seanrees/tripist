# Tripist

This is a simple proof-of-concept to create Todoist projects (e.g; to
pack) from TripIt trips. The PoC will only scan for trips upto 45 days
away.

To use this, you'll need API keys. If I know you, just ask and I'll give
you the ones I'm using.

To use this, you'll need to modify the source and call the Authorize
functions and then update the OAuth tokens in the source. I plan to make
this a bit more user-friendly later.
```
tripit.Authorize()
todoist.Authorize()
```
