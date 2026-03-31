package entity_test

import (
	"strings"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitCategoryValidateExtended(t *testing.T) {
	tests := []struct {
		name    string
		cat     entity.Category
		wantErr error
	}{
		{"negative sort order", entity.Category{Name: "Кофе", Slug: "coffee", SortOrder: -1}, nil},
		{"zero sort order", entity.Category{Name: "Кофе", Slug: "coffee", SortOrder: 0}, nil},
		{"large sort order", entity.Category{Name: "Кофе", Slug: "coffee", SortOrder: 999999}, nil},
		{"slug with numbers", entity.Category{Name: "Кофе", Slug: "coffee-123"}, nil},
		{"slug with underscores", entity.Category{Name: "Кофе", Slug: "iced_coffee"}, nil},
		{"slug with dots", entity.Category{Name: "Кофе", Slug: "v2.0.coffee"}, nil},
		{"name with newlines", entity.Category{Name: "Кофе\nнапитки", Slug: "coffee"}, nil},
		{"emoji slug", entity.Category{Name: "Кофе", Slug: "☕-coffee"}, nil},
		{"single char name", entity.Category{Name: "K", Slug: "k"}, nil},
		{"single char slug", entity.Category{Name: "Кофе", Slug: "k"}, nil},
		{"description with HTML", entity.Category{Name: "Кофе", Slug: "coffee", Description: "<b>Горячие</b> напитки"}, nil},
		{"very long slug", entity.Category{Name: "Кофе", Slug: strings.Repeat("a", 500)}, nil},
		{"name and slug same", entity.Category{Name: "coffee", Slug: "coffee"}, nil},
		{"multibyte slug", entity.Category{Name: "Кофе", Slug: "кофе-напитки"}, nil},
		{"slug with spaces", entity.Category{Name: "Кофе", Slug: "iced coffee"}, nil},
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

func TestUnitCategoryValidateNameBeforeSlug(t *testing.T) {
	c := entity.Category{Name: "", Slug: ""}
	if err := c.Validate(); err != entity.ErrEmptyName {
		t.Errorf("expected ErrEmptyName first, got %v", err)
	}
}

func TestUnitCategoryValidateAllFieldsPopulated(t *testing.T) {
	c := entity.Category{
		Name:        "Эспрессо напитки",
		Slug:        "espresso-drinks",
		Description: "Все виды эспрессо",
		SortOrder:   5,
	}
	if err := c.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}
