package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		if err := pool.Ping(deadline); err == nil {
			return pool, nil
		} else {
			log.Printf("db not ready, retry: %v", err)
		}
		select {
		case <-deadline.Done():
			return nil, fmt.Errorf("db ping timeout")
		case <-time.After(2 * time.Second):
		}
	}
}

// SeedPasswords восстанавливает обязательных сотрудников и задаёт им пароли.
func SeedPasswords(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `UPDATE "user" SET user_login = 'seller' WHERE user_login = 'seller1'`); err != nil {
		return fmt.Errorf("migrate seller login: %w", err)
	}

	staff := []struct {
		fullName string
		login    string
		role     string
		pwd      string
	}{
		{"Гусев Иван Михайлович", "admin", "administrator", "admin123"},
		{"Белослудцев Никита Викторович", "seller", "seller", "seller123"},
		{"Осипов Павел Дмитриевич", "storekeeper", "storekeeper", "store123"},
	}
	for _, s := range staff {
		b, err := bcrypt.GenerateFromPassword([]byte(s.pwd), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		tag, err := pool.Exec(ctx, `
			INSERT INTO "user"(user_full_name, user_login, user_role, user_password_hash)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_login) DO UPDATE
			SET user_full_name = EXCLUDED.user_full_name,
			    user_role = EXCLUDED.user_role,
			    user_password_hash = EXCLUDED.user_password_hash`,
			s.fullName, s.login, s.role, string(b))
		if err != nil {
			return fmt.Errorf("seed password %s: %w", s.login, err)
		}
		if tag.RowsAffected() > 0 {
			log.Printf("seed: ensured user %s", s.login)
		}
	}
	return nil
}

func EnsureSuppliers(ctx context.Context, pool *pgxpool.Pool) error {
	// Только переименование устаревших записей; новые поставщики — из db/03_seed.sql
	renames := map[string]string{
		"ООО BRAVO DOORS": `ООО "BRAVO DOORS"`,
		"ООО СОМ":         `ООО "СОМ"`,
		"ИП Гусев М.В.":   `ООО "ДекорПлюс"`,
		"ИП Воронин А.И.": `ООО "ДекорПлюс"`,
	}
	for oldName, newName := range renames {
		if _, err := pool.Exec(ctx, `
			UPDATE supplier SET organization_name = $1
			WHERE organization_name = $2
			AND NOT EXISTS (SELECT 1 FROM supplier WHERE organization_name = $1)`,
			newName, oldName); err != nil {
			return fmt.Errorf("rename supplier %s: %w", oldName, err)
		}
	}
	// Удалить дубликаты поставщиков (оставить запись с минимальным id)
	_, err := pool.Exec(ctx, `
		DELETE FROM supplier s
		USING supplier d
		WHERE s.organization_name = d.organization_name
		  AND s.supplier_id > d.supplier_id`)
	if err != nil {
		return fmt.Errorf("dedupe suppliers: %w", err)
	}
	return nil
}

// EnsureExtraProducts добавляет 3 товара для «квадрата» каталога, если их ещё нет (старые тома БД).
func EnsureExtraProducts(ctx context.Context, pool *pgxpool.Pool) error {
	type extra struct {
		categoryID int
		name       string
		desc       string
		dims       *string
		purchase   float64
		retail     float64
		supplierID int
		qty        int
	}
	items := []extra{
		{1, "Дверь входная «Сигма»", "Порошковая окраска, утеплитель 30 мм", strPtr("2050×880"), 12500, 16800, 1, 10},
		{2, "Дверь межкомнатная «Альфа»", "Экошпон, цвет орех", strPtr("2000×800"), 6100, 8200, 2, 14},
		{4, "Короб дверной «Стандарт»", "МДФ, комплект 3 шт.", strPtr("2100 мм"), 950, 1350, 3, 25},
	}
	var storekeeperID int
	err := pool.QueryRow(ctx, `SELECT user_id FROM "user" WHERE user_role = 'storekeeper' LIMIT 1`).Scan(&storekeeperID)
	if err != nil {
		return nil
	}
	for _, it := range items {
		var productID int
		err := pool.QueryRow(ctx, `
			INSERT INTO product (category_id, product_name, product_description, product_dimensions,
			                     product_purchase_price, product_retail_price)
			SELECT $1::integer, $2::varchar, $3::text, $4::varchar, $5::decimal, $6::decimal
			WHERE NOT EXISTS (SELECT 1 FROM product WHERE product_name = $2)
			RETURNING product_id`,
			it.categoryID, it.name, it.desc, it.dims, it.purchase, it.retail).Scan(&productID)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				continue
			}
			return fmt.Errorf("extra product %s: %w", it.name, err)
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO receipt (supplier_id, user_id, product_id, receipt_date, receipt_quantity, receipt_purchase_price)
			SELECT $1, $2, $3, CURRENT_DATE, $4, $5
			WHERE NOT EXISTS (
				SELECT 1 FROM receipt r WHERE r.product_id = $3 AND r.receipt_date = CURRENT_DATE
			)`,
			it.supplierID, storekeeperID, productID, it.qty, it.purchase)
		if err != nil {
			return fmt.Errorf("extra receipt %s: %w", it.name, err)
		}
		log.Printf("seed: ensured catalog product %s", it.name)
	}
	return nil
}

// EnsureDecorationCategory добавляет категорию «Декор для дверей» и товары (старые тома БД).
func EnsureDecorationCategory(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO category (category_name, category_description)
		SELECT 'Декор для дверей', 'Украшения, таблички, венки и аксессуары для оформления двери'
		WHERE NOT EXISTS (SELECT 1 FROM category WHERE category_name = 'Декор для дверей')`)
	if err != nil {
		return fmt.Errorf("ensure decor category: %w", err)
	}
	var categoryID int
	if err := pool.QueryRow(ctx, `SELECT category_id FROM category WHERE category_name = 'Декор для дверей' LIMIT 1`).Scan(&categoryID); err != nil {
		return nil
	}
	type extra struct {
		name     string
		desc     string
		dims     *string
		purchase float64
		retail   float64
		supplier int
		qty      int
	}
	items := []extra{
		{"Венок на дверь «Новогодний»", "Искусственная хвоя, красные ягоды", strPtr("Ø 35 см"), 890, 1290, 4, 40},
		{"Табличка на дверь «Добро пожаловать»", "Металл, крепление на липучке", strPtr("20×8 см"), 420, 650, 4, 35},
		{"Молоток дверной «Классик»", "Латунь, полированная поверхность", nil, 1850, 2650, 3, 15},
		{"Номер квартиры «Современный»", "Акрил, цифры 0–9 в комплекте", strPtr("12×4 см"), 310, 490, 4, 50},
		{"Наклейка на дверь «Цветы»", "Винил, съёмная без следов", strPtr("30×60 см"), 180, 290, 2, 80},
	}
	var storekeeperID int
	if err := pool.QueryRow(ctx, `SELECT user_id FROM "user" WHERE user_role = 'storekeeper' LIMIT 1`).Scan(&storekeeperID); err != nil {
		return nil
	}
	for _, it := range items {
		var productID int
		err := pool.QueryRow(ctx, `
			INSERT INTO product (category_id, product_name, product_description, product_dimensions,
			                     product_purchase_price, product_retail_price)
			SELECT $1::integer, $2::varchar, $3::text, $4::varchar, $5::decimal, $6::decimal
			WHERE NOT EXISTS (SELECT 1 FROM product WHERE product_name = $2)
			RETURNING product_id`,
			categoryID, it.name, it.desc, it.dims, it.purchase, it.retail).Scan(&productID)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				continue
			}
			return fmt.Errorf("decor product %s: %w", it.name, err)
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO receipt (supplier_id, user_id, product_id, receipt_date, receipt_quantity, receipt_purchase_price)
			SELECT $1, $2, $3, CURRENT_DATE, $4, $5
			WHERE NOT EXISTS (
				SELECT 1 FROM receipt r WHERE r.product_id = $3 AND r.receipt_date = CURRENT_DATE
			)`,
			it.supplier, storekeeperID, productID, it.qty, it.purchase)
		if err != nil {
			return fmt.Errorf("decor receipt %s: %w", it.name, err)
		}
		log.Printf("seed: ensured decor product %s", it.name)
	}
	return nil
}

func EnsureNonNegativeBalances(ctx context.Context, pool *pgxpool.Pool) error {
	var storekeeperID int
	if err := pool.QueryRow(ctx, `SELECT user_id FROM "user" WHERE user_role = 'storekeeper' LIMIT 1`).Scan(&storekeeperID); err != nil {
		return nil
	}
	var supplierID int
	if err := pool.QueryRow(ctx, `SELECT supplier_id FROM supplier ORDER BY supplier_id LIMIT 1`).Scan(&supplierID); err != nil {
		return nil
	}
	rows, err := pool.Query(ctx, `SELECT product_id, product_name, balance FROM v_stock_balance WHERE balance < 0`)
	if err != nil {
		return fmt.Errorf("select negative balances: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var productID, balance int
		var name string
		if err := rows.Scan(&productID, &name, &balance); err != nil {
			return err
		}
		qty := -balance + 5
		_, err := pool.Exec(ctx, `
			INSERT INTO receipt(supplier_id, user_id, product_id, receipt_date, receipt_quantity, receipt_purchase_price)
			SELECT $1, $2, $3, CURRENT_DATE, $4, COALESCE(product_purchase_price, 0)
			FROM product
			WHERE product_id = $3`,
			supplierID, storekeeperID, productID, qty)
		if err != nil {
			return fmt.Errorf("fix balance %s: %w", name, err)
		}
		log.Printf("seed: fixed negative balance for %s by +%d", name, qty)
	}
	return rows.Err()
}

func EnsureStockReservationObjects(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE OR REPLACE FUNCTION fn_product_balance(
    p_product_id INTEGER,
    p_exclude_sale_id INTEGER DEFAULT NULL
)
RETURNS INTEGER
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE((SELECT SUM(receipt_quantity) FROM receipt WHERE product_id = p_product_id), 0)
         - COALESCE((SELECT SUM(sale_quantity) FROM sale WHERE product_id = p_product_id
             AND (p_exclude_sale_id IS NULL OR sale_id <> p_exclude_sale_id)), 0)
         - COALESCE((SELECT SUM(order_quantity) FROM "order" WHERE product_id = p_product_id
             AND order_status = 'новый'), 0);
$$;

CREATE OR REPLACE FUNCTION fn_product_balance_available(
    p_product_id INTEGER,
    p_exclude_sale_id INTEGER DEFAULT NULL,
    p_exclude_order_id INTEGER DEFAULT NULL
)
RETURNS INTEGER
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE((SELECT SUM(receipt_quantity) FROM receipt WHERE product_id = p_product_id), 0)
         - COALESCE((SELECT SUM(sale_quantity) FROM sale WHERE product_id = p_product_id
             AND (p_exclude_sale_id IS NULL OR sale_id <> p_exclude_sale_id)), 0)
         - COALESCE((SELECT SUM(order_quantity) FROM "order" WHERE product_id = p_product_id
             AND order_status = 'новый'
             AND (p_exclude_order_id IS NULL OR order_id <> p_exclude_order_id)), 0);
$$;

DROP VIEW IF EXISTS v_stock_balance;

CREATE OR REPLACE VIEW v_stock_balance AS
SELECT
    p.product_id,
    p.product_name,
    p.product_dimensions,
    p.product_retail_price,
    c.category_id,
    c.category_name,
    COALESCE(r_sum.received_qty, 0) AS received_qty,
    COALESCE(s_sum.sold_qty, 0) AS sold_qty,
    COALESCE(o_sum.reserved_qty, 0) AS reserved_qty,
    COALESCE(r_sum.received_qty, 0) - COALESCE(s_sum.sold_qty, 0) - COALESCE(o_sum.reserved_qty, 0) AS balance
FROM product p
JOIN category c ON c.category_id = p.category_id
LEFT JOIN (
    SELECT product_id, SUM(receipt_quantity) AS received_qty
    FROM receipt
    GROUP BY product_id
) r_sum ON r_sum.product_id = p.product_id
LEFT JOIN (
    SELECT product_id, SUM(sale_quantity) AS sold_qty
    FROM sale
    GROUP BY product_id
) s_sum ON s_sum.product_id = p.product_id
LEFT JOIN (
    SELECT product_id, SUM(order_quantity) AS reserved_qty
    FROM "order"
    WHERE order_status = 'новый'
    GROUP BY product_id
) o_sum ON o_sum.product_id = p.product_id;

CREATE OR REPLACE FUNCTION trg_fn_order_check_stock()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_balance INTEGER;
BEGIN
    IF NEW.order_status = 'новый' THEN
        v_balance := fn_product_balance_available(
            NEW.product_id,
            NULL,
            CASE WHEN TG_OP = 'UPDATE' THEN OLD.order_id ELSE NULL END
        );
        IF v_balance < NEW.order_quantity THEN
            RAISE EXCEPTION 'Недостаточно товара на складе для заказа: остаток=%, запрошено=%',
                v_balance, NEW.order_quantity USING ERRCODE = 'P0001';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION trg_fn_sale_check_stock()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_balance INTEGER;
    v_exclude INTEGER;
BEGIN
    IF NEW.sale_quantity IS NULL OR NEW.sale_quantity <= 0 THEN
        RAISE EXCEPTION 'Количество продажи должно быть положительным' USING ERRCODE = 'P0001';
    END IF;
    v_exclude := CASE WHEN TG_OP = 'UPDATE' THEN OLD.sale_id ELSE NULL END;
    v_balance := fn_product_balance_available(NEW.product_id, v_exclude, NEW.order_id);
    IF v_balance < NEW.sale_quantity THEN
        RAISE EXCEPTION 'Продажа отклонена: остаток=%, запрошено=%', v_balance, NEW.sale_quantity USING ERRCODE = 'P0001';
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE PROCEDURE sp_execute_order(
    IN p_order_id INTEGER,
    IN p_user_id INTEGER,
    OUT p_sale_id INTEGER
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_product_id INTEGER;
    v_quantity INTEGER;
    v_status VARCHAR(30);
    v_price DECIMAL(12, 2);
    v_balance INTEGER;
BEGIN
    SELECT o.product_id, o.order_quantity, o.order_status
    INTO v_product_id, v_quantity, v_status
    FROM "order" o
    WHERE o.order_id = p_order_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Заказ с id=% не найден', p_order_id USING ERRCODE = 'P0002';
    END IF;
    IF v_status IS DISTINCT FROM 'новый' THEN
        RAISE EXCEPTION 'Заказ % уже обработан (статус: %)', p_order_id, v_status USING ERRCODE = 'P0001';
    END IF;
    v_balance := fn_product_balance_available(v_product_id, NULL, p_order_id);
    IF v_balance < v_quantity THEN
        RAISE EXCEPTION 'Недостаточно товара на складе: остаток=%, требуется=%', v_balance, v_quantity USING ERRCODE = 'P0001';
    END IF;
    SELECT p.product_retail_price INTO v_price FROM product p WHERE p.product_id = v_product_id;
    IF v_price IS NULL THEN
        RAISE EXCEPTION 'Не задана розничная цена для товара id=%', v_product_id USING ERRCODE = 'P0001';
    END IF;
    INSERT INTO sale (order_id, product_id, user_id, sale_date, sale_quantity, sale_price)
    VALUES (p_order_id, v_product_id, p_user_id, CURRENT_DATE, v_quantity, v_price)
    RETURNING sale.sale_id INTO p_sale_id;
    UPDATE "order" o
    SET order_status = 'выполнен', sale_id = p_sale_id
    WHERE o.order_id = p_order_id;
END;
$$;`)
	if err != nil {
		return fmt.Errorf("ensure stock reservation objects: %w", err)
	}
	return nil
}

func strPtr(s string) *string { return &s }
