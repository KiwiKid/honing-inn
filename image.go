package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func SaveImage(saveDir string, imageData []byte, imageID string) error {
	// Define the directory where the image will be saved

	// Ensure the directory exists, create it if it doesn't
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create directory: %w", err)
	}

	// Define the file path using the imageID as the filename
	filePath := filepath.Join(saveDir, fmt.Sprintf("%s.png", imageID))

	// Write the image data to the file
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		return fmt.Errorf("unable to save image: %w", err)
	}

	return nil
}

func DeleteImage(saveDir string, imageID string) error {
	// Define the file path using the imageID as the filename
	filePath := filepath.Join(saveDir, fmt.Sprintf("%s.png", imageID))

	// Remove the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("unable to delete image: %w", err)
	}

	return nil
}
