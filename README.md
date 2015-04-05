
# Sling

Sling is a Go REST client library for creating and sending requests. 

Slings store http Request properties to simplify creating new Requests, sending them, and decoding responses. See the tutorial to learn how to compose a Sling into your API client.

## Features

* Setters for Get/Post/Put/Patch/Delete/Head
* Resolves base and path urls
* Decodes JSON responses

## Install

    go get github.com/dghubble/sling

## Documentation

Read [GoDoc](https://godoc.org/github.com/dghubble/sling)

## Intro

Use a simple Sling to set request properties (`Path`, `QueryParams`, etc.) and create a new `http.Request` by calling `HttpRequest()`.

```go
req, err := sling.New(nil).Base("https://api.github.io").HttpRequest()
client.Do(req)
```

Slings are much more powerful though. Use them to create REST clients which wrap complex API endpoints. Copy a base Sling with `Request()` to avoid repeating common configuration.

```go
client := &http.Client{}
base := sling.New(client).Base("https://api.twitter.com/1.1")
issueServiceSling := base.Request().Path("/users")
statusesSling := base.Request().Path("/statuses")
```

Avoid writing another client with encoding, decoding, and network logic. Define your models and use Sling's `Do(interface{})` to send a new Request and decode the response.

```go
type Tweet struct {
    ScreenName string `json:"screen_name"`
    Text       string `json:"text"`
    ...
}
```

```go
var tweets []Tweet
resp, err := statusesSling.Request().Get("/show.json").Do(&tweets)
fmt.Println(tweets, resp, err)
```

## Tutorial

## Motivation

Sling was inspired by ideas from the [google/go-github](https://github.com/google/go-github) API.

Sling picks ideas from a handful of good REST API clients in order to provide common primitives for building REST APIs. The hope is that authors of new API clients can use Sling instead of reimplementing logic common to all clients.

## License

[MIT License](LICENSE)

