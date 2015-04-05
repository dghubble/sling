
# Sling

Sling is a Go REST client library for creating and sending requests. 

Sling structs store http request properties to provide a simple way to build requests, send them, and decode responses. Compose a Sling into your API client or endpoints and get started with the tutorial.

## Features

* Get/Post/Put/Patch/Delete/Head
* Resolves base and path urls
* Decodes JSON responses

## Install

    go get github.com/dghubble/sling

## Documentation

Read [GoDoc](https://godoc.org/github.com/dghubble/sling)

## Intro

In the simplest case, a Sling is configured (with `Base`, `Path`, `Get`, `Post`, etc.) and each time `HttpRequest()` is called, a new `http.Request` is returned.

    req, err := sling.New(nil).Base("https://api.github.io").HttpRequest()
    client.Do(req)


Slings are much more powerful though. Use them to create REST clients which wrap complex API endpoints. Copy a base Sling with `Request()` to avoid repeating common configuration.

    client := &http.Client{}
    base := sling.New(client).Base("https://api.twitter.com/1.1")
    issueServiceSling := base.Request().Path("/users")
    statusesSling := base.Request().Path("/statuses")

Avoid writing another client with encoding, decoding, and network logic. Define your models and use Sling's `Do(interface{})` to send a new Request and decode the response.

    type Tweet struct {
        ScreenName string `json:"screen_name"`
        Text       string `json:"text"`
        ...
    }

    var tweets []Tweet
    resp, err := statusesSling.Request().Get("/show.json").Do(&tweets)
    fmt.Println(tweets, resp, err)

## Tutorial

## Motivation

Sling was inspired by ideas from the [google/go-github](https://github.com/google/go-github) API.

Sling picks ideas from a handful of good REST API clients in order to provide common primitives for building REST APIs. The hope is that authors of new API clients can use Sling instead of reimplementing logic common to all clients.

## License

[MIT License](LICENSE)

