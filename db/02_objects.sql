-- Хранимые объекты: функция остатка, процедура, триггеры, представления

-- Вспомогательная функция: остаток товара на складе
CREATE OR REPLACE FUNCTION fn_product_balance(
    p_product_id INTEGER,
    p_exclude_sale_id INTEGER DEFAULT NULL
)
RETURNS INTEGER
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE((
        SELECT SUM(r.receipt_quantity)
        FROM receipt r
        WHERE r.product_id = p_product_id
    ), 0) - COALESCE((
        SELECT SUM(s.sale_quantity)
        FROM sale s
        WHERE s.product_id = p_product_id
        AND (p_exclude_sale_id IS NULL OR s.sale_id <> p_exclude_sale_id)
    ), 0) - COALESCE((
        SELECT SUM(o.order_quantity)
        FROM "order" o
        WHERE o.product_id = p_product_id
        AND o.order_status = 'новый'
    ), 0);
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
    SELECT COALESCE((
        SELECT SUM(r.receipt_quantity)
        FROM receipt r
        WHERE r.product_id = p_product_id
    ), 0) - COALESCE((
        SELECT SUM(s.sale_quantity)
        FROM sale s
        WHERE s.product_id = p_product_id
        AND (p_exclude_sale_id IS NULL OR s.sale_id <> p_exclude_sale_id)
    ), 0) - COALESCE((
        SELECT SUM(o.order_quantity)
        FROM "order" o
        WHERE o.product_id = p_product_id
        AND o.order_status = 'новый'
        AND (p_exclude_order_id IS NULL OR o.order_id <> p_exclude_order_id)
    ), 0);
$$;

-- 2.8.1. Процедура выполнения заказа
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
$$;

-- Контроль остатка при создании заказа (статус «новый»)
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

DROP TRIGGER IF EXISTS trg_order_check_stock ON "order";

CREATE TRIGGER trg_order_check_stock
BEFORE INSERT OR UPDATE OF product_id, order_quantity, order_status ON "order"
FOR EACH ROW
EXECUTE PROCEDURE trg_fn_order_check_stock();

-- 2.8.2. Триггер № 1: контроль остатка при продаже
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

DROP TRIGGER IF EXISTS trg_sale_check_stock ON sale;

CREATE TRIGGER trg_sale_check_stock
BEFORE INSERT OR UPDATE OF product_id, sale_quantity ON sale
FOR EACH ROW
EXECUTE PROCEDURE trg_fn_sale_check_stock();

-- 2.8.3. Триггер № 2: защита статуса заказа
CREATE OR REPLACE FUNCTION trg_fn_order_status_guard()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP <> 'UPDATE' THEN
        RETURN NEW;
    END IF;

    IF OLD.order_status = 'выполнен' AND NEW.order_status IS DISTINCT FROM 'выполнен' THEN
        RAISE EXCEPTION 'Нельзя изменить статус выполненного заказа %', OLD.order_id USING ERRCODE = 'P0001';
    END IF;

    IF NEW.order_status NOT IN ('новый', 'выполнен', 'отклонён') THEN
        RAISE EXCEPTION 'Недопустимый статус заказа: %', NEW.order_status USING ERRCODE = 'P0001';
    END IF;

    IF NEW.order_status = 'выполнен' AND NEW.sale_id IS NULL THEN
        RAISE EXCEPTION 'Для статуса «выполнен» необходимо указать sale_id' USING ERRCODE = 'P0001';
    END IF;

    IF NEW.order_status = 'отклонён' AND NEW.sale_id IS NOT NULL THEN
        RAISE EXCEPTION 'Отклонённый заказ не должен иметь sale_id' USING ERRCODE = 'P0001';
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_order_status_guard ON "order";

CREATE TRIGGER trg_order_status_guard
BEFORE UPDATE OF order_status, sale_id ON "order"
FOR EACH ROW
EXECUTE PROCEDURE trg_fn_order_status_guard();

-- Представления (для отчётов)

-- VIEW 1: продажи по категориям (детальная)
CREATE OR REPLACE VIEW v_sales_by_category AS
SELECT
    s.sale_id,
    s.sale_date,
    c.category_id,
    c.category_name,
    p.product_id,
    p.product_name,
    p.product_dimensions,
    s.sale_quantity,
    s.sale_price,
    (s.sale_quantity * s.sale_price)::DECIMAL(14, 2) AS sale_amount,
    u.user_id,
    u.user_full_name AS seller_name
FROM sale s
JOIN product p ON p.product_id = s.product_id
JOIN category c ON c.category_id = p.category_id
JOIN "user" u ON u.user_id = s.user_id;

-- VIEW 2: остатки на складе
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
