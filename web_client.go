package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/html"
)

type SiteMeta struct {
	Url         string
	Title       string
	Address     string
	Suburb      string
	Description string
	Keywords    string
	MetaImage   string
}

func GetWebMeta(url string) (*SiteMeta, error) {
	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("GetWebMeta - Failed to fetch URL: %v", err)
		return nil, fmt.Errorf("error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	// Parse the HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Printf("GetWebMeta - Failed to parse URL: %v", err)
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	sm := &SiteMeta{
		Url: url,
	} // Initialize SiteMeta

	// Extract meta tags
	var extractMetaTags func(*html.Node)
	extractMetaTags = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string

			for _, attr := range n.Attr {
				if attr.Key == "name" || attr.Key == "property" {
					name = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}

			}

			switch name {
			case "title", "og:title":
				sm.Title = content
				log.Printf("Mapped meta tag %s: %s", name, content)
			case "description", "og:description":
				sm.Description = content
				log.Printf("Mapped meta tag %s: %s", name, content)
			case "keywords":
				sm.Keywords = content
				log.Printf("Mapped meta tag %s: %s", name, content)
			case "og:image":
				sm.MetaImage = content
				log.Printf("Mapped meta tag %s: %s", name, content)

			default:
				if len(content) > 30 {
					log.Printf("Unmapped meta tag %s length: %d", name, len(content))
				} else {
					log.Printf("Unmapped meta tag %s: (%s)", name, content)
				}
			}
		}
		// Recursively visit child nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMetaTags(c)
		}
	}

	extractMetaTags(doc)

	if len(sm.Title) > 0 {
		sm.Address = cleanAddress(sm.Title)
	}

	return sm, nil
}
