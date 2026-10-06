package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	CategoryQuery   CategoryQueryRepository
	CategoryCommand CategoryCommandRepository
}

func NewRepositories(DB *sqlx.DB) *Repositories {
	return &Repositories{
		CategoryQuery:   NewCategoryQueryRepository(DB),
		CategoryCommand: NewCategoryCommandRepository(DB),
	}
}
