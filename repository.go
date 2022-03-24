package checkgitci

import (
	"encoding/json"
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

// setRunsURL sets the GitHub API url on a repository for the check-runs API endpoint.
func (r *Repository) setRunsURL() {
	r.RunsURL = fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-runs", baseURL, r.Owner, r.Name, r.Sha)

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

// GetMostRecentCommit queries the GitHub commits API endpoint,
// finds the Sha hash for the most recent Git commit in a repository,
// and stores it in the Sha field of a Repository struct. Except in testing, it should
// not take any arguments. The function returns an error or nil if no error.
func (r *Repository) GetMostRecentCommit(params ...getMostRecentCommitArgs) error {
	// Get commits API url.
	url := r.CommitsURL

	// If parameters are provided, then try to override the default url.
	// This should be for testing purposes only.
	if len(params) > 0 {
		url = params[0].url
	}

	// Make the GET request.
	bodyBytes, err := makeGetRequest(url)
	if err != nil {
		return err
	}

	// Unmarshall into the response object.
	var responseObject CommitsAPI
	json.Unmarshal(bodyBytes, &responseObject)

	// If the slice is empty, return empty string.
	if len(responseObject) == 0 {
		return nil
	}

	// Set last commit.
	r.Sha = responseObject[0].Sha

	return nil
}

// CheckRuns queries the GitHub check-runs API endpoint,
// and attaches select JSON to the Repository struct RunsResult field.
// This function should not take any arguments, except during testing.
// CheckRuns returns an error or nil if no error.
func (r *Repository) CheckRuns(params ...checkRunsArgs) error {

	// Set the check runs url.
	r.setRunsURL()
	url := r.RunsURL

	// Override the url if user supplies one (like in testing).
	if len(params) > 0 {
		url = params[0].url
	}

	// Make the request.
	bodyBytes, err := makeGetRequest(url)

	// Check for error.
	if err != nil {
		return err
	}

	// Unmarshall into the RunsResult field.
	json.Unmarshal(bodyBytes, &r.RunsResult)

	// No error, so return nil.
	return nil
}
