package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const GitHubAPI = "https://api.github.com"

func GetChangedFiles(repo string, prNumber int) ([]string, error) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d/files", GitHubAPI, repo, prNumber)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var files []struct {
		Filename string `json:"filename"`
		Patch    string `json:"patch"`
	}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &files)

	var changedFiles []string
	for _, file := range files {
		changedFiles = append(changedFiles, file.Patch)
	}
	return changedFiles, nil
}

func PostReviewComments(repo string, prNumber int, comments []string) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d/reviews", GitHubAPI, repo, prNumber)
	payload := map[string]interface{}{
		"body":   "AI Code Review Suggestions:\n" + formatComments(comments),
		"event":  "COMMENT",
	}
	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "token "+os.Getenv("GITHUB_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Do(req)
}

func formatComments(comments []string) string {
	var result string
	for _, comment := range comments {
		result += "- " + comment + "\n"
	}
	return result
}