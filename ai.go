package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type PerplexityRequest struct {
	Model                  string    `json:"model"`
	Messages               []Message `json:"messages"`
	MaxTokens              int       `json:"max_tokens,omitempty"`
	Temperature            float64   `json:"temperature"`
	TopP                   float64   `json:"top_p"`
	ReturnCitations        bool      `json:"return_citations"`
	SearchDomainFilter     []string  `json:"search_domain_filter"`
	ReturnImages           bool      `json:"return_images"`
	ReturnRelatedQuestions bool      `json:"return_related_questions"`
	SearchRecencyFilter    string    `json:"search_recency_filter"`
	TopK                   int       `json:"top_k"`
	Stream                 bool      `json:"stream"`
	PresencePenalty        float64   `json:"presence_penalty"`
	FrequencyPenalty       float64   `json:"frequency_penalty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ErrorResponse represents the error structure returned by the API
type ErrorResponse struct {
	Detail []ErrorDetail `json:"detail"`
}

type ErrorDetail struct {
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
}

// SuccessResponse represents the structure of a successful API call
type SuccessResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type PerplexitySuccessResponse struct {
	Choices []Choice
	prompt  string
}

type PerplexityResult struct {
	SuccessResults PerplexitySuccessResponse
	ErrorMessage   string
}

type RunSettings struct {
	MaxTokens int
}

func cleanAddress(address string) string {
	address = strings.Replace(address, "For sale |", "", -1)
	address = strings.Replace(address, "homes.co.nz", "", -1)
	return address
}

func getReplacements(home Home) map[string]string {
	return map[string]string{
		"{address}": cleanAddress(home.Title),
	}
}

type PromptConfig struct {
	Token        string
	Prompt       string
	Replacements map[string]string
}

func callPerplexityAPI(config PromptConfig) (PerplexityResult, error) {
	url := "https://api.perplexity.ai/chat/completions"

	if config.Token == "" {
		return PerplexityResult{ErrorMessage: "API token not set"}, errors.New("API token not set")
	}

	runSettings := RunSettings{
		MaxTokens: 512, //512,
	}

	var prompt string = config.Prompt
	for key, value := range config.Replacements {
		prompt = strings.Replace(prompt, key, value, -1)
		log.Printf("After replacement (%s|%s): %s", key, value, prompt)
	}

	// Construct the Perplexity API request body
	reqBody := PerplexityRequest{
		Model: "llama-3.1-sonar-small-128k-online",
		Messages: []Message{
			{Role: "system", Content: "Be precise and concise."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:              runSettings.MaxTokens,
		Temperature:            0.2,
		TopP:                   0.9,
		ReturnCitations:        true,
		SearchDomainFilter:     []string{"perplexity.ai"},
		ReturnImages:           false,
		ReturnRelatedQuestions: false,
		SearchRecencyFilter:    "month",
		TopK:                   0,
		Stream:                 false,
		PresencePenalty:        0,
		FrequencyPenalty:       1,
	}

	// Convert the struct to JSON
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return PerplexityResult{ErrorMessage: "Failed to marshal request body"}, err
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return PerplexityResult{ErrorMessage: "Failed to create request"}, err
	}

	// Set headers for the API call
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return PerplexityResult{ErrorMessage: "Failed to call API"}, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PerplexityResult{ErrorMessage: "Failed to read response body"}, err
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Try to unmarshal into ErrorResponse
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return PerplexityResult{ErrorMessage: "Failed to unmarshal error response"}, err
		}

		log.Printf("\n\n====Response: %+v\n\n", body)

		// Extract error details and return them
		return PerplexityResult{ErrorMessage: errResp.Detail[0].Msg}, fmt.Errorf("API error: %s", errResp.Detail[0].Msg)
	}

	// Try to unmarshal into SuccessResponse
	var successResp SuccessResponse
	if err := json.Unmarshal(body, &successResp); err != nil {
		return PerplexityResult{ErrorMessage: "Failed to unmarshal success response"}, err
	}

	// Return the content of the assistant's message
	if len(successResp.Choices) > 0 {
		return PerplexityResult{
			SuccessResults: PerplexitySuccessResponse{
				Choices: successResp.Choices,
				prompt:  prompt,
			},
		}, nil
	}

	return PerplexityResult{ErrorMessage: "No choices returned in the response"}, fmt.Errorf("no choices returned")
}
