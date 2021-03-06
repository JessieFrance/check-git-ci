package checkgitci

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock date for commits API endpoint.
var mockCommitsAPI1 = `
[{"sha": "hijklmnop"}, {"sha": "qrstuv"}, {"sha": "wxyz123"}]
`

// Mock data for check-runs API endpoint.
var mockRunsAPI1 = `{
		   "total_count": 3,
		   "check_runs": [
		     {
		       "name": "Node.js 14 on windows",
		       "status": "completed",
		       "conclusion": "success",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     },
		     {
		       "name": "Node.js 14 on ubuntu",
		       "status": "completed",
		       "conclusion": "success",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     },
		     {
		       "name": "Node.js 14 on mac",
		       "status": "completed",
		       "conclusion": "success",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     }
		   ]
		 }`
var mockRunsAPI2 = `{
		   "total_count": 3,
		   "check_runs": [
		     {
		       "name": "Node.js 14 on windows",
		       "status": "completed",
		       "conclusion": "success",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     },
		     {
		       "name": "Node.js 14 on ubuntu",
		       "status": "completed",
		       "conclusion": "success",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     },
		     {
		       "name": "Node.js 14 on mac",
		       "status": "completed",
		       "conclusion": "failure",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     }
		   ]
		 }`

var mockRunsAPI3 = `{
		   "total_count": 1,
		   "check_runs": [
		     {
		       "name": "Node.js 14 on mac",
		       "status": "pending",
		       "conclusion": "pending",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     }
		   ]
		 }`

var mockRunsAPINoRuns = `{
		   "total_count": 0,
		   "check_runs": []
		 }`

var skippedAPIRuns = `{
		   "total_count": 3,
		   "check_runs": [
		     {
		       "name": "Node.js 14 on windows",
		       "status": "completed",
		       "conclusion": "success",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     },
		     {
		       "name": "Node.js 14 on ubuntu",
		       "status": "completed",
		       "conclusion": "skipped",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     },
		     {
		       "name": "Node.js 14 on mac",
		       "status": "completed",
		       "conclusion": "success",
		       "started_at": "2022-02-14T01:38:26Z",
		       "completed_at": "2022-02-14T01:42:29Z"
		     }
		   ]
		 }`

type TestResult struct {
	err          error
	success      bool
	completed    bool
	hasCheckRuns bool
}

func TestMostRecentCommitWasSuccess(t *testing.T) {

	// Setup test cases.
	testCases := []struct {
		testName      string
		repoOwner     string
		repoName      string
		commitsServer *httptest.Server
		runsServer    *httptest.Server
		expected      TestResult
	}{
		{
			testName:  "no repository name",
			repoOwner: "facebook",
			repoName:  "",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This should never be called since name is blank...
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This should never be called since name is blank...
			})),
			expected: TestResult{
				err:          ErrorNoRepositoryName,
				success:      false,
				completed:    false,
				hasCheckRuns: false,
			},
		},
		{
			testName:  "no repository owner",
			repoOwner: "",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This should never be called since owner is blank...
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This should never be called since owner is blank...
			})),
			expected: TestResult{
				err:          ErrorNoRepositoryOwner,
				success:      false,
				completed:    false,
				hasCheckRuns: false,
			},
		},
		{
			testName:  "bad response from GitHub commits API",
			repoOwner: "bad-octocat-owner",
			repoName:  "bad-octocat-name",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Send back a bad request.
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`octocat says that is not a real repository owner or name`))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This server's response should no matter...
			})),
			expected: TestResult{
				err:          ErrorFailedAPICall,
				success:      false,
				completed:    false,
				hasCheckRuns: false,
			},
		},
		{
			testName:  "bad response from GitHub check-runs API",
			repoOwner: "facebook",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockCommitsAPI1))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Send bad response.
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`octocat says that is a bad request to check-runs api`))
			})),
			expected: TestResult{
				err:          ErrorFailedAPICall,
				success:      false,
				completed:    false,
				hasCheckRuns: false,
			},
		},
		{
			testName:  "io.ReadAll error",
			repoOwner: "facebook",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This reponse doesn't matter here, only the runsServer.
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockCommitsAPI1))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate an io.ReadAll error:
				// https://stackoverflow.com/questions/53171123/how-to-force-error-on-reading-response-body
				w.Header().Set("Content-Length", "1")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`The response here is not of length 1 as specified in the header`))
			})),
			expected: TestResult{
				err:          ErrorIOReadAll,
				success:      false,
				completed:    false,
				hasCheckRuns: false,
			},
		},
		{
			testName:  "3 fully successful and complete runs",
			repoOwner: "facebook",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// The response doesn't matter here as long as the status is ok, because only
				// the response from the runsServer is important in this test case.
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockCommitsAPI1))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockRunsAPI1))
			})),
			expected: TestResult{
				err:          nil,
				success:      true,
				completed:    true,
				hasCheckRuns: true,
			},
		},
		{
			testName:  "3 complete runs, 2 success, 1 skipped",
			repoOwner: "facebook",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// The response doesn't matter here as long as the status is ok, because only
				// the response from the runsServer is important in this test case.
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockCommitsAPI1))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockRunsAPI1))
			})),
			expected: TestResult{
				err:          nil,
				success:      true,
				completed:    true,
				hasCheckRuns: true,
			},
		},
		{
			testName:  "1 failed run but all runs complete",
			repoOwner: "facebook",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// The response doesn't matter here as long as the status is ok, because only
				// the response from the runsServer is important in this test case.
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockCommitsAPI1))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockRunsAPI2))
			})),
			expected: TestResult{
				err:          nil,
				success:      false,
				completed:    true,
				hasCheckRuns: true,
			},
		},
		{
			testName:  "1 incomplete run",
			repoOwner: "facebook",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// The response doesn't matter here as long as the status is ok, because only
				// the response from the runsServer is important in this test case.
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockCommitsAPI1))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockRunsAPI3))
			})),
			expected: TestResult{
				err:          nil,
				success:      false,
				completed:    false,
				hasCheckRuns: true,
			},
		},
		{
			testName:  "no runs",
			repoOwner: "facebook",
			repoName:  "react",
			commitsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// The response doesn't matter here as long as the status is ok, because only
				// the response from the runsServer is important in this test case.
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockCommitsAPI1))
			})),
			runsServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(mockRunsAPINoRuns))
			})),
			expected: TestResult{
				err:          nil,
				success:      false,
				completed:    true,
				hasCheckRuns: false,
			},
		},
	}

	// Iterate over each individual test case (tc).
	for _, tc := range testCases {
		// Close servers when we're done.
		defer tc.commitsServer.Close()
		defer tc.runsServer.Close()

		// Create a repository from the test case.
		repo := NewRepository(tc.repoOwner, tc.repoName)

		// Make a call via receiver to the TestMostRecentCommitWasSuccess
		// function, but override the urls to use test urls.
		err := repo.MostRecentCommitWasSuccess(mostRecentCommitArgs{
			commitsURL: tc.commitsServer.URL,
			runsURL:    tc.runsServer.URL,
		})

		// Check for the expected error.
		if err != tc.expected.err {
			t.Errorf("expected error to be %v but got %v", tc.expected.err, err)
		}

		// Check for expected success state.
		if repo.Success != tc.expected.success {
			t.Errorf("expected repository success field to be %v but got %v", tc.expected.success, repo.Success)
		}

		// Check for expected completed state.
		if repo.Completed != tc.expected.completed {
			t.Errorf("expected repository completed field to be %v but got %v", tc.expected.completed, repo.Completed)
		}

		// Check for expected hasCheckRuns status.
		if repo.HasCheckRuns != tc.expected.hasCheckRuns {
			t.Errorf("expected repository hasCheckRuns field to be %v but got %v", tc.expected.hasCheckRuns, repo.HasCheckRuns)
		}

	}
}
