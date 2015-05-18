package main

import (
	"fmt"
	"github.com/dghubble/sling"
	"golang.org/x/oauth2"
	"net/http"
	"os"
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
type IssueListParams struct {
	Filter    string `url:"filter,omitempty"`
	State     string `url:"state,omitempty"`
	Labels    string `url:"labels,omitempty"`
	Sort      string `url:"sort,omitempty"`
	Direction string `url:"direction,omitempty"`
	Since     string `url:"since,omitempty"`
}

// https://developer.github.com/v3/#client-errors
type GithubError struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors"`
	DocumentationURL string `json:"documentation_url"`
}

func (e GithubError) Error() string {
	return fmt.Sprintf("github: %v %+v %v", e.Message, e.Errors, e.DocumentationURL)
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

func (s *IssueService) List(params *IssueListParams) ([]Issue, *http.Response, error) {
	issues := new([]Issue)
	githubError := new(GithubError)
	resp, err := s.sling.New().Path("issues").QueryStruct(params).Receive(issues, githubError)
	if err != nil {
		return *issues, resp, err
	}
	return *issues, resp, githubError
}

func (s *IssueService) ListByRepo(owner, repo string, params *IssueListParams) ([]Issue, *http.Response, error) {
	issues := new([]Issue)
	githubError := new(GithubError)
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
	resp, err := s.sling.New().Get(path).QueryStruct(params).Receive(issues, githubError)
	if err != nil {
		return *issues, resp, err
	}
	return *issues, resp, githubError
}

func (s *IssueService) Create(owner, repo string, issueBody *IssueRequest) (*Issue, *http.Response, error) {
	issue := new(Issue)
	githubError := new(GithubError)
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
	resp, err := s.sling.New().Post(path).JsonBody(issueBody).Receive(issue, githubError)
	if err != nil {
		return issue, resp, err
	}
	return issue, resp, githubError
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

	// Github Unauthenticated API
	client := NewClient(nil)
	params := &IssueListParams{Sort: "updated"}
	issues, _, _ := client.IssueService.ListByRepo("golang", "go", params)
	fmt.Printf("Public golang/go Issues:\n%v\n", issues)

	// Github OAuth2 Example - httpClient handles authorization
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Run 'export GITHUB_ACCESS_TOKEN=mytoken' and retry to list your public/private issues")
		os.Exit(0)
	}

	ts := &tokenSource{
		&oauth2.Token{AccessToken: accessToken},
	}
	httpClient := oauth2.NewClient(oauth2.NoContext, ts)

	client = NewClient(httpClient)
	issues, _, _ = client.IssueService.List(params)
	fmt.Printf("Your Github Issues:\n%v\n", issues)

	// body := &IssueRequest{
	// 	Title: "Test title",
	// 	Body:  "Some test issue",
	// }
	// issue, _, _ := client.IssueService.Create("username", "myrepo", body)
	// fmt.Println(issue)
}

// for using golang/oauth2
type tokenSource struct {
	token *oauth2.Token
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	return t.token, nil
}
