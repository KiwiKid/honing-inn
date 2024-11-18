package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"
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
	Rating  int
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

func getReplacements(home Home, addressType string, chatType ChatType) map[string]string {
	return map[string]string{
		"{address}": home.CleanAddress,
		"{suburb}":  home.CleanSuburb,
		"{topic}":   chatType.Name,
	}
}

func getGeoReplacements(fs FractalSearch) map[string]string {
	return map[string]string{
		"{location}": fs.DisplayName,
	}
}

func extractRating(message string) int {
	// Define a regular expression to match "Rating:" followed by a number
	re := regexp.MustCompile(`Rating:\s*(\d+)`)

	// Find the rating in the message
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		// matches[1] contains the captured number
		ratingStr := strings.TrimSpace(matches[1])
		rating, err := strconv.Atoi(ratingStr)
		if err != nil {
			log.Printf("Failed to convert rating to integer: %v", err)
			return -1
		}
		return rating
	}

	// Return -1 if no rating is found
	return -1
}

func getStartSystemPrompt(theme Theme, chatType ChatType) string {
	if chatType.StartSystemPromptOverride != "" {
		return chatType.StartSystemPromptOverride
	} else if theme.StartSystemPrompt != "" {
		return theme.StartSystemPrompt
	} else {
		return `You are speaking to a expert assistant in home researching specific aspects of home buying researching {topic}. Be concise and truthfully answer questions in the context of purchasing a home. Always finish your answers. Give the address a relative rating (1 to 3 with 3 being the best score) compared to other addresses in the city like this:

		This property is located in a relatively quiet area. However, it is situated near major roads like Riccarton Road and Blenheim Road, which are significant thoroughfares in the city. These roads can generate some traffic noise and activity at peak hours.
		Given the proximity to these roads, I would rate the traffic noise and activity around this property is relatively high but the area is generally residential and not directly adjacent to high-traffic zones like highways or major commercial centers.
		For a more precise assessment, it's important to note that Riccarton Road and Blenheim Road are local roads with moderate traffic levels, especially during peak hours. However, they do not compare to the high-traffic volumes found on major highways or central business districts in Christchurch.
	
		The overall environment around 7 Middleton Road is relatively peaceful and suitable for residential living.
	
		Rating: 3
		`
	}
}

func buildPromptConfig(home Home, chatType ChatType, theme Theme) PromptConfig {
	replacements := getReplacements(home, chatType.AddressType, chatType)
	startSystemPrompt := getStartSystemPrompt(theme, chatType)

	return PromptConfig{
		Token:             os.Getenv("PERPLEXITY_API_TOKEN"),
		StartSystemPrompt: startSystemPrompt,
		UserPrompt:        chatType.Prompt,
		Replacements:      replacements,
		ExistingMessages:  []Message{},
	}

}

type PromptConfig struct {
	Token             string
	StartSystemPrompt string
	UserPrompt        string
	Replacements      map[string]string
	ExistingMessages  []Message
}

func callPerplexityAPI(config PromptConfig) (PerplexityResult, error) {
	url := "https://api.perplexity.ai/chat/completions"

	if config.Token == "" {
		return PerplexityResult{ErrorMessage: "API token not set"}, errors.New("API token not set")
	}

	runSettings := RunSettings{
		MaxTokens: 512,
	}

	var prompt string = config.UserPrompt
	for key, value := range config.Replacements {
		prompt = strings.Replace(prompt, key, value, -1)
		log.Printf("After replacement (%s|%s): %s", key, value, prompt)
	}

	messages := []Message{}
	if len(config.ExistingMessages) > 0 {
		messages = append(config.ExistingMessages, Message{Role: "user", Content: prompt})
	} else {
		messages = []Message{
			{Role: "system", Content: config.StartSystemPrompt},
			{Role: "user", Content: prompt},
		}
	}

	// Construct the Perplexity API request body
	/*
		llama-3.1-sonar-small-128k-online (8B parameters)
		llama-3.1-sonar-large-128k-online (70B parameters)
		llama-3.1-sonar-huge-128k-online (405B parameters)
	*/
	reqBody := PerplexityRequest{
		Model:                  "llama-3.1-sonar-small-128k-online",
		Messages:               messages,
		MaxTokens:              runSettings.MaxTokens,
		Temperature:            0.2,
		TopP:                   0.9,
		ReturnCitations:        true,
		SearchDomainFilter:     []string{}, // []string{"perplexity.ai"},
		ReturnImages:           false,
		ReturnRelatedQuestions: true,
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

	// log.Printf("\n\n====Response: %+v\n\n", body)

	rating := extractRating(successResp.Choices[0].Message.Content)

	log.Printf("extractRating returned: %d", rating)

	// Return the content of the assistant's message
	if len(successResp.Choices) > 0 {
		return PerplexityResult{
			SuccessResults: PerplexitySuccessResponse{
				Choices: successResp.Choices,
				Rating:  rating,
				prompt:  prompt,
			},
		}, nil
	}

	return PerplexityResult{ErrorMessage: "No choices returned in the response"}, fmt.Errorf("no choices returned")
}

func buildChat(response PerplexityResult, home Home, chatType ChatType) Chat {

	var chatResults []ChatResult
	for _, cr := range response.SuccessResults.Choices {
		chatResults = append(chatResults, ChatResult{
			Result: cr.Message.Content,
			Role:   cr.Message.Role,
		})
	}

	return Chat{
		HomeID:        home.ID,
		ChatTypeTitle: chatType.Name,
		ChatType:      uint(chatType.ID),
		ThemeID:       uint(chatType.ThemeID),
		Results:       chatResults,
		Rating:        response.SuccessResults.Rating,
		Prompt:        response.SuccessResults.prompt,
	}
}
func parseFractalSearchResult(response PerplexityResult, fs FractalSearch) ([]FractalSearchResult, bool, error) {
	if len(response.SuccessResults.Choices) == 0 {
		return nil, true, errors.New("no choices returned in the response")
	}

	// Extract the message from the last choice
	msg := response.SuccessResults.Choices[len(response.SuccessResults.Choices)-1].Message
	fs.Messages = append(fs.Messages, msg)

	// Split message content into lines
	lines := strings.Split(msg.Content, "\n")
	var results []FractalSearchResult
	var currentResult *FractalSearchResult
	var isComplete bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle the different line types
		if strings.HasPrefix(line, "#") {
			// New section (reset the current result)
			if currentResult != nil {
				results = append(results, *currentResult)
			}
			currentResult = &FractalSearchResult{
				FractalSearchID: fs.ID,
				DisplayName:     strings.TrimPrefix(line, "# "),
				Points:          []string{},
			}
		} else if strings.HasPrefix(line, "-") {
			// List item (add to Points)
			if currentResult != nil {
				currentResult.Points = append(currentResult.Points, strings.TrimPrefix(line, "- "))
			}
		} else if line == "LIST IS COMPLETE" {
			// Completion indicator
			isComplete = true
		}
	}

	if currentResult != nil {
		results = append(results, *currentResult)
	}

	return results, isComplete, nil
}

func progressFractalGeoSearch(db *gorm.DB, fs FractalSearch, theme Theme) (FractalSearch, error) {

	promptConfig := buildGeoPromptConfig(fs, theme)
	response, err := callPerplexityAPI(promptConfig)
	if err != nil {
		return fs, err
	}

	log.Printf("progress Fractal search: %+v", response)
	msg := response.SuccessResults.Choices[len(response.SuccessResults.Choices)-1].Message
	fs.Messages = append(fs.Messages, msg)

	searchResult, isComplete, err := parseFractalSearchResult(response, fs)
	if err != nil {
		return fs, err
	}

	var points []Point
	for _, result := range searchResult {
		for _, point := range result.Points {
			point, err := CreatePoint(db, Point{
				ThemeID:         theme.ID,
				Title:           point,
				PointType:       result.PointTypeName,
				FractalSearchID: fs.ID,
			})
			if err != nil {
				return fs, err
			}
			points = append(points, *point)
		}
	}

	if isComplete {
		fs.Status = "complete"
	} else {
		fs.Status = "in-progress"
	}

	return fs, nil

}

func buildGeoPromptConfig(fs FractalSearch, theme Theme) PromptConfig {

	replacements := getGeoReplacements(fs)
	startSystemPrompt := getStartGeoSearchPrompt(theme)

	promptConfig := PromptConfig{
		StartSystemPrompt: startSystemPrompt,
		UserPrompt:        fs.Query,
		Replacements:      replacements,
		ExistingMessages:  fs.Messages,
		Token:             os.Getenv("PERPLEXITY_API_TOKEN"),
	}

	log.Printf("PromptConfig: %+v", promptConfig)

	return promptConfig
}

func getStartGeoSearchPrompt(theme Theme) string {
	if len(theme.StartGeoSystemPrompt) > 0 {
		return theme.StartGeoSystemPrompt
	} else {
		return `You are designed to take a user's query, search for places or items matching the query, and return results in a structured format. Follow these steps:  
		1. Categorize the results into logical groups, if applicable.  
		2. Use headers ('#' for main categories, '##' for subcategories) to organize the information.  
		3. Present items as bullet points, each on a single line, with relevant details, where applicable. Include details on the same line as the item.
		4. At the end of your response, indicate whether the list is complete or incomplete by adding either 'LIST IS COMPLETE' or ''LIST IS INCOMPLETE'.  
		
**Example 1:**  
# Great Walks  
	- Routeburn Track  
	- Milford Track  
	- Kepler Track  
	- Heaphy Track  
	- Abel Tasman Coast Track  
	- Tongariro Northern Circuit  
	- Whanganui Journey  
	- Lake Waikaremoana  
	- Paparoa Track  
	- Rakiura Track  
	- Hump Ridge Track  
	- Hollyford Track  
	- Te Araroa Trail  
		
LIST IS COMPLETE  
		
**Example 2:**
Query: Parks in Christchurch by size
Result: 
# Parks in Christchurch by size
## **Big Parks**  
- **Hagley Park**  
- **Bottle Lake Forest Park**  
- **Halswell Quarry Park**  
- **The Groynes**  
- **Travis Wetland Nature Heritage Park**

## **Medium Parks**  
- **Victoria Park**  
- **Spencer Park**  
- **Abberley Park**  
- **Risingholme Park**

## **Small Parks**  
- **Woodham Park**  
- **Local neighborhood reserves (e.g., Beverley Park, Linwood Park)**  

---

This grouping reflects relative size and significance based on their amenities and land area. Would you like more detailed descriptions or additional parks?
		LIST IS INCOMPLETE  
		
		Be precise, clear, and consistent in your responses.
`
	}
}
