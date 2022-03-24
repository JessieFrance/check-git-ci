package checkgitci

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Base URL for GitHub API
const baseURL = "https://api.github.com"

// ErrorFailedAPICall is returned when we receive an error or bad response
// from the GitHub API.
var ErrorFailedAPICall = errors.New("bad Response from GitHub API")

// commitsURL takes a repository owner and name, and returns the url to the
// GitHub API for viewing commmits.
func commitsURL(owner, name string) string {
	return fmt.Sprintf("%s/repos/%s/%s/commits", baseURL, owner, name)
}

// NewRepository takes an owner and name as string fields, and returns
// a pointer to a Repository. It automatically uses the owner and name fields
// to set the CommitsURL field.
func NewRepository(owner, name string) *Repository {
	return &Repository{
		Owner:      owner,
		Name:       name,
		CommitsURL: commitsURL(owner, name),
	}
}

// makeGetRequest helps make get requests. It takes a url, and
// returns a slice of bytes and an error (or nil if no error).
func makeGetRequest(url string) ([]byte, error) {

	// Get http request.
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers.
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	// Make request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check that the response was ok.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d - %s: %w", resp.StatusCode, resp.Status, ErrorFailedAPICall)
	}

	// Read response body into slice of bytes.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
