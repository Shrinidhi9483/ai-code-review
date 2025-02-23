package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/Shrinidhi9483/ai-code-review/models"
	gh "github.com/Shrinidhi9483/ai-code-review/github"
	lm "github.com/Shrinidhi9483/ai-code-review/llama"
)

// type PullRequestEvent struct {
// 	Action      string      `json:"action"`
// 	PullRequest PullRequest `json:"pull_request"`
// }

// type Head struct {
// 	SHA string `json:"sha"`
// }

// type PullRequest struct {
// 	Number int  `json:"number"`
// 	Base   Base `json:"base"`
// 	Head   Head `json:"head"`
// }

// type Base struct {
// 	Repo Repo `json:"repo"`
// }

// type Repo struct {
// 	FullName string `json:"full_name"`
// }

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Recieved request on webhook")

	var event models.PullRequestEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Only process PRs when opened or updated
	if event.Action == "opened" || event.Action == "synchronize" || event.Action == "reopened" {
		repo := event.PullRequest.Base.Repo.FullName
		prNumber := event.PullRequest.Number

		log.Printf("fetching the commit ID")
		commitID, err := gh.GetLatestCommitID(repo, prNumber)
		if err != nil {
			log.Fatal("Error fetching commit ID:", err)
		}

		// Fetch changed files from PR
		log.Println("getting changes from github")
		files, err := gh.GetChangedFiles(repo, prNumber)
		if err != nil {
			log.Println("Error fetching PR files:", err)
			return
		}

		// Analyze files using LLaMA
		log.Println("analyzing the files with LLaMA")
		reviewComments := lm.AnalyzeCodeWithLLaMA(files)
		
		// Post review comments on GitHub
		log.Println("posting review comments")
		gh.PostReviewComments(repo, prNumber, commitID, reviewComments)
	}

	fmt.Fprintln(w, "Webhook received!")
}

func main() {
	// Load environment variables from .env file
	log.Printf("loading env cariables")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/webhook", webhookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("AI Code Review Bot running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
