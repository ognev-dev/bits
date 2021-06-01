// GitHub's repository
// https://docs.github.com/en/rest/reference/repos#get-a-repository
// GET https://api.github.com/repos/{user}/{name}

package model

import (
	"context"
	"time"
)

type Repository struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"-"`
	User   *User `json:"-"`

	// Path is just a "{user}/{repo}" on GitHub
	// note that this is equals to "full_name" on GitHub API
	Path string `json:"path"`

	Name        string `json:"name"`
	Description string `json:"description"`
	CloneURL    string `json:"-"`

	// Secret is needed to authenticate repository from Github
	// It is random string given to the user who creates repository
	// and must be set on repository's push webhook on Github
	// https://github.com/{account}/{repo}/settings/hooks
	// Secret is hashed using bcrypt and available only once
	Secret string `json:"-"`

	Type RepoType `json:"type"`

	// Confirmed is a flag to mark that user is confirmed access to repo
	// Confirmed is set to true on first successful import
	Confirmed bool `json:"confirmed"`

	// TODO make proper import logs
	ImportStatus RepoImportStatus `json:"import_status"`
	ImportLog    string           `json:"import_log"`
	ImportAt     *time.Time       `json:"import_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (m *Repository) BeforeInsert(ctx context.Context) (context.Context, error) {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	return ctx, nil
}

func (m *Repository) BeforeUpdate(ctx context.Context) (context.Context, error) {
	now := time.Now()
	m.UpdatedAt = &now

	return ctx, nil
}

func (m *Repository) IsPrivate() bool {
	return m.Type == RepoTypePrivate
}

type RepoImportStatus string

const (
	RepoImportStatusOK  = "ok"
	RepoImportStatusErr = "error"
)

type RepoType string

const (
	// RepoTypePrivate
	// Private repos is available only to user who added it
	RepoTypePrivate RepoType = "private"

	// RepoTypeGlobal
	// Data from global repos is available by default at global level
	// (it's kind of "trusted" or "featured")
	// This type cannot be selected by user
	RepoTypeGlobal RepoType = "global"

	// RepoTypePublic
	// Public repos will be listed in repos page and in search filters
	RepoTypePublic RepoType = "public"

	// RepoTypeHidden
	// Hidden repos will NOT be listed in repos page and in search filters
	// but can be viewed using link
	RepoTypeHidden RepoType = "hidden"
)

var RepositoryTypesList = []string{
	"private", "global", "public", "hidden",
}

func (t RepoType) IsValid() bool {
	switch t {
	case
		RepoTypePrivate,
		RepoTypeGlobal,
		RepoTypeHidden,
		RepoTypePublic:
		return true
	}

	return false
}

func (t RepoType) String() string {
	return string(t)
}
