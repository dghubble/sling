/*
Package sling is a Go REST client library for creating and sending requests.

Slings store http Request properties to simplify sending requests and decoding
responses. Check the examples to learn how to compose a Sling into your API
client.

Use a simple Sling to set request properties (Path, QueryParams, etc.) and
then create a new http.Request by calling Request().

	req, err := sling.New().Get("https://example.com").Request()
	client.Do(req)

Slings are much more powerful though. Use them to create REST clients
which wrap complex API endpoints. Copy a base Sling with New() to avoid
repeating common configuration.

	const twitterApi = "https://https://api.twitter.com/1.1/"
	base := sling.New().Base(twitterApi).Client(httpAuthClient)

	users := base.New().Path("users/")
	statuses := base.New().Path("statuses/")

Choose an http method, set query parameters, and send the request.

	statuses.New().Get("show.json").QueryStruct(params).Receive(tweet)

The usage README provides more details about setting headers, query parameters,
body data, and decoding a typed response after sending.
*/
package sling
