package entity

import "time"

type NewsItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (n *NewsItem) Validate() error {
	if n.Title == "" {
		return ErrEmptyName
	}
	if n.Content == "" {
		return ErrEmptyContent
	}
	return nil
}

func (n *NewsItem) Summary(maxRunes int) string {
	runes := []rune(n.Content)
	if len(runes) <= maxRunes {
		return n.Content
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
