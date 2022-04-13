package checkgitci

import (
	"fmt"
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

// GetRemaining returns the remaining GitHub API calls for a repository.
func (r *Repository) GetRemaining() int {
	return r.RateManager.Remaining[r.getRepoLookup()]
}

// setCommitsHeader sets the header for the GitHub commits API.
func (r *Repository) setCommitsHeader(lookup, commitsHeader string) {
	entry := r.RateManager.CallHeaders[lookup]
	entry.CommitsHeader = commitsHeader
	r.RateManager.CallHeaders[lookup] = entry
}

// setRunsHeader sets the header for the GitHub check-runs API.
func (r *Repository) setRunsHeader(lookup, runsHeader string) {
	entry := r.RateManager.CallHeaders[lookup]
	entry.RunsHeader = runsHeader
	r.RateManager.CallHeaders[lookup] = entry
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

// getCorrectKey first checks if a repository has its own key set,
// and if not, it trieds to access the general key (which might also
// not be set).
func (r *Repository) getCorrectKey() string {
	// Try to lookup the key for this repository.
	lookup := r.getRepoLookup()
	key := r.RateManager.APIKeys[lookup]

	// If the key for this repository has not been set, check if there is
	// a general key (i.e. for a common rate manager).
	if len(key) == 0 {
		key = r.RateManager.APIKeys[generalKey]
	}
	return key
}
