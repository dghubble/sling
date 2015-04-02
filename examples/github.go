package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
)

const baseUrl = "https://api.github.com"

// Create a client

type Client struct {
	Issues *IssueService
	// other service endpoints...
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		Issues: NewIssueService(httpClient),
	}
}

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
	endpoint *sling.Sling
}

func NewIssueService(httpClient *http.Client) *IssueService {
	return &IssueService{
		endpoint: sling.New(httpClient),
	}
}

func (srvc IssueService) List(owner, repo string) ([]Issue, *http.Response, error) {
	var issues []Issue
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	req, _ := http.NewRequest("GET", baseUrl+path, nil)
	resp, err := srvc.endpoint.Fire(req, &issues)
	return issues, resp, err
}

// example use of the tiny Github API above

func main() {
	client := NewClient(&http.Client{})
	issues, _, _ := client.Issues.List("golang", "go")
	fmt.Printf("issues: %#v\n", issues)
}
