package github

import (
	"strconv"
	"time"

	"github.com/google/go-github/v60/github"
)

// Timestamp represents a time that can be unmarshalled from a JSON string
// formatted as either an RFC3339 or Unix timestamp. This is necessary for some
// fields since the GitHub API is inconsistent in how it represents times. All
// exported methods of time.Time can be called on Timestamp.
type Timestamp struct {
	time.Time
}

func (t Timestamp) String() string {
	return t.Time.String()
}

// GetTime returns std time.Time.
func (t *Timestamp) GetTime() *time.Time {
	if t == nil {
		return nil
	}
	return &t.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Time is expected in RFC3339 or Unix format.
func (t *Timestamp) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	i, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		t.Time = time.Unix(i, 0)
		if t.Time.Year() > 3000 {
			t.Time = time.Unix(0, i*1e6)
		}
	} else {
		t.Time, err = time.Parse(`"`+time.RFC3339+`"`, str)
	}
	return
}

// Equal reports whether t and u are equal based on time.Equal
func (t Timestamp) Equal(u Timestamp) bool {
	return t.Time.Equal(u.Time)
}

// UserProfile represents GitHub user information sent to the client
type UserProfile struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Bio       string `json:"bio,omitempty"`
	Company   string `json:"company,omitempty"`
	Location  string `json:"location,omitempty"`
}

// Repository represents a GitHub repository
type Repository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url,omitempty"`
	SSHURL      string `json:"ssh_url,omitempty"`
	Owner       struct {
		Login     string `json:"login"`
		ID        int64  `json:"id"`
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
	DefaultBranch string `json:"default_branch,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
	IsEnterprise  bool   `json:"is_enterprise"`
	Team          string `json:"team,omitempty"`
	ProdBranch    string `json:"prodBranch,omitempty"`
	DevBranch     string `json:"devBranch,omitempty"`
}

// Branch represents a GitHub repository branch
type Branch struct {
	Name      string `json:"name"`
	CommitSHA string `json:"commit_sha"`
	Protected bool   `json:"protected"`
	IsDefault bool   `json:"is_default,omitempty"`
}

// FileContent represents a file content from GitHub
type FileContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	Type        string `json:"type"` // "file" or "dir"
	Content     string `json:"content,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url,omitempty"`
}

// Organization represents a GitHub organization
type Organization struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	HTMLURL   string `json:"html_url,omitempty"`
}

// DirectoryContent represents a directory listing from GitHub
type DirectoryContent []FileContent

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// PullRequestUser represents a GitHub user who created or is assigned to a pull request
type PullRequestUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID        int64             `json:"id"`
	Number    int               `json:"number"`
	Title     string            `json:"title"`
	State     string            `json:"state"` // "open", "closed"
	User      PullRequestUser   `json:"user"`
	Body      string            `json:"body"`
	CreatedAt github.Timestamp  `json:"created_at"`
	UpdatedAt github.Timestamp  `json:"updated_at"`
	ClosedAt  *github.Timestamp `json:"closed_at"`
	MergedAt  *github.Timestamp `json:"merged_at"`
	HTMLURL   string            `json:"html_url"`
}

// PullRequestResponse represents the response for a pull requests API request
type PullRequestResponse struct {
	PullRequests []PullRequest `json:"pull_requests"`
}

// Secret represents a repository action secret.
type Secret struct {
	Name                    string    `json:"name"`
	CreatedAt               Timestamp `json:"created_at"`
	UpdatedAt               Timestamp `json:"updated_at"`
	Visibility              string    `json:"visibility,omitempty"`
	SelectedRepositoriesURL string    `json:"selected_repositories_url,omitempty"`
}

// Secrets represents one item from the ListSecrets response.
type Secrets struct {
	TotalCount int       `json:"total_count"`
	Secrets    []*Secret `json:"secrets"`
}
