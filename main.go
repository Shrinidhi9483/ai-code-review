package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type PullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
}

type PullRequest struct {
	Number int  `json:"number"`
	Base   Base `json:"base"`
}

type Base struct {
	Repo Repo `json:"repo"`
}

type Repo struct {
	FullName string `json:"full_name"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	var event PullRequestEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Only process PRs when opened or updated
	if event.Action == "opened" || event.Action == "synchronize" {
		repo := event.PullRequest.Base.Repo.FullName
		prNumber := event.PullRequest.Number

		// Fetch changed files from PR
		files, err := GetChangedFiles(repo, prNumber)
		if err != nil {
			log.Println("Error fetching PR files:", err)
			return
		}

		// Analyze files using LLaMA
		reviewComments := AnalyzeCodeWithLLaMA(files)

		// Post review comments on GitHub
		PostReviewComments(repo, prNumber, reviewComments)
	}

	fmt.Fprintln(w, "Webhook received!")
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("AI Code Review Bot running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}