package models

//  For Parsing the Pull Request Event metadata.
type PullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
}

type PullRequest struct {
	Number int  `json:"number"`
	Base   Base `json:"base"`
	Head   Head `json:"head"`
}

type Head struct {
	SHA string `json:"sha"`
}

type Base struct {
	Repo Repo `json:"repo"`
}

type Repo struct {
	FullName string `json:"full_name"`
}

//  For sending request to the LLM to review the code
type RequestBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// For holding the changed files and their content. 
type PrFileMetadata struct {
	Filename string `json:"filename"`
	Patch    string `json:"patch"`
}

// Holds the response from LLM for code review
type ResponseBody struct {
	Response string `json:"response"`
}

// Used for making POST request for adding review comments to the PR
type ReviewRequest struct {
	CommitID string          `json:"commit_id"`
	Body     string          `json:"body"`
	Event    string          `json:"event"`
	Comments []ReviewComments `json:"comments"`
}

type ReviewComments struct {
	Path     string `json:"path"`
	Position int    `json:"position"`
	Body     string `json:"body"`
}