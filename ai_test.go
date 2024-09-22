package main

import (
	"log"
	"strings"
	"testing"
)

func TestCleanAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		give string
		want string
	}{
		{
			name: "1 No replacements",
			give: "123 Main Street",
			want: "123 Main Street",
		},
		{
			name: "2 Remove For sale",
			give: "For sale | 123 Main Street",
			want: " 123 Main Street",
		},
		{
			name: "3 Remove homes.co.nz",
			give: "123 Main Street homes.co.nz",
			want: "123 Main Street ",
		},
		{
			name: "4 Remove both For sale and homes.co.nz",
			give: "For sale | 123 Main Street homes.co.nz",
			want: " 123 Main Street ",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := cleanAddress(tt.give)
			if got != tt.want {
				t.Errorf("cleanAddress(%q) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}

func TestCallPerplexityAPI(t *testing.T) {
	// Sample prompt and token (replace with a valid token)
	prompt := "Tell me about Go programming."
	token := "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" // Replace with your actual API token

	// Define replacements
	replacements := map[string]string{
		"Go": "Golang",
	}

	config := PromptConfig{
		Token:        token,
		Prompt:       prompt,
		Replacements: replacements,
	}

	// Call the API
	result, err := callPerplexityAPI(config)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Check if the result has choices
	if len(result.SuccessResults.Choices) == 0 {
		t.Fatalf("Expected choices in the result, but got none")
	}

	// Log the result to see what was returned
	log.Printf("API response: %+v", result)

	// Example check for content in the first choice
	expectedSubstring := "Golang"
	if !contains(result.SuccessResults.Choices[0].Message.Content, expectedSubstring) {
		t.Errorf("Expected result to contain '%s', but got: %s", expectedSubstring, result.SuccessResults.Choices[0].Message.Content)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
