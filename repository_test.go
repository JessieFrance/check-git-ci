package checkgitci

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestResult struct {
	err       error
	success   bool
	completed bool
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
				err:       NoRepositoryName,
				success:   false,
				completed: false,
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
				err:       NoRepositoryOwner,
				success:   false,
				completed: false,
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
			commitsUrl: tc.commitsServer.URL,
			runsUrl:    tc.runsServer.URL,
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

	}
}
