package models

type Category struct {
	ID          int64  `json:"id" db:"id"`
	Slug        string `json:"slug" db:"slug"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
}
