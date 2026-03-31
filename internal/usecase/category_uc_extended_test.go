package usecase_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitCategoryUCCreateValidatesEmpty(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	c := &entity.Category{Name: "", Slug: "slug"}
	if err := categoryUC.Create(context.Background(), c); err != entity.ErrEmptyName {
		t.Errorf("Create empty name = %v, want ErrEmptyName", err)
	}
}

func TestUnitCategoryUCCreateValidatesSlug(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	c := &entity.Category{Name: "Кофе", Slug: ""}
	if err := categoryUC.Create(context.Background(), c); err != entity.ErrEmptySlug {
		t.Errorf("Create empty slug = %v, want ErrEmptySlug", err)
	}
}

func TestUnitCategoryUCUpdateValidatesName(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	c := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, c)

	c.Name = ""
	if err := categoryUC.Update(ctx, c); err != entity.ErrEmptyName {
		t.Errorf("Update empty name = %v, want ErrEmptyName", err)
	}
}

func TestUnitCategoryUCUpdateValidatesSlug(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	c := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, c)

	c.Slug = ""
	if err := categoryUC.Update(ctx, c); err != entity.ErrEmptySlug {
		t.Errorf("Update empty slug = %v, want ErrEmptySlug", err)
	}
}

func TestUnitCategoryUCDeleteNonExistent(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	err := categoryUC.Delete(context.Background(), 999)
	if err == nil {
		t.Error("expected error deleting non-existent category")
	}
}

func TestUnitCategoryUCGetByIDNonExistent(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	_, err := categoryUC.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("GetByID 999 = %v, want ErrNotFound", err)
	}
}

func TestUnitCategoryUCGetBySlugNonExistent(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	_, err := categoryUC.GetBySlug(context.Background(), "nope")
	if err != entity.ErrNotFound {
		t.Errorf("GetBySlug nope = %v, want ErrNotFound", err)
	}
}

func TestUnitCategoryUCListEmpty(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	cats, err := categoryUC.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(cats) != 0 {
		t.Errorf("expected 0, got %d", len(cats))
	}
}

func TestUnitCategoryUCListMultiple(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	for i := range 5 {
		c := &entity.Category{Name: "Cat" + string(rune('A'+i)), Slug: "cat-" + string(rune('a'+i))}
		categoryUC.Create(ctx, c)
	}
	cats, _ := categoryUC.List(ctx)
	if len(cats) != 5 {
		t.Errorf("count = %d, want 5", len(cats))
	}
}

func TestUnitCategoryUCCreateAndRetrieve(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	c := &entity.Category{Name: "Десерты", Slug: "desserts", Description: "Сладости", SortOrder: 3}
	categoryUC.Create(ctx, c)

	got, err := categoryUC.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Десерты" || got.Slug != "desserts" || got.SortOrder != 3 {
		t.Errorf("fields mismatch: %+v", got)
	}
}

func TestUnitCategoryUCCreateAndGetBySlug(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	c := &entity.Category{Name: "Чай", Slug: "tea"}
	categoryUC.Create(ctx, c)

	got, err := categoryUC.GetBySlug(ctx, "tea")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.Name != "Чай" {
		t.Errorf("name = %q, want Чай", got.Name)
	}
}

func TestUnitCategoryUCUpdateAllFields(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	c := &entity.Category{Name: "Old", Slug: "old", Description: "old", SortOrder: 1}
	categoryUC.Create(ctx, c)

	c.Name = "New"
	c.Slug = "new"
	c.Description = "new desc"
	c.SortOrder = 99
	categoryUC.Update(ctx, c)

	got, _ := categoryUC.GetByID(ctx, c.ID)
	if got.Name != "New" || got.Slug != "new" || got.Description != "new desc" || got.SortOrder != 99 {
		t.Errorf("not updated: %+v", got)
	}
}

func TestUnitCategoryUCUpdateNonExistent(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	c := &entity.Category{ID: 999, Name: "Ghost", Slug: "ghost"}
	err := categoryUC.Update(context.Background(), c)
	if err == nil {
		t.Error("expected error updating non-existent")
	}
}
