package response

import "github.com/refto/server/database/model"

type CreateRepository struct {
	Secret string `json:"secret"`
}

type FilterRepositories struct {
	Data  []model.Repository `json:"data"`
	Count int                `json:"count"`
}
