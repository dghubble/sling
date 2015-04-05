
# Sling [![Build Status](https://travis-ci.org/dghubble/sling.png)](https://travis-ci.org/dghubble/sling) [![GoDoc](http://godoc.org/github.com/dghubble/sling?status.png)](http://godoc.org/github.com/dghubble/sling)

Sling is a Go REST client library for creating and sending requests. 

Slings store http Request properties to simplify creating new Requests, sending them, and decoding responses. See the tutorial to learn how to compose a Sling into your API client.

## Features

* Base/Path - path extend a Sling for different endpoints
* Method Setters: Get/Post/Put/Patch/Delete/Head
* Encode structs into URL query parameters
* Receive decoded JSON success responses

## Install

    go get github.com/dghubble/sling

## Documentation

Read [GoDoc](https://godoc.org/github.com/dghubble/sling)

## Usage

Use a simple Sling to set request properties (`Path`, `QueryParams`, etc.) and then create a new `http.Request` by calling `Request()`.

```go
req, err := sling.New().Get("https://example.com").Request()
client.Do(req)
```

Slings are much more powerful though. Use them to create REST clients which wrap complex API endpoints. Copy a base Sling with `New()` to avoid repeating common configuration.

```go
const twitterApi = "https://https://api.twitter.com/1.1/"
base := sling.New().Base(twitterApi).Client(httpAuthClient)

users := base.New().Path("users/")
statuses := base.New().Path("statuses/")
search := base.New().Path("search/") 
```

### Encode / Decode

Define [url parameter structs](https://godoc.org/github.com/google/go-querystring/query) and use `QueryStruct` to encode query parameters.

```go
// Github Issue Parameters
type IssueParams struct {
    Filter    string `url:"filter,omitempty"`
    State     string `url:"state,omitempty"`
    Labels    string `url:"labels,omitempty"`
    Sort      string `url:"sort,omitempty"`
    Direction string `url:"direction,omitempty"`
    Since     string `url:"since,omitempty"`
}
```

```go
githubBase := sling.New().Base("https://api.github.com/").Client(httpClient)
path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)

params := {Sort: "updated"}
req, err := githubBase.New().Get(path).QueryStruct(params).Request()
```

Define expected value structs. Use `Receive(v interface{})` to send a new Request and decode the response into the value.

```go
// Github Issue (abbreviated)
type Issue struct {
    Id     int    `json:"id"`
    Url    string `json:"url"`
    Number int    `json:"number"`
    State  string `json:"state"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}
```

```go

var issues []Issue
req, err := githubBase.New().Get(path).QueryStruct(params).Receive(&issues)
fmt.Println(issues, resp, err)
```

### Build an API

APIs typically define an endpoint (also called a service) for each type of resource. For example, here is a tiny Github IssueService which supports the [repository issues](https://developer.github.com/v3/issues/#list-issues-for-a-repository) route.

```go
type IssueService struct {
    sling *sling.Sling
}

func NewIssueService(httpClient *http.Client) *IssueService {
    return &IssueService{
        sling: sling.New().Client(httpClient).Base(baseUrl),
    }
}

func (srvc IssueService) List(owner, repo string, params *IssueParams) ([]Issue, *http.Response, error) {
    var issues []Issue
    path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
    resp, err := srvc.sling.New().Get(path).QueryStruct(params).Receive(&issues)
    return issues, resp, err
}
```

## APIs using Sling

None yet! Create a Pull Request to add a link to your own API.

## Motivation

Many client libraries follow the lead of [google/go-github](https://github.com/google/go-github) (our inspiration!), but do so by reimplementing logic common to all clients.

This project borrows and abstracts those ideas into a Sling, an agnostic component any API client can use for creating and sending requests.

## License

[MIT License](LICENSE)

