package model

type Topic struct {
	ID     int64  `json:"id"`
	RepoID int64  `json:"-"`
	Name   string `json:"name"`
}
