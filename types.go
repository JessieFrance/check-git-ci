package checkgitci

import "time"

// Repository type holds information for individual Git repositories.
type Repository struct {
	Owner        string
	Name         string
	Sha          string
	RunsResult   CheckRunsAPI
	HasCheckRuns bool
	Success      bool
	Completed    bool
	CommitsURL   string
	RunsURL      string
}

// CommitsAPI holds selected information on the response from GitHub commits API.
type CommitsAPI []struct {
	Sha string
}

// CheckRunsAPI holds selected information from the GitHub check-runs API.
type CheckRunsAPI struct {
	TotalCount int   `json:"total_count"`
	CheckRuns  []Run `json:"check_runs"`
}

// Run holds selected information on an individual GitHub CI workflow run.
type Run struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// getMostRecentCommitArgs overrides the url used in
// the GetMostRecentCommit function. This argument
// should usually only be supplied for testing.
type getMostRecentCommitArgs struct {
	url string
}

// checkRunsArgs overrides the url used in
// the CheckRuns function. This argument should
// usually only be supplied for testing.
type checkRunsArgs struct {
	url string
}

// mostRecentCommitArgs overrides the urls used in
// the MostRecentCommitWasSuccess function. This argument
// should usually only be supplied for testing.
type mostRecentCommitArgs struct {
	commitsURL string
	runsURL    string
}
