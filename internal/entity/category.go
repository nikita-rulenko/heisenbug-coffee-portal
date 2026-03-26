package entity

import "time"

type Category struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrEmptyName
	}
	if c.Slug == "" {
		return ErrEmptySlug
	}
	return nil
}
