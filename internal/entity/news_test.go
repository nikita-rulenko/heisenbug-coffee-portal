package entity_test

import (
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitNewsItemValidate(t *testing.T) {
	tests := []struct {
		name    string
		news    entity.NewsItem
		wantErr error
	}{
		{
			name:    "valid news item",
			news:    entity.NewsItem{Title: "Новая акция", Content: "Скидка 20% на все напитки"},
			wantErr: nil,
		},
		{
			name:    "empty title",
			news:    entity.NewsItem{Title: "", Content: "Какой-то контент"},
			wantErr: entity.ErrEmptyName,
		},
		{
			name:    "empty content",
			news:    entity.NewsItem{Title: "Заголовок", Content: ""},
			wantErr: entity.ErrEmptyContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.news.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnitNewsItemSummary(t *testing.T) {
	n := entity.NewsItem{Content: "Это длинный текст новости, который нужно обрезать для превью в ленте"}

	contentRunes := len([]rune(n.Content))
	tests := []struct {
		name     string
		maxRunes int
		want     string
	}{
		{"short enough", 200, n.Content},
		{"truncated", 20, "Это длинный текст..."},
		{"very short", 5, "Эт..."},
		{"tiny maxLen", 3, "Это"},
		{"exact length", contentRunes, n.Content},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.Summary(tt.maxRunes)
			if got != tt.want {
				t.Errorf("Summary(%d) = %q, want %q", tt.maxRunes, got, tt.want)
			}
		})
	}
}

func TestUnitCategoryValidate(t *testing.T) {
	tests := []struct {
		name    string
		cat     entity.Category
		wantErr error
	}{
		{
			name:    "valid category",
			cat:     entity.Category{Name: "Эспрессо", Slug: "espresso"},
			wantErr: nil,
		},
		{
			name:    "empty name",
			cat:     entity.Category{Name: "", Slug: "espresso"},
			wantErr: entity.ErrEmptyName,
		},
		{
			name:    "empty slug",
			cat:     entity.Category{Name: "Эспрессо", Slug: ""},
			wantErr: entity.ErrEmptySlug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cat.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
