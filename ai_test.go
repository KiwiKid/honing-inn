package main

import (
	"errors"
	//	"os"
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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestExtractRating(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    int
	}{
		{
			name:    "Valid rating in message",
			message: "User's feedback. Rating: 5",
			want:    5,
		},
		{
			name:    "Valid rating with large number",
			message: "User's feedback. Rating: 123",
			want:    123,
		},
		{
			name:    "Invalid rating (non-numeric)",
			message: "User's feedback. Rating: five",
			want:    -1,
		},
		{
			name:    "No rating in message",
			message: "User's feedback.",
			want:    -1,
		},
		{
			name:    "Rating keyword present, no number",
			message: "User's feedback. Rating:",
			want:    -1,
		},
		{
			name: "Rating at the end of message",
			message: `User liked the product.
			**Rating: 8**`,
			want: 8,
		},
		{
			name:    "Multiple Ratings, takes first",
			message: "Rating: 5 and Rating: 10",
			want:    5,
		},
		{
			name:    "Rating with whitespace",
			message: "Rating:    9",
			want:    9,
		},
		{
			name:    "Negative rating",
			message: "Rating: -2",
			want:    -1,
		},
		{
			name:    "No space after 'Rating:'",
			message: "Rating:10",
			want:    10,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractRating(tt.message)
			if got != tt.want {
				t.Errorf("extractRating(%q) = %d, want %d", tt.message, got, tt.want)
			}
		})
	}
}

/*
	func TestBuildGeoPromptConfig(t *testing.T) {
		// Mock environment variable
		os.Setenv("PERPLEXITY_API_TOKEN", "test-token")
		defer os.Unsetenv("PERPLEXITY_API_TOKEN")

		mockReplacements := map[string]string{
			"replacement":        "REPLACED VALUE",
			"anotherReplacement": "ANOTHER REPLACED VALUE",
		}

		tests := []struct {
			name      string
			fs        FractalSearch
			theme     Theme
			expected  PromptConfig
			expectLog bool
		}{
			{
				name: "Basic Functionality",
				fs: FractalSearch{
					Query: "example query",
				},
				theme: Theme{
					StartGeoSystemPrompt: "you are a test",
				},
				expected: PromptConfig{
					StartSystemPrompt: "you are a test",
					UserPrompt:        "find all great walks in {replacement}",
					Replacements:      mockReplacements,
					ExistingMessages: []Message{{
						Role:    "user",
						Content: "Woah some content",
					},
					},
					Token: "test-token",
				},
				expectLog: true,
			},
			{
				name: "Empty Query",
				fs: FractalSearch{
					Query: "",
				},
				theme: Theme{
					StartGeoSystemPrompt: "you are a test",
				},
				expected: PromptConfig{
					StartSystemPrompt: "you are a test",
					UserPrompt:        "find all great walks in {replacement}",
					Replacements:      mockReplacements,
					ExistingMessages:  []Message{},
					Token:             "test-token",
				},
				expectLog: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				//var loggedOutput strings.Builder
				//log.SetOutput(&loggedOutput)

				result := buildGeoPromptConfig(tt.fs, []Message{}, tt.theme)

				if result.StartSystemPrompt != tt.expected.StartSystemPrompt {
					t.Errorf("Expected %+v, got %+v", tt.expected.StartSystemPrompt, result.StartSystemPrompt)
				}

				if len(result.ExistingMessages) != len(tt.expected.ExistingMessages) {
					t.Errorf("Expected %+v, got %+v", len(tt.expected.ExistingMessages), len(result.ExistingMessages))
				}

				for key, value := range tt.expected.ExistingMessages {
					if result.ExistingMessages[key].Content != value.Content {
						t.Errorf("Expected %+v, got %+v", value.Content, result.ExistingMessages[key].Content)
					}
				}

				//if tt.expectLog && !strings.Contains(loggedOutput.String(), "PromptConfig:") {
				//	t.Errorf("Expected log output not found")
				//}
			})
		}
	}
*/
func TestParseFractalSearchResult(t *testing.T) {
	fs := FractalSearch{
		ID:      1,
		ThemeID: 2,
	}

	tests := []struct {
		name          string
		response      PerplexityResult
		expected      []FractalSearchParseResult
		expectedError error
	}{
		{
			name: "Empty Choices",
			response: PerplexityResult{
				SuccessResults: PerplexitySuccessResponse{
					Choices: []Choice{}, // No choices in response
				},
				ErrorMessage: "", // ErrorMessage is ignored by logic
			},
			expected:      nil,
			expectedError: errors.New("no choices returned in the response"),
		},
		{
			name: "Single Section with Points",
			response: PerplexityResult{
				SuccessResults: PerplexitySuccessResponse{
					Choices: []Choice{
						{
							Message: Message{
								Content: "# Section 1\n- Point 1\n- Point 2", // Valid content
							},
						},
					},
				},
			},
			expected: []FractalSearchParseResult{
				{
					FractalSearchID: 1,
					DisplayName:     "Section 1",
					PointTypeName:   "main",
					Points: []Point{
						{ThemeID: 2, Title: "Point 1", PointType: "main"},
						{ThemeID: 2, Title: "Point 2", PointType: "main"},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Multiple Sections with Points",
			response: PerplexityResult{
				SuccessResults: PerplexitySuccessResponse{
					Choices: []Choice{
						{
							Message: Message{
								Content: "# Section 1\n- Point 1 - Auckland \n- Point 2 - Dunedin \n# Section 2\n- Point 3 - Christchurch\n- Point 4 - Invercargill",
							},
						},
					},
				},
			},
			expected: []FractalSearchParseResult{
				{
					FractalSearchID: 1,
					DisplayName:     "Section 1",
					PointTypeName:   "main",
					Points: []Point{
						{ThemeID: 2, Title: "Point 1", PointType: "main", Description: "Auckland"},
						{ThemeID: 2, Title: "Point 2", PointType: "main", Description: "Dunedin"},
					},
				},
				{
					FractalSearchID: 1,
					DisplayName:     "Section 2",
					PointTypeName:   "main",
					Points: []Point{
						{ThemeID: 2, Title: "Point 3", PointType: "main", Description: "Christchurch"},
						{ThemeID: 2, Title: "Point 4", PointType: "main", Description: "Invercargill"},
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := parseFractalSearchResult(tt.response, fs)
			// Error validation
			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error()) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			// Results validation
			if len(results) != len(tt.expected) {
				t.Errorf("expected %d results, got %d", len(tt.expected), len(results))
			}

			for i := range results {

				if results[i].FractalSearchID != tt.expected[i].FractalSearchID {
					t.Errorf("mismatch FractalSearchID in result %d: expected %+v, got %+v", i, tt.expected[i].FractalSearchID, results[i].FractalSearchID)
				}

				if results[i].DisplayName != tt.expected[i].DisplayName {
					t.Errorf("mismatch DisplayName in result %d: expected %+v, got %+v", i, tt.expected[i].DisplayName, results[i].DisplayName)
				}

				if results[i].PointTypeName != tt.expected[i].PointTypeName {
					t.Errorf("mismatch  PointTypeName in result %d: expected %+v, got %+v", i, tt.expected[i].PointTypeName, results[i].PointTypeName)
				}

				if len(results[i].Points) != len(tt.expected[i].Points) {
					t.Errorf("mismatch Points in result %d: expected %+v, got %+v", i, tt.expected[i].Points, results[i].Points)
				}

				for j, point := range results[i].Points {
					expectedPoint := tt.expected[i].Points[j]

					if point.ThemeID != expectedPoint.ThemeID {
						t.Errorf("mismatch ThemeID in result %d, point %d: expected %d, got %d", i, j, expectedPoint.ThemeID, point.ThemeID)
					}

					if point.Title != expectedPoint.Title {
						t.Errorf("mismatch Title in result %d, point %d: expected %q, got %q", i, j, expectedPoint.Title, point.Title)
					}

					if point.PointType != expectedPoint.PointType {
						t.Errorf("mismatch PointType in result %d, point %d: expected %q, got %q", i, j, expectedPoint.PointType, point.PointType)
					}

					if point.FractalSearchID != expectedPoint.FractalSearchID {
						t.Errorf("mismatch FractalSearchID in result %d, point %d: expected %d, got %d", i, j, expectedPoint.FractalSearchID, point.FractalSearchID)
					}

					if point.Description != expectedPoint.Description {
						t.Errorf("mismatch Description in result %d, point %d: expected %q, got %q", i, j, expectedPoint.Description, point.Description)
					}
				}

			}
		})
	}
}
