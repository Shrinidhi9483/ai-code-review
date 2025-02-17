package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type RequestBody struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Stream  bool   `json:"stream"`
}

type ResponseBody struct {
	Response string `json:"response"`
}


func AnalyzeCodeWithLLaMA(changedFiles []string) []string {
	var reviewComments []string
	for _, codeSnippet := range changedFiles {
		reviewComment, err :=  getLLaMAReview(codeSnippet)
		if err != nil {
			log.Printf("failed to get review from llama: %v",  err)
			continue
		}
		reviewComments = append(reviewComments,  reviewComment)	
	}
	return reviewComments
}

// func getLLaMAReview(codeSnippet string) string {
// 	cmd := exec.Command("./llama.cpp/main", "-m", "llama/ggml-model-q4_0.bin", "-p", "Review this code:\n"+codeSnippet)

// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Run()

// 	// Extract relevant output
// 	reviewOutput := out.String()
// 	reviewLines := strings.Split(reviewOutput, "\n")

// 	// Return first few lines as the review
// 	if len(reviewLines) > 5 {
// 		return strings.Join(reviewLines[:5], "\n")
// 	}
// 	return reviewOutput
// }

func getLLaMAReview(codeSnippet string) (string, error) {
	url := "http://localhost:11434/api/generate"

	// Create request body
	requestBody, _ := json.Marshal(RequestBody{
		Model:  "llama3.2",
		Prompt: codeSnippet,
		Stream: false,
	})
	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return  "", err
	}
	defer resp.Body.Close()

	// Read response
	body, _ := ioutil.ReadAll(resp.Body)

	// Parse JSON response
	var result ResponseBody
	json.Unmarshal(body, &result)
	
	return result.Response, nil
}