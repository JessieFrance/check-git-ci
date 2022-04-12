package checkgitci

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// commitsURL takes a repository owner and name, and returns the url to the
// GitHub API for viewing commmits.
func commitsURL(owner, name string) string {
	return fmt.Sprintf("%s/repos/%s/%s/commits", baseURL, owner, name)
}

// setRunsURL sets the GitHub API url on a repository for the check-runs API endpoint.
func (r *Repository) setRunsURL() {
	r.RunsURL = fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-runs", baseURL, r.Owner, r.Name, r.Sha)

}

// getRepoLookup returns a string that represents the lookup key for a
// repository.
func (r *Repository) getRepoLookup() string {
	return fmt.Sprintf("%s%s", r.Owner, r.Name)
}

// NewRepository takes an owner and name as string fields, and returns
// a pointer to a Repository. It automatically uses the owner and name fields
// to set the CommitsURL field.
func NewRepository(owner, name string) *Repository {
	rm := NewRateManager()
	return &Repository{
		Owner:       owner,
		Name:        name,
		CommitsURL:  commitsURL(owner, name),
		RateManager: &rm,
	}
}

// SetKey sets a GitHub API key for a specific repository.
func (r *Repository) SetKey(key string) {
	rm := r.RateManager
	rm.APIKeys[r.getRepoLookup()] = key
	r.RateManager = rm
}

// genKeyIsSet returns true if the general key (intended to be shared amongst
// several repositories) is set for a RateManager, and false otherwise.
func (r *Repository) genKeyIsSet() bool {
	if len(r.RateManager.APIKeys[generalKey]) > 0 {
		return true
	}
	return false
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
		// TODO: Consider giving more informative error.
		return nil, ErrorFailedAPICall
	}

	// Read response body into slice of bytes.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorIOReadAll
	}
	return bodyBytes, nil
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

	// Make the GET request.
	bodyBytes, err := makeGetRequest(url)
	if err != nil {
		return err
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

// RunsAreSuccessful iterates over a repository's CI runs, and
// sets the repository's "Success" field as either true or false.
// The Success field will be set to false if there are no
// runs (in CheckRuns function), or if some runs were not successful.
// The Success field will be true if there are runs, and they are
// all marked as "success" or "skipped".
func (r *Repository) RunsAreSuccessful() {

	// If there are no runs, then return early.
	if !r.HasCheckRuns {
		return
	}

	// Iterate over runs.
	for _, run := range r.RunsResult.CheckRuns {
		// If current run is neither successful nor skipped, then
		// return early.
		if run.Conclusion != "success" && run.Conclusion != "skipped" {
			r.Success = false
			return
		}
	}
	// If we made it this far, runs were successful.
	r.Success = true

}

// RunsAreComplete sets the repository "Completed" field to true
// if all the CI runs for the last commit are complete.
// This function sets the Completed state to false if some runs
// are still pending.
func (r *Repository) RunsAreComplete() {
	// If there are no runs, return early.
	if !r.HasCheckRuns {
		return
	}

	// Iterate over runs.
	for _, run := range r.RunsResult.CheckRuns {
		// If current run is not complete,
		// return early.
		if run.Status != "completed" {
			r.Completed = false
			return
		}
	}
	// All runs have been checked, and are complete
	// if we made it this far.
	r.Completed = true
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
