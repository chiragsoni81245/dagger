package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// Function to calculate SHA-256 hash of a file
func CalculateSHA256(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	// Create a new SHA-256 hasher
	hasher := sha256.New()

	// Copy file content into the hasher
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("could not hash file: %w", err)
	}

	// Get final SHA-256 hash sum
	hash := hasher.Sum(nil)

	// Convert hash to a hexadecimal string
	return fmt.Sprintf("%x", hash), nil
}

