package llama

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"strconv"
	"github.com/tidwall/gjson"

	"github.com/Shrinidhi9483/ai-code-review/models"
)

func addLineNumbersToCode(code string) (string) {
	var result string 
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		result += fmt.Sprintf("%v	%v\n", i, line)
	}
	return result
}

func AnalyzeCodeWithLLaMA(changedFiles []models.PrFileMetadata) []models.ReviewComments {
	var reviewComments []models.ReviewComments

	for _, file := range changedFiles {
		codeWithLineNumbers := addLineNumbersToCode(file.Patch)
		reviewComment, err := getLLaMAReview(codeWithLineNumbers)
		if err != nil {
			log.Printf("failed to get review from llama: %v", err)
			continue
		}

		reviewComment.ForEach(func(key, value gjson.Result) bool {
			var reviewComment models.ReviewComments
			reviewComment.Path = file.Filename
			lineInt, err := strconv.Atoi(key.String())
			if err == nil {	
				// Some jitter is involved in the 1st line of the file when it is fetched from github, so skipping the first line (0th position)
				if lineInt != 0 {
					// Sometimes the AI model just outputs empty string or comments as a review, its a strange behavior which may not happen with all models, regardless its handled here.
					if value.String() != "" && value.String() != "//" {
						reviewComment.Position = lineInt
						reviewComment.Body = value.String()
						reviewComments = append(reviewComments,  reviewComment)
					}
				}
			} else {
				log.Printf("failed to conver line number to int: %v", err)
			}	
			return true
		})
	}
	return reviewComments
}

func getLLaMAReview(codeSnippet string) (gjson.Result, error) {
	var response models.ResponseBody
	var jsonResult gjson.Result
	
	url := "http://localhost:11434/api/generate"
	prompt :=  fmt.Sprintf("Review following code changes and suggest best practice to improvise the code. Only provide review comment if there is something that can be improvised, otherwise ignore that line and don't include it in response.  Line numbers is provided at beginning of each line of code, use them for providing review comment at relevant line numbers. Provide the output in JSON format, where each key is a line number in the provided code and the corresponding value is the review comment. Do not output anything else except the json in response\n %v", codeSnippet)

	// Create request body
	requestBody, _ := json.Marshal(models.RequestBody{
		Model: "llama3.2",
		Prompt: prompt,
		Stream: false,
	})
	
	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return jsonResult, err
	}
	defer resp.Body.Close()

	// Read response
	body, _ := ioutil.ReadAll(resp.Body)

	// Parse JSON response
	json.Unmarshal(body, &response)

	re := regexp.MustCompile(`\{[\s\S]*\}`)
	match := re.FindString(response.Response)
	if match == "" {
		return jsonResult, fmt.Errorf("failed to parse json from the response: %v",  response.Response)
	} 

	jsonResult = gjson.Parse(match)
	return jsonResult, nil
}
