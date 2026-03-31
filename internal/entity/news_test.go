package entity_test

import (
	"strings"
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
		{
			name:    "both empty",
			news:    entity.NewsItem{Title: "", Content: ""},
			wantErr: entity.ErrEmptyName, // title checked first
		},
		{
			name:    "unicode title",
			news:    entity.NewsItem{Title: "新メニュー", Content: "New items"},
			wantErr: nil,
		},
		{
			name:    "whitespace title is valid",
			news:    entity.NewsItem{Title: "  ", Content: "Content"},
			wantErr: nil,
		},
		{
			name:    "long title",
			news:    entity.NewsItem{Title: strings.Repeat("A", 500), Content: "Content"},
			wantErr: nil,
		},
		{
			name:    "long content",
			news:    entity.NewsItem{Title: "Title", Content: strings.Repeat("Текст ", 1000)},
			wantErr: nil,
		},
		{
			name:    "with author",
			news:    entity.NewsItem{Title: "Тест", Content: "Контент", Author: "Админ"},
			wantErr: nil,
		},
		{
			name:    "empty author is valid",
			news:    entity.NewsItem{Title: "Тест", Content: "Контент", Author: ""},
			wantErr: nil,
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
		{"one less than length", contentRunes - 1, string([]rune(n.Content)[:contentRunes-4]) + "..."},
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

func TestUnitNewsItemSummaryEmptyContent(t *testing.T) {
	n := entity.NewsItem{Content: ""}
	got := n.Summary(10)
	if got != "" {
		t.Errorf("Summary on empty content = %q, want empty", got)
	}
}

func TestUnitNewsItemSummaryOneChar(t *testing.T) {
	n := entity.NewsItem{Content: "X"}
	got := n.Summary(10)
	if got != "X" {
		t.Errorf("Summary(10) on single char = %q, want %q", got, "X")
	}
}

func TestUnitNewsItemSummaryMaxRunesZero(t *testing.T) {
	n := entity.NewsItem{Content: "Hello"}
	got := n.Summary(0)
	if got != "" {
		t.Errorf("Summary(0) = %q, want empty", got)
	}
}

func TestUnitNewsItemSummaryMaxRunesOne(t *testing.T) {
	n := entity.NewsItem{Content: "Hello"}
	got := n.Summary(1)
	if got != "H" {
		t.Errorf("Summary(1) = %q, want %q", got, "H")
	}
}

func TestUnitNewsItemSummaryMaxRunesTwo(t *testing.T) {
	n := entity.NewsItem{Content: "Hello"}
	got := n.Summary(2)
	if got != "He" {
		t.Errorf("Summary(2) = %q, want %q", got, "He")
	}
}

func TestUnitNewsItemSummaryUnicode(t *testing.T) {
	n := entity.NewsItem{Content: "Привет мир!"}
	got := n.Summary(6)
	if got != "При..." {
		t.Errorf("Summary(6) = %q, want %q", got, "При...")
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
		{
			name:    "both empty",
			cat:     entity.Category{Name: "", Slug: ""},
			wantErr: entity.ErrEmptyName, // name checked first
		},
		{
			name:    "with description",
			cat:     entity.Category{Name: "Кофе", Slug: "coffee", Description: "Кофейные напитки"},
			wantErr: nil,
		},
		{
			name:    "with sort order",
			cat:     entity.Category{Name: "Кофе", Slug: "coffee", SortOrder: 10},
			wantErr: nil,
		},
		{
			name:    "unicode slug",
			cat:     entity.Category{Name: "Кофе", Slug: "кофе"},
			wantErr: nil,
		},
		{
			name:    "slug with hyphens",
			cat:     entity.Category{Name: "Iced Coffee", Slug: "iced-coffee"},
			wantErr: nil,
		},
		{
			name:    "long name",
			cat:     entity.Category{Name: strings.Repeat("Категория", 50), Slug: "long"},
			wantErr: nil,
		},
		{
			name:    "whitespace name is valid",
			cat:     entity.Category{Name: " ", Slug: "space"},
			wantErr: nil,
		},
		{
			name:    "whitespace slug is valid",
			cat:     entity.Category{Name: "Test", Slug: " "},
			wantErr: nil,
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
