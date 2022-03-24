package checkgitci

import (
	"fmt"
)

// Base URL for GitHub API
const baseURL = "https://api.github.com"

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
