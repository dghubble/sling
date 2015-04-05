package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
)

const baseUrl = "https://api.github.com"

// Define models

type Issue struct {
	Id     int    `json:"id"`
	Url    string `json:"url"`
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// Implement services

type IssueService struct {
	sling *sling.Sling
}

func NewIssueService(httpClient *http.Client) *IssueService {
	return &IssueService{
		sling: sling.New().Client(httpClient).Base(baseUrl),
	}
}

func (srvc IssueService) List(owner, repo string) ([]Issue, *http.Response, error) {
	var issues []Issue
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	resp, err := srvc.sling.Request().Get(path).Do(&issues)
	return issues, resp, err
}

// (optional) Create a client to wrap services

// Tiny Github client
type Client struct {
	IssueService *IssueService
	// other service endpoints...
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		IssueService: NewIssueService(httpClient),
	}
}

// example use of the tiny Github API above

func main() {
	client := NewClient(&http.Client{})
	issues, resp, err := client.IssueService.List("golang", "go")
	fmt.Printf("%#v\n", issues)
	fmt.Println(resp, err)
}
