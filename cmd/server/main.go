package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikita-rulenko/heisenbug-portal/internal/handler"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
	"github.com/nikita-rulenko/heisenbug-portal/internal/usecase"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "data/portal.db", "SQLite database path")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	db, err := sqliteRepo.Open(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := sqliteRepo.RunMigrations(db); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	if err := sqliteRepo.SeedData(db); err != nil {
		log.Fatalf("seed data: %v", err)
	}

	productRepo := sqliteRepo.NewProductRepo(db)
	categoryRepo := sqliteRepo.NewCategoryRepo(db)
	newsRepo := sqliteRepo.NewNewsRepo(db)
	orderRepo := sqliteRepo.NewOrderRepo(db)

	productUC := usecase.NewProductUseCase(productRepo, categoryRepo)
	categoryUC := usecase.NewCategoryUseCase(categoryRepo)
	newsUC := usecase.NewNewsUseCase(newsRepo)
	orderUC := usecase.NewOrderUseCase(orderRepo, productRepo)

	productH := handler.NewProductHandler(productUC)
	categoryH := handler.NewCategoryHandler(categoryUC)
	newsH := handler.NewNewsHandler(newsUC)
	orderH := handler.NewOrderHandler(orderUC)
	pageH := handler.NewPageHandler(productUC, categoryUC, newsUC, "templates")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Get("/", pageH.Home)
	r.Get("/catalog", pageH.Catalog)
	r.Get("/news", pageH.NewsFeed)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/products", productH.Routes())
		r.Mount("/categories", categoryH.Routes())
		r.Mount("/news", newsH.Routes())
		r.Mount("/orders", orderH.Routes())
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("Starting Bean & Brew portal on %s", *addr)
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
