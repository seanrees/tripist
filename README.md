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
2. After login, browse to: https://www.tripit.com/oauth/authorize?oauth_callback=https%3A%2F%2Ffreyr.erifax.org%2Ftripist%2F&oauth_token=605dc50e77b8e232badffd42a353df15f6e9a598
3. Grant access and copy the 'oauth_token' parameter displayed.

Enter oauth_token: <CODE>

% ./tripist -authorize_todoist
1. Browse to: https://todoist.com/oauth/authorize?access_type=offline&client_id=db0d3f274c864a9f8a429283514b92d1&redirect_uri=https%3A%2F%2Ffreyr.erifax.org%2Ftripist%2F&response_type=code&scope=data%3Aread_write%2Cdata%3Adelete%2Cproject%3Adelete&state=erifax
2. Grant access and copy the 'code' parameter displayed.

Enter verification code: <CODE>
```
