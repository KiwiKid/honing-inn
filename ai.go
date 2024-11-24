package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// ErrorResponse represents the error structure returned by the API
type ErrorResponse struct {
	Error  Error
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
	config  PromptConfig
	prompt  string
}

type PerplexityResult struct {
	SuccessResults PerplexitySuccessResponse
	ErrorMessage   string
	config         PromptConfig
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
		"{address}":     home.CleanAddress,
		"{suburb}":      home.CleanSuburb,
		"{addressType}": addressType,
		"{topic}":       chatType.Name,
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
		Messages:          []Message{},
	}

}

type PromptConfig struct {
	Token             string
	DryRun            bool
	StartSystemPrompt string
	UserPrompt        string
	Replacements      map[string]string
	Messages          []Message
}

func callPerplexityAPI(config PromptConfig) (PerplexityResult, error) {
	url := "https://api.perplexity.ai/chat/completions"

	if config.Token == "" {
		return PerplexityResult{ErrorMessage: "API token not set"}, errors.New("API token not set")
	}

	runSettings := RunSettings{
		MaxTokens: 512,
	}

	// Construct the Perplexity API request body
	/*
		llama-3.1-sonar-small-128k-online (8B parameters)
		llama-3.1-sonar-large-128k-online (70B parameters)
		llama-3.1-sonar-huge-128k-online (405B parameters)
	*/
	reqBody := PerplexityRequest{
		Model:                  "llama-3.1-sonar-large-128k-online",
		Messages:               config.Messages,
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
		return PerplexityResult{ErrorMessage: "Failed to marshal request body", config: config}, err
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return PerplexityResult{ErrorMessage: "Failed to create request", config: config}, err
	}

	// Set headers for the API call
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Content-Type", "application/json")

	if config.DryRun {
		return PerplexityResult{
			config:         config,
			SuccessResults: PerplexitySuccessResponse{},
		}, nil
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return PerplexityResult{ErrorMessage: "Failed to call API", config: config}, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PerplexityResult{ErrorMessage: "Failed to read response body", config: config}, err
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Try to unmarshal into ErrorResponse
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return PerplexityResult{ErrorMessage: "Failed to unmarshal error response", config: config}, err
		}

		log.Printf("\n\n====Response: %+v\n\n", errResp)

		if len(errResp.Detail) == 0 {
			msg := fmt.Sprintf("API error %d - %v", resp.StatusCode, errResp)
			return PerplexityResult{ErrorMessage: msg, config: config}, fmt.Errorf(msg)
		}

		// Extract error details and return them
		return PerplexityResult{ErrorMessage: errResp.Detail[0].Msg, config: config}, fmt.Errorf("API error: %s", errResp.Detail[0].Msg)
	}

	// Try to unmarshal into SuccessResponse
	var successResp SuccessResponse
	if err := json.Unmarshal(body, &successResp); err != nil {
		return PerplexityResult{ErrorMessage: "Failed to unmarshal success response", config: config}, err
	}

	// log.Printf("\n\n====Response: %+v\n\n", body)

	// Return the content of the assistant's message
	if len(successResp.Choices) > 0 {
		return PerplexityResult{
			SuccessResults: PerplexitySuccessResponse{
				Choices: successResp.Choices,
				config:  config,
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

	msg := response.SuccessResults.Choices[len(response.SuccessResults.Choices)-1].Message

	rating := extractRating(msg.Content)

	return Chat{
		HomeID:        home.ID,
		ChatTypeTitle: chatType.Name,
		ChatType:      uint(chatType.ID),
		ThemeID:       uint(chatType.ThemeID),
		Results:       chatResults,
		Rating:        rating,
		Prompt:        response.SuccessResults.prompt,
	}
}

type FractalSearchParseResult struct {
	FractalSearchID uint
	DisplayName     string
	PointTypeName   string
	IsComplete      bool
	Points          []Point
}

func parseFractalSearchResult(response PerplexityResult, fs FractalSearch) ([]FractalSearchParseResult, error) {
	if len(response.SuccessResults.Choices) == 0 {
		return nil, errors.New("no choices returned in the response")
	}

	log.Printf("parseFractalSearchResult: %+v", response)
	// Extract the message from the last choice
	msg := response.SuccessResults.Choices[len(response.SuccessResults.Choices)-1].Message

	// Split message content into lines
	lines := strings.Split(msg.Content, "\n")
	var results []FractalSearchParseResult
	var currentResult *FractalSearchParseResult
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
			currentResult = &FractalSearchParseResult{
				FractalSearchID: fs.ID,
				DisplayName:     strings.TrimPrefix(line, "# "),
				PointTypeName:   "main",
				Points:          []Point{},
			}
		} else if strings.HasPrefix(line, "-") {

			var title string
			var description string
			fullLine := strings.TrimPrefix(line, "- ")
			lintParts := strings.Split(fullLine, " - ")
			if len(lintParts) != 2 {
				title = fullLine
			} else {
				title = lintParts[0]
				description = lintParts[1]
			}

			if currentResult != nil {
				point := &Point{
					ThemeID:     fs.ThemeID,
					Title:       title,
					Description: description,
					PointType:   currentResult.PointTypeName,
				}
				currentResult.Points = append(currentResult.Points, *point)
			}
		}
	}

	if currentResult != nil {
		results = append(results, *currentResult)
	}

	return results, nil
}

func progressFractalGeoSearch(db *gorm.DB, fs FractalSearch, existingMessages []Message, theme Theme, request FractalSearchRequest) (FractalSearch, PromptConfig, error) {

	log.Printf("progressFractalGeoSearch %d messages", len(existingMessages))
	promptConfig := buildGeoPromptConfig(fs, existingMessages, theme, request.dryRun)
	response, err := callPerplexityAPI(promptConfig)
	if err != nil {
		return fs, promptConfig, err
	}

	if request.dryRun {
		return fs, promptConfig, nil
	}

	log.Printf("❤️ progress Fractal search: response.SuccessResults.Choices:%d", len(response.SuccessResults.Choices))

	mewMsg := response.SuccessResults.Choices[len(response.SuccessResults.Choices)-1].Message

	newMsg, err := CreateMessage(db, Message{
		Role:            mewMsg.Role,
		Content:         mewMsg.Content,
		FractalSearchID: fs.ID,
	})

	log.Printf("Saved msg with length %d", len(newMsg.Content))

	searchResult, err := parseFractalSearchResult(response, fs)
	if err != nil {
		return fs, promptConfig, err
	}

	log.Printf("Saving Search Results %d", len(searchResult))

	var points []Point
	for _, result := range searchResult {
		newGroup, err := CreateFractalSearchResultGroup(db, FractalSearchResultGroup{
			FractalSearchID: fs.ID,
			DisplayName:     result.DisplayName,
			PointTypeName:   result.PointTypeName,
		})
		if err != nil {
			return fs, promptConfig, err
		}

		log.Printf("Saving Search Results - Points %d", len(result.Points))
		for _, point := range result.Points {
			point.ThemeID = fs.ThemeID
			point.FractalSearchID = fs.ID
			point.FractalSearchResultGroupID = newGroup.ID

			point, err := CreatePoint(db, point)
			if err != nil {
				return fs, promptConfig, err
			}
			points = append(points, *point)
		}
	}

	err = db.Save(&fs).Error
	if err != nil {
		return fs, promptConfig, err
	}
	return fs, promptConfig, nil

}
func placeFractalSearchPoints(db *gorm.DB, fs FractalSearchFull, osmClient *osmClient) []Point {
	var points []Point
	var totalLat, totalLng float64
	var processedPoints int

	// Helper function to safely parse coordinates
	parseCoord := func(coord string) float64 {
		val, _ := strconv.ParseFloat(coord, 64)
		return val
	}

	// Helper function to perform geocoding with a lookup string
	geocode := func(lookupStr string) ([]GeocodeResult, error) {
		geoRes, err := osmClient.GeocodeAddress(lookupStr)
		if err != nil || len(geoRes) == 0 {
			return nil, fmt.Errorf("geocoding failed: %w", err)
		}
		return geoRes, nil
	}

	// Helper function to calculate the distance between two coordinates
	calculateDistance := func(lat1, lng1, lat2, lng2 float64) float64 {
		return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lng1-lng2, 2))
	}

	// Average latitude and longitude for selecting closest option
	getClosestResult := func(geoRes []GeocodeResult, avgLat, avgLng float64) GeocodeResult {
		closest := geoRes[0]
		closestDistance := calculateDistance(avgLat, avgLng, parseCoord(geoRes[0].Lat), parseCoord(geoRes[0].Lon))

		for _, res := range geoRes[1:] {
			lat := parseCoord(res.Lat)
			lng := parseCoord(res.Lon)
			distance := calculateDistance(avgLat, avgLng, lat, lng)

			if distance < closestDistance {
				closest = res
				closestDistance = distance
			}
		}
		return closest
	}

	for _, point := range fs.Points {
		var lookupStrs = []string{
			point.Title + " " + point.Description,
			point.Title,
			point.Description,
		}

		var geoRes []GeocodeResult
		var err error

		// Try each lookup strategy in order
		for _, lookupStr := range lookupStrs {
			geoRes, err = geocode(lookupStr)
			if err == nil {
				break // Exit loop if successful
			}
		}

		// Handle geocoding failure
		if err != nil {
			point.WarningMessage = err.Error()
			points = append(points, point)
			continue
		}

		var selectedRes GeocodeResult

		if len(geoRes) > 1 && processedPoints > 0 { // Use closest to average if multiple results exist
			avgLat := totalLat / float64(processedPoints)
			avgLng := totalLng / float64(processedPoints)
			selectedRes = getClosestResult(geoRes, avgLat, avgLng)
		} else {
			selectedRes = geoRes[0]
		}

		// Parse and assign latitude and longitude
		latFloat := parseCoord(selectedRes.Lat)
		lngFloat := parseCoord(selectedRes.Lon)

		point.Lat = latFloat
		point.Lng = lngFloat
		point.WarningMessage = ""

		// Update total lat/lng and processed points for averaging
		totalLat += latFloat
		totalLng += lngFloat
		processedPoints++

		points = append(points, point)
	}

	return points
}

func buildGeoPromptMessages(messages []Message, query string, theme Theme) []Message {
	startSystemPrompt := getStartGeoSearchPrompt(theme)

	if len(messages) == 0 {
		messages = append(messages, Message{
			Role:    "system",
			Content: startSystemPrompt,
		})
	}

	messages = append(messages, Message{
		Role:    "user",
		Content: query,
	})
	return messages
}

func buildGeoPromptConfig(fs FractalSearch, messages []Message, theme Theme, dryRun bool) PromptConfig {

	replacements := getGeoReplacements(fs)

	var prompt string = fs.Query
	for key, value := range replacements {
		prompt = strings.Replace(prompt, key, value, -1)
		log.Printf("After replacement (%s|%s): %s", key, value, prompt)
	}

	promptMessages := buildGeoPromptMessages(messages, fs.Query, theme)
	for _, msg := range promptMessages {
		log.Printf("%+v", msg)
	}
	promptConfig := PromptConfig{
		StartSystemPrompt: "",
		UserPrompt:        fs.Query,
		Replacements:      replacements,
		Messages:          promptMessages,
		Token:             os.Getenv("PERPLEXITY_API_TOKEN"),
		DryRun:            dryRun,
	}

	log.Printf("PromptConfig: %+v", promptConfig)

	return promptConfig
}

func getStartGeoSearchPrompt(theme Theme) string {
	if len(theme.StartGeoSystemPrompt) > 0 {
		return theme.StartGeoSystemPrompt
	} else {
		return `You are designed to take a user's query, search for places matching the query, and return places in a structured format. 
Follow these steps:  
1. Categorize the results into logical groups, if applicable.  
2. Use headers ('#' for main categories) to organize the list of locations  
3. Only include the title and location of the place
4. Only Present items short title and location, as bullet points, with the item name and the specific street address or location:
5. Dont include any additional information or descriptions.



Like this:
# [category name]
- [item name] - [location]



Query: Great Walks in New Zealand
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
		

	

Query: Parks in Christchurch by size
Result: 
# Big Parks
- Hagley Park  
- Bottle Lake Forest Park
- Halswell Quarry Park
- The Groynes**  
- Travis Wetland Nature Heritage Park

# Medium Parks
- Victoria Park  
- Spencer Park  
- Abberley Park  
- Risingholme Park

## Small Parks  
- Woodham Park
- Beverley Park
- Linwood Park
		
Be precise, clear, and consistent in your responses.
`
	}
}
