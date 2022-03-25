package checkgitci

import "errors"

// ErrorFailedAPICall is returned when we receive an error or bad response
// from the GitHub API.
var ErrorFailedAPICall = errors.New("Error: bad Response from GitHub API")

// ErrorNoRepositoryName is returned when trying to perform an operation that requires
// a repository name that has not yet been set.
var ErrorNoRepositoryName = errors.New("Error: repository name field cannot be blank")

// ErrorNoRepositoryOwner is returned when trying to perform an operation that requires
// a repository owner that has not yet been set.
var ErrorNoRepositoryOwner = errors.New("Error: repository owner field cannot be blank")
