package checkgitci

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// makeGetRequest helps make get requests. It takes a sturct that
// contains a url, lastModified, and API key field, and
// returns a slice of bytes for the API content, a string for the new last modified
// header, an int for the remaining API calls, and an error (or nil if no error).
func makeGetRequest(requestInput makeRequestArgs) ([]byte, string, int, error) {

	// Extract information.
	url := requestInput.url
	lastModified := requestInput.lastModified
	apiKey := requestInput.apiKey

	// Get http request.
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", -1, err
	}

	// Add headers.
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Content-Type", "application/json")
	if len(lastModified) > 0 {
		req.Header.Add("If-Modified-Since", lastModified)
	}
	if len(apiKey) > 0 {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", apiKey))
	}

	// Make request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", -1, err
	}
	defer resp.Body.Close()

	// Check that the response was ok.
	if resp.StatusCode != http.StatusOK {
		return nil, "", -1, ErrorFailedAPICall
	}

	// Get last modified and remaining.
	newLastModified := resp.Header.Get("Last-Modified")
	remainingStr := resp.Header.Get("x-ratelimit-remaining")

	// Convert remaining string to int.
	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		return nil, newLastModified, -1, err
	}

	// Check for 304, and return early if so.
	if resp.StatusCode == http.StatusNotModified {
		return nil, newLastModified, remaining, nil
	}

	// Read response body into slice of bytes.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", -1, ErrorIOReadAll
	}

	// Return.
	return bodyBytes, newLastModified, remaining, nil
}

// GetMostRecentCommit queries the GitHub commits API endpoint,
// finds the Sha hash for the most recent Git commit in a repository,
// and stores it in the Sha field of a Repository struct.
// This function returns an error or nil if no error. This function also
// takes optional arguments (for example to override the url for the
// GitHub commits API), but these are mostly for testing.
func (r *Repository) GetMostRecentCommit(params ...getMostRecentCommitArgs) error {
	// Get commits API url.
	url := r.CommitsURL

	// If parameters are provided, then try to override the default url.
	// This should be for testing purposes only.
	if len(params) > 0 {
		url = params[0].url
	}

	// Get the API key (if there is one) for either this repository, or
	// a RateManager for multiple repositories.
	key := r.getCorrectKey()

	// Generate arguments for making GET request.
	lookup := r.getRepoLookup()
	args := makeRequestArgs{
		url:          url,
		lastModified: r.RateManager.CallHeaders[lookup].CommitsHeader,
		apiKey:       key,
	}
	// Make the GET request.
	bodyBytes, commitsHeader, remaining, err := makeGetRequest(args)
	if err != nil {
		return err
	}

	// Set the commits header for making future GitHub API requests for
	// this repository.
	r.setCommitsHeader(lookup, commitsHeader)

	// Set the 'remaining' variable on either a general
	// rate manager, or on this repository.
	r.updateRemaining(lookup, remaining)

	// If bodyBytes are nil (for respones 304), return early.
	if bodyBytes == nil {
		return nil
	}

	// Unmarshall into the response object.
	var responseObject CommitsAPI
	json.Unmarshal(bodyBytes, &responseObject)

	// If the slice is empty, just return nil without setting
	// most recent commit.
	if len(responseObject) == 0 {
		return nil
	}

	// Set last commit.
	r.Sha = responseObject[0].Sha

	// Set the check runs url.
	r.setRunsURL()

	// No error.
	return nil
}

// CheckRuns queries the GitHub check-runs API endpoint for workflows,
// and attaches select JSON to the Repository struct RunsResult field.
// CheckRuns returns an error or nil if no error.
// This function also takes optional arguments (for example to override
// the url for the GitHub workflows API), but these are mostly for testing.
func (r *Repository) CheckRuns(params ...checkRunsArgs) error {

	// Override the url if user supplies one (like in testing).
	url := r.RunsURL
	if len(params) > 0 {
		url = params[0].url
	}

	// TODO: Check url is not blank if user is calling this function
	// independently.

	// Make the request.
	bodyBytes, err := makeGetRequest(url)

	// Check for error.
	if err != nil {
		return err
	}

	// Unmarshall into the RunsResult field.
	json.Unmarshal(bodyBytes, &r.RunsResult)

	// Check if there are runs...
	if r.RunsResult.TotalCount == 0 {
		// No runs, so set state variables.
		r.HasCheckRuns = false
		r.Success = false
		r.Completed = true
	} else {
		// There are runs, and other state variables can be
		// set later.
		r.HasCheckRuns = true
	}

	// No error, so return nil.
	return nil
}

// MostRecentCommitWasSuccess makes API calls to get the most recent commit
// and GitHub CI runs associated with that commit. This function then checks
// if last commit was successful, and if the runs were all completed. This
// function stores the results of these checks on the repository
// Success and Completed fields. This function will return an error (or
// nil if there is not an error).
func (r *Repository) MostRecentCommitWasSuccess(params ...mostRecentCommitArgs) error {

	// Throw errors if no owner/name.
	if r.Name == "" {
		return ErrorNoRepositoryName
	}
	if r.Owner == "" {
		return ErrorNoRepositoryOwner
	}

	// Set the url for checking the most recent commit.
	// If the user provided a url (like in a test),
	// then override it.
	url := r.CommitsURL
	if len(params) > 0 {
		url = params[0].commitsURL
	}

	// Get the most recent commit.
	err := r.GetMostRecentCommit(getMostRecentCommitArgs{url})
	if err != nil {
		return err
	}

	// Get url for checking the check-runs API.
	// Override it (like for tests) if user provided arguments.
	url = r.RunsURL
	if len(params) > 0 {
		url = params[0].runsURL
	}

	// Check the individual CI runs.
	err = r.CheckRuns(checkRunsArgs{url})
	if err != nil {
		return err
	}

	// Check if the CI runs were successful
	// and Completed.
	r.RunsAreSuccessful()
	r.RunsAreComplete()

	// No error.
	return nil
}
