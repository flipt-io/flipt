package github

// The GH OAuth needs both the clientID, clientSecret, and the requestToken for authentication.
// We will pass back to the client a clientToken, that the UI will use on every subsequent request.

// 1. Get requests from a client getting to the handler
// 2. Interact with Github API to get an access token
// 3. CreateAuthentication with the Github Method, and store the access token in the metadata
// 4.
