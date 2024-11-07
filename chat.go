package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

func callAndSavePerplexityAPI(db *gorm.DB, home Home, chatType ChatType, theme Theme) (*Chat, error) {

	log.Printf("Calling Perplexity API for home %v and chat type %v", home.ID, chatType.ID)

	config := buildPromptConfig(home, chatType, theme)

	log.Printf("Using Config %s", config.UserPrompt)
	log.Printf("Using Config %s", config.StartSystemPrompt)
	log.Printf("Using Config %v", config.Replacements)
	response, err := callPerplexityAPI(config)
	if err != nil {
		return nil, err
	}

	if response.ErrorMessage != "" {
		return nil, fmt.Errorf("Perplexity API returned error: %v", response.ErrorMessage)
	}

	newChat := buildChat(response, home, chatType)

	if err := db.Create(&newChat).Error; err != nil {
		return nil, fmt.Errorf("Failed to save chat %v", err)
	}

	return &newChat, nil
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
			all := r.FormValue("All") == "true"
			themeID := r.FormValue("ThemeID")

			if homeID == "" || (chatTypeID == "" && !all) || themeID == "" {
				http.Error(w, "Missing required form values", http.StatusBadRequest)
				return
			}

			homeIDUint, err := strconv.ParseUint(homeID, 10, 32)
			if err != nil {
				http.Error(w, "Invalid HomeID", http.StatusBadRequest)
				return
			}

			themeIDUint, err := strconv.ParseUint(themeID, 10, 32)
			if err != nil {
				http.Error(w, "Invalid ThemeID", http.StatusBadRequest)
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

			if home.CleanAddress == "" {
				warn := warning("Home Address is empty - try adding a URL or enter it manually via home edit")
				warn.Render(GetContext(r), w)
				return
			}

			if home.CleanAddress == "" {
				warn := warning("Home title is empty - try adding a URL or enter it manually via home edit")
				warn.Render(GetContext(r), w)
				return
			}

			if all {

				chatTypes, err := GetChatTypes(db, uint(themeIDUint))
				if err != nil {
					waring := warning("Failed to get chat types for all processing")
					waring.Render(GetContext(r), w)
					return
				}

				theme := GetActiveTheme(db, uint(themeIDUint))

				newChats := make([]Chat, 0)
				for _, chatType := range chatTypes {

					newChat, err := callAndSavePerplexityAPI(db, *home, chatType, theme)
					if err != nil {
						warning := warning(fmt.Sprintf("Failed to call Perplexity API: %v", err))
						warning.Render(GetContext(r), w)
						return
					}

					newChats = append(newChats, *newChat)

				}
				ratingListView := chatRatingListView(newChats)
				ratingListView.Render(GetContext(r), w)
				return
			}

			chatTypeIDUint, err := strconv.ParseUint(chatTypeID, 10, 32)
			if err != nil {
				http.Error(w, "Invalid chatTypeID", http.StatusBadRequest)
				return
			}

			chatType, err := GetChatType(db, uint(chatTypeIDUint))
			if err != nil {
				http.Error(w, "Failed to get chat type", http.StatusInternalServerError)
				return
			}

			theme := GetActiveTheme(db, uint(themeIDUint))

			newChat, err := callAndSavePerplexityAPI(db, *home, *chatType, theme)
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to call Perplexity API: %v", err))
				warning.Render(GetContext(r), w)
				return
			}

			chatRes := chat(*newChat)
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

			log.Printf("DELETE chat %s!", chatIdStr)

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

			msg := fmt.Sprintf("Deleted chat %s (%d)", chat.ChatTypeTitle, chat.ID)

			log.Printf("loadEmptyChat(%v, %v, %v)", chat.ThemeID, chat.HomeID, 0)
			reload := loadEmptyChat(chat.ThemeID, chat.HomeID, msg)
			reload.Render(GetContext(r), w)
			return
		}
	}
}
