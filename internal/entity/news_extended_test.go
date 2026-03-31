package entity_test

import (
	"strings"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitNewsItemSummaryVariousLengths(t *testing.T) {
	content := "Это тестовая новость для проверки обрезки текста по рунам в системе"
	n := entity.NewsItem{Content: content}
	runes := []rune(content)

	tests := []struct {
		name     string
		maxRunes int
	}{
		{"max 1", 1},
		{"max 2", 2},
		{"max 3", 3},
		{"max 4", 4},
		{"max 5", 5},
		{"max 6", 6},
		{"max 7", 7},
		{"max 8", 8},
		{"max 9", 9},
		{"max 10", 10},
		{"max 15", 15},
		{"max 20", 20},
		{"max 50", 50},
		{"max 100", 100},
		{"max 200", 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := n.Summary(tt.maxRunes)
			gotRunes := []rune(got)
			if tt.maxRunes >= len(runes) {
				if got != content {
					t.Errorf("Summary(%d) = %q, want full content", tt.maxRunes, got)
				}
			} else if tt.maxRunes <= 3 {
				want := string(runes[:tt.maxRunes])
				if got != want {
					t.Errorf("Summary(%d) = %q, want %q", tt.maxRunes, got, want)
				}
			} else {
				if len(gotRunes) > tt.maxRunes {
					t.Errorf("Summary(%d) returned %d runes, want <= %d", tt.maxRunes, len(gotRunes), tt.maxRunes)
				}
				if !strings.HasSuffix(got, "...") {
					t.Errorf("Summary(%d) = %q, expected ... suffix", tt.maxRunes, got)
				}
			}
		})
	}
}

func TestUnitNewsItemSummaryMultibyteEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxRunes int
		wantLen  int // max rune length of result
	}{
		{"CJK chars", "你好世界测试文字", 4, 4},
		{"emoji", "\U0001f600\U0001f601\U0001f602\U0001f923\U0001f603\U0001f604", 3, 3},
		{"mixed latin cyrillic", "Hello Мир!", 6, 6},
		{"arabic", "مرحبا بالعالم", 5, 5},
		{"only spaces", "     ", 3, 3},
		{"only newlines", "\n\n\n\n\n", 3, 3},
		{"single emoji", "\U0001f600", 5, 1},
		{"cyrillic long", "Привет мир как дела сегодня", 6, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := entity.NewsItem{Content: tt.content}
			got := n.Summary(tt.maxRunes)
			gotRunes := []rune(got)
			if len(gotRunes) > tt.maxRunes {
				t.Errorf("Summary(%d) on %q returned %d runes", tt.maxRunes, tt.content, len(gotRunes))
			}
		})
	}
}

func TestUnitNewsItemValidateFieldCombinations(t *testing.T) {
	tests := []struct {
		name    string
		news    entity.NewsItem
		wantErr error
	}{
		{"valid title+content", entity.NewsItem{Title: "T", Content: "C"}, nil},
		{"valid with author", entity.NewsItem{Title: "T", Content: "C", Author: "A"}, nil},
		{"empty title valid content", entity.NewsItem{Title: "", Content: "C"}, entity.ErrEmptyName},
		{"valid title empty content", entity.NewsItem{Title: "T", Content: ""}, entity.ErrEmptyContent},
		{"empty title empty content", entity.NewsItem{Title: "", Content: ""}, entity.ErrEmptyName},
		{"empty title with author", entity.NewsItem{Title: "", Content: "C", Author: "A"}, entity.ErrEmptyName},
		{"empty content with author", entity.NewsItem{Title: "T", Content: "", Author: "A"}, entity.ErrEmptyContent},
		{"all empty", entity.NewsItem{Title: "", Content: "", Author: ""}, entity.ErrEmptyName},
		{"all populated", entity.NewsItem{Title: "T", Content: "C", Author: "A"}, nil},
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

func TestUnitNewsItemSummaryExactBoundary(t *testing.T) {
	n := entity.NewsItem{Content: "1234567890"} // exactly 10 runes
	got10 := n.Summary(10)
	if got10 != "1234567890" {
		t.Errorf("Summary(10) = %q, want full content", got10)
	}
	got9 := n.Summary(9)
	if got9 != "123456..." {
		t.Errorf("Summary(9) = %q, want %q", got9, "123456...")
	}
}

func TestUnitNewsItemSummaryNegativeMaxRunes(t *testing.T) {
	n := entity.NewsItem{Content: "Hello"}
	// Negative maxRunes causes panic in current implementation (slice bounds out of range).
	// This is expected behavior — callers should not pass negative values.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative maxRunes, got none")
		}
	}()
	_ = n.Summary(-1)
}

func TestUnitNewsItemValidateContentPriorityOrder(t *testing.T) {
	n := entity.NewsItem{Title: "", Content: ""}
	err := n.Validate()
	if err != entity.ErrEmptyName {
		t.Errorf("expected ErrEmptyName (title checked first), got %v", err)
	}
}
