package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const perplexityAPIURL = "https://api.perplexity.ai/chat/completions"

type OpenAIRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func chatHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Extract query from request
			query := r.URL.Query().Get("q")
			if query == "" {
				http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
				return
			}

			// Create the request to Perplexity API
			openAIReq := OpenAIRequest{
				Model:  "gpt-4", // Assuming model similar to OpenAI
				Prompt: query,
				Stream: true, // Stream response
			}

			// Make the API call
			reqBody, _ := json.Marshal(openAIReq)
			req, err := http.NewRequest("POST", perplexityAPIURL, strings.NewReader(string(reqBody)))
			if err != nil {
				http.Error(w, "Failed to create API request", http.StatusInternalServerError)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			// Add API key if required
			// req.Header.Set("Authorization", "Bearer YOUR_API_KEY")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, "OLDOLDOLD Failed to call Perplexity API", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				http.Error(w, fmt.Sprintf("Perplexity API returned status: %s", resp.Status), http.StatusInternalServerError)
				return
			}

			// Stream the response to the client
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				http.Error(w, "Failed to stream response", http.StatusInternalServerError)
				return
			}
		case http.MethodPost:
			// Parse the form values from the request
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}

			continueChatID := r.FormValue("ChatID")
			if len(continueChatID) > 0 {
				warn := warning("continueChatID not yet implemented")
				warn.Render(GetContext(r), w)
				return
			}

			homeID := r.FormValue("HomeID")
			chatTypeID := r.FormValue("chatTypeID")
			themeID := r.FormValue("ThemeID")

			if homeID == "" || chatTypeID == "" || themeID == "" {
				http.Error(w, "Missing required form values", http.StatusBadRequest)
				return
			}

			homeIDUint, err := strconv.ParseUint(homeID, 10, 32)
			if err != nil {
				http.Error(w, "Invalid HomeID", http.StatusBadRequest)
				return
			}
			chatTypeIDUint, err := strconv.ParseUint(chatTypeID, 10, 32)
			if err != nil {
				http.Error(w, "Invalid chatTypeID", http.StatusBadRequest)
				return
			}
			themeIDUint, err := strconv.ParseUint(themeID, 10, 32)
			if err != nil {
				http.Error(w, "Invalid ThemeID", http.StatusBadRequest)
				return
			}

			chatType, err := GetChatType(db, uint(chatTypeIDUint))
			if err != nil {
				http.Error(w, "Failed to get chat type", http.StatusInternalServerError)
				return
			}

			var home *Home
			if homeIDUint > 0 {
				home, err = GetHome(db, uint(homeIDUint))
				if err != nil {
					warn := warning("Failed to get home")
					warn.Render(GetContext(r), w)
					return
				}
			}

			if home.Title == "" {
				warn := warning("Home title is empty - try adding a title or enter the url")
				warn.Render(GetContext(r), w)
				return
			}

			// Call Perplexity API based on chatTypeID
			replacements := getReplacements(*home)

			config := PromptConfig{
				Token:        os.Getenv("PERPLEXITY_API_TOKEN"),
				Prompt:       chatType.Prompt,
				Replacements: replacements,
			}

			response, err := callPerplexityAPI(config)
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to call Perplexity API: %v", err))
				warning.Render(GetContext(r), w)
				return
			}

			if response.ErrorMessage != "" {
				warning := warning(fmt.Sprintf("Perplexity API returned error: %v", response.ErrorMessage))
				warning.Render(GetContext(r), w)
				return
			}

			var chatResults []ChatResult
			for _, cr := range response.SuccessResults.Choices {
				chatResults = append(chatResults, ChatResult{
					Result: cr.Message.Content,
					Role:   cr.Message.Role,
				})
			}

			// Process the chat submission (this is an example, modify as per your logic)
			newChat := Chat{
				HomeID:   uint(homeIDUint),
				ChatType: uint(chatTypeIDUint),
				ThemeID:  uint(themeIDUint),
				Results:  chatResults,
				Prompt:   response.SuccessResults.prompt,
			}

			// Save chat to the database (optional)
			if err := db.Create(&newChat).Error; err != nil {
				warning := warning(fmt.Sprintf("Failed to save chat %v", err))
				warning.Render(GetContext(r), w)
				return
			}

			chatRes := chat(newChat)
			chatRes.Render(GetContext(r), w)
			return
		case http.MethodDelete:
			// Delete chat
			chatIdStr := chi.URLParam(r, "chatId")
			if chatIdStr == "" {
				warning := warning("chatId ID not provided")
				warning.Render(GetContext(r), w)
				return
			}

			id, err := strconv.Atoi(chatIdStr)
			if err != nil {
				warning := warning("Invalid chatId ID")
				warning.Render(GetContext(r), w)
				return
			}

			chat, dErr := DeleteChat(db, uint(id))
			if dErr != nil {
				warning := warning("Failed to delete factor")
				warning.Render(GetContext(r), w)
				return
			}

			reload := loadChat(chat.ThemeID, chat.HomeID, chat.ChatType)
			reload.Render(GetContext(r), w)
			return
		}
	}
}
