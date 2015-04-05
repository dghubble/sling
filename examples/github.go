package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
)

const baseUrl = "https://api.github.com/"

// Define models

// Simplified https://developer.github.com/v3/issues/#response
type Issue struct {
	Id     int    `json:"id"`
	Url    string `json:"url"`
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// https://developer.github.com/v3/issues/#parameters
type IssueParams struct {
	Filter    string `url:"filter,omitempty"`
	State     string `url:"state,omitempty"`
	Labels    string `url:"labels,omitempty"`
	Sort      string `url:"sort,omitempty"`
	Direction string `url:"direction,omitempty"`
	Since     string `url:"since,omitempty"`
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

func (srvc IssueService) List(owner, repo string, params *IssueParams) ([]Issue, *http.Response, error) {
	var issues []Issue
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
	resp, err := srvc.sling.New().Get(path).QueryStruct(params).Receive(&issues)
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

// example use of the tiny Github API

func main() {
	client := NewClient(&http.Client{})
	params := &IssueParams{Sort: "updated"}
	issues, resp, err := client.IssueService.List("golang", "go", params)
	fmt.Printf("%#v\n", issues)
	fmt.Println(resp, err)
}
