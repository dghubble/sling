package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
	//"golang.org/x/oauth2"
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

type IssueRequest struct {
	Title     string   `json:"title,omitempty"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
	Labels    []string `json:"labels,omitempty"`
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

func (s IssueService) List(owner, repo string, params *IssueParams) ([]Issue, *http.Response, error) {
	var issues []Issue
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
	resp, err := s.sling.New().Get(path).QueryStruct(params).Receive(&issues)
	return issues, resp, err
}

func (s *IssueService) Create(owner, repo string, issueBody *IssueRequest) (*Issue, *http.Response, error) {
	issue := new(Issue)
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
	resp, err := s.sling.New().Post(path).JsonBody(issueBody).Receive(issue)
	return issue, resp, err
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
	// Use httpClient with your token to perform authenticated operations
	// ts := &tokenSource{
	// 	&oauth2.Token{AccessToken: "_your_token_"},
	// }
	// httpClient := oauth2.NewClient(oauth2.NoContext, ts)
	client := NewClient(nil)
	// body := &IssueRequest{
	// 	Title: "Test title",
	// 	Body:  "Some test issue",
	// }
	// issue, resp, err := client.IssueService.Create("username", "my-repo", body)
	// fmt.Println(issue, resp, err)

	// Unauthenticated
	params := &IssueParams{Sort: "updated"}
	issues, resp, err := client.IssueService.List("golang", "go", params)
	fmt.Println(issues, resp, err)
}

// for using oauth2
// type tokenSource struct {
// 	token *oauth2.Token
// }

// func (t *tokenSource) Token() (*oauth2.Token, error) {
// 	return t.token, nil
// }
