
# Sling

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

Use a simple Sling to set request properties (`Path`, `QueryParams`, etc.) and then create a new `http.Request` by calling `HttpRequest()`.

```go
req, err := sling.New().Client(client).Base("https://api.github.com/")
    .HttpRequest()
client.Do(req)
```

Slings are much more powerful though. Use them to create REST clients which wrap complex API endpoints. Copy a base Sling with `Request()` to avoid repeating common configuration.

```go
base := sling.New().Base("https://https://api.twitter.com/1.1/")
users := base.Request().Path("users/")
statuses := base.Request().Path("statuses/")
search := base.Request().Path("search/") 
```

Define url parameter structs and use `QueryStruct` to encode query parameters.

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
githubBase := sling.New().Base("https://api.github.com/")

params := {Sort: "updated"}
path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
req, err := githubBase.Request().Get(path).QueryStruct(params).HttpRequest()
```

Define JSON models and use `Do(v interface{})` to send a new Request and decode the response Body.

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
path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
req, err := githubBase.Request().Get(path).QueryStruct(params).Do(&issues)
fmt.Println(issues, resp, err)
```

### Building an API

A realistic API might define an endpoint (also called a service) for each type of resource. For example, here is a tiny Github IssueService which supports the [repository issues](https://developer.github.com/v3/issues/#list-issues-for-a-repository) route.

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
    resp, err := srvc.sling.Request().Get(path).QueryStruct(params).Do(&issues)
    return issues, resp, err
}
```

## APIs using Sling

None yet! Create a Pull Request to add a link to your own API.

## Motivation

Many client libraries follow the lead of [google/go-github](https://github.com/google/go-github) (our inspiration!), but do so by reimplementing logic common to all clients.

Sling combines those good ideas to provide general primitives for building REST API clients.

## License

[MIT License](LICENSE)

