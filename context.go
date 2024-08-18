package main

import (
	"context"
	"net/http"
	"os"
)

type Config struct {
	Title string
}

// Use a custom type for keys to avoid conflicts in context values.
type contextKey string

const (
	titleKey contextKey = "Title"
)

func GetContext(r *http.Request) context.Context {
	title := os.Getenv("TITLE")
	if title == "" {
		title = "Honing-Inn on your new home"
	}
	ctx := context.WithValue(r.Context(), titleKey, title)

	return ctx
}
