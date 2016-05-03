# Tripist

This is a simple proof-of-concept to create Todoist projects (e.g; to
pack) from TripIt trips. The PoC will only scan for trips upto 45 days
away.

To use this, you'll need API keys. If I know you, just ask and I'll give
you the ones I'm using.

Once you have the API keys, you need to authorize the application to read
your Tripit data and generate Todoist projects. To do this:
```
% go build github.com/seanrees/tripist

% ./tripist -authorize_tripit
1. Login to TripIt in your browser.
2. After login, browse to: <URL>
3. Grant access & copy the verification code.

Enter verification code: <CODE>

% ./tripist -authorize_todoist
1. Browse to: <URL>
2. Copy the code= parameter from your URL bar.

Enter verification code: <CODE>
```
