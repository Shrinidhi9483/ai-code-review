package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Shrinidhi9483/ai-code-review/models"
)

const GitHubAPI = "https://api.github.com"

func GetLatestCommitID(repo string, prNumber int) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls/%d", repo, prNumber)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "",  err
	}

	var pr models.PullRequest
	json.Unmarshal(body, &pr)

	if pr.Head.SHA == "" {
		return "", fmt.Errorf("commit ID not found")
	}

	return pr.Head.SHA, nil
}

func GetChangedFiles(repo string, prNumber int) ([]models.PrFileMetadata, error) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d/files", GitHubAPI, repo, prNumber)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var files []models.PrFileMetadata
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &files)
	return files, nil
}

func PostReviewComments(repo string, prNumber int, commitID string, comments []models.ReviewComments) {
	var reviewRequest models.ReviewRequest
	reviewRequest.CommitID = commitID
	reviewRequest.Event = "COMMENT"
	reviewRequest.Body = fmt.Sprintf("Automated Code Review from AI for PR Number: %v, Commit: %v",  prNumber, commitID)
	reviewRequest.Comments = comments

	// Convert to JSON
	jsonData, err := json.Marshal(reviewRequest)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/repos/%s/pulls/%d/reviews", GitHubAPI, repo, prNumber)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	client.Do(req)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to post review comment: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusOK {
		log.Println("Success! Status Code:", resp.StatusCode)
	} else {
		log.Printf("Request failed! Status Code: %d, Reason: %s\n", resp.StatusCode, resp.Status)
	}
}

func formatComments(comments []string) string {
	var result string
	for _, comment := range comments {
		result += "- " + comment + "\n"
	}
	return result
}
