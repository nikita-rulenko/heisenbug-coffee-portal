package sqlite

import "database/sql"

func SeedData(db *sql.DB) error {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if count > 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	categories := []struct {
		name, slug, description string
		sortOrder               int
	}{
		{"Эспрессо", "espresso", "Классические напитки на основе эспрессо", 1},
		{"Фильтр-кофе", "filter", "Альтернативные способы заваривания", 2},
		{"Холодные напитки", "cold", "Освежающие холодные кофейные напитки", 3},
		{"Зерно", "beans", "Свежеобжаренное кофейное зерно", 4},
		{"Аксессуары", "accessories", "Всё для приготовления кофе дома", 5},
	}

	for _, c := range categories {
		_, err := tx.Exec(
			"INSERT INTO categories (name, slug, description, sort_order) VALUES (?, ?, ?, ?)",
			c.name, c.slug, c.description, c.sortOrder,
		)
		if err != nil {
			return err
		}
	}

	products := []struct {
		catID                int
		name, desc, imageURL string
		price                float64
	}{
		{1, "Американо", "Двойной эспрессо с горячей водой", "/static/img/americano.jpg", 250},
		{1, "Капучино", "Эспрессо с молочной пенкой", "/static/img/cappuccino.jpg", 320},
		{1, "Латте", "Эспрессо с большим количеством молока", "/static/img/latte.jpg", 350},
		{1, "Флэт Уайт", "Двойной ристретто с бархатным молоком", "/static/img/flatwhite.jpg", 380},
		{1, "Раф", "Эспрессо со сливками и ванильным сахаром", "/static/img/raf.jpg", 400},
		{2, "V60 Эфиопия Иргачеффе", "Яркий, цветочный, цитрусовый", "/static/img/v60.jpg", 350},
		{2, "Аэропресс Колумбия", "Сбалансированный, шоколадный", "/static/img/aeropress.jpg", 320},
		{2, "Кемекс Кения АА", "Насыщенный, ягодный", "/static/img/chemex.jpg", 380},
		{3, "Колд Брю", "Кофе холодного заваривания, 12 часов", "/static/img/coldbrew.jpg", 300},
		{3, "Айс Латте", "Эспрессо с молоком и льдом", "/static/img/icelatte.jpg", 350},
		{3, "Бамбл", "Эспрессо с апельсиновым соком", "/static/img/bumble.jpg", 380},
		{4, "Эфиопия Сидамо 250г", "Средняя обжарка, ноты черники и жасмина", "/static/img/beans-ethiopia.jpg", 1200},
		{4, "Бразилия Сантос 250г", "Тёмная обжарка, шоколад и орех", "/static/img/beans-brazil.jpg", 980},
		{4, "Колумбия Супремо 250г", "Средняя обжарка, карамель и цитрус", "/static/img/beans-colombia.jpg", 1100},
		{5, "Гейзерная кофеварка Bialetti", "Классическая мока на 3 чашки", "/static/img/bialetti.jpg", 4500},
		{5, "Ручная кофемолка Timemore C2", "Жерновая, 38мм", "/static/img/grinder.jpg", 6500},
		{5, "Френч-пресс 350мл", "Стекло и нержавейка", "/static/img/frenchpress.jpg", 1800},
	}

	for _, p := range products {
		_, err := tx.Exec(
			"INSERT INTO products (category_id, name, description, price, image_url, in_stock) VALUES (?, ?, ?, ?, ?, 1)",
			p.catID, p.name, p.desc, p.price, p.imageURL,
		)
		if err != nil {
			return err
		}
	}

	news := []struct {
		title, content, author string
	}{
		{
			"Новое зерно: Эфиопия Гуджи",
			"Мы привезли потрясающий лот из региона Гуджи — натуральная обработка, ноты персика, манго и тёмного шоколада. Доступно ограниченное количество!",
			"Бариста Алексей",
		},
		{
			"Мастер-класс по латте-арту",
			"В эту субботу в 15:00 проводим бесплатный мастер-класс по латте-арту. Научим рисовать розетту и тюльпан. Запись в директ!",
			"Команда Bean & Brew",
		},
		{
			"Летнее меню 2026",
			"Представляем новое летнее меню: тоник-эспрессо, лавандовый колд брю и ягодный бамбл. Попробуйте все три и получите стикер в подарок!",
			"Шеф-бариста Мария",
		},
	}

	for _, n := range news {
		_, err := tx.Exec(
			"INSERT INTO news (title, content, author) VALUES (?, ?, ?)",
			n.title, n.content, n.author,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
