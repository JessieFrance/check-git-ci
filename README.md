# Check Git CI (checkgitci)

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/JessieFrance/check-git-ci)
[![Actions Status](https://github.com/JessieFrance/check-git-ci/workflows/Build%20and%20Test/badge.svg)](https://github.com/JessieFrance/check-git-ci/actions)
[![Coverage](https://gocover.io/_badge/github.com/JessieFrance/check-git-ci)](https://gocover.io/github.com/JessieFrance/check-git-ci)
[![GitHub license](https://img.shields.io/github/license/JessieFrance/check-git-ci?style=flat-square)](https://github.com/JessieFrance/check-git-ci/blob/main/LICENSE)

> A go module for checking GitHub CI endpoints.

## Introduction

The `checkgitci` go module provides users with functions that access GitHub API endpoints related to continuous integration. This module contains functionality for getting the most recent commit hash, and checking if the most recent commit passed GitHub workflows.

## Installation

	go get github.com/JessieFrance/check-git-ci

## Examples

### Obtain the Hash for the Most Recent Commit with `GetMostRecentCommit`

`GetMostRecentCommit` makes a request to GitHub's `commits` API and finds the most recent commit:

```go
package main

import (
	"fmt"
	"os"

	checkgitci "github.com/JessieFrance/check-git-ci"
)

func main() {

	// Initialize a GitHub repository by its owner and name.
	r := checkgitci.NewRepository("caddyserver", "caddy")

	// Get the most recent commit in this repository.
	err := r.GetMostRecentCommit()
	if err != nil {
		fmt.Println("Unable to access GitHub commits API:", err)
		os.Exit(1)
	}

	// Print the most recent commit.
	fmt.Println("Most recent commit hash for repository: ", r.Sha)
	
}
```

### Check if the Most Recent Commit had Successful CI Runs

`MostRecentCommitWasSuccess` internally calls `GetMostRecentCommit` to get the last commit, and then makes a request to GitHub's `check-runs` API to see if that commit had successful GitHub CI workflows:

```go
package main

import (
	"fmt"
	"os"

	checkgitci "github.com/JessieFrance/check-git-ci"
)

func main() {

	// Initialize a GitHub repository by its owner and name.
	r := checkgitci.NewRepository("caddyserver", "caddy")

	// Get the most recent commit, and check if it was successful.
	err := r.MostRecentCommitWasSuccess()
	if err != nil {
		fmt.Println("Something went wrong:", err)
		os.Exit(1)
	}

	// First, check if there are runs for this repository, as not all
	// repositories have GitHub workflows. Additionally, for a repository that
	// uses workflows, there may be brief period of time between when a user 
	// pushes a commit, and when GitHub sets up the workflows that say "pending"
	// for that commit.
	if r.HasCheckRuns {
		// Second, check if the workflows are all complete,
		// or if some are still pending.
		if r.Completed {
			// Third, check if the runs were all successful.
			// Like GitHub default behavior, this will also be true if
			// some workflow runs are skipped.
			if r.Success == true {
				fmt.Println("Workflows runs were successful!")
			} else {
				fmt.Println("Some workflow runs were not successful.")
			}
		} else {
			fmt.Println("Some workflows are still pending.")
		}
	} else {
		fmt.Println("No workflows.")
	}
}
```


## License

MIT
