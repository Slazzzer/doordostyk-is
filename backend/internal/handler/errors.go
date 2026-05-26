package handler

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// mapDBError переводит ошибки PostgreSQL в понятные сообщения на русском.
func mapDBError(err error) string {
	if err == nil {
		return "неизвестная ошибка"
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		msg := strings.TrimSpace(pg.Message)
		if readable := stockErrorMessage(msg); readable != "" {
			return readable
		}
		if strings.Contains(msg, "Заказ") && strings.Contains(msg, "не найден") ||
			strings.Contains(msg, "уже обработан") {
			return msg
		}
		switch pg.Code {
		case "23001", "23503":
			cn := strings.ToLower(pg.ConstraintName)
			switch {
			case strings.Contains(cn, "fk_order_becomes_sale"):
				return "Нельзя удалить продажу: по ней оформлен заказ."
			case strings.Contains(cn, "fk_product_groups_category") || strings.Contains(cn, "category"):
				return "Нельзя удалить категорию: к ней привязаны товары."
			case strings.Contains(cn, "fk_receipt_ships_supplier") || strings.Contains(cn, "supplier"):
				return "Нельзя удалить поставщика: по нему уже есть поступления."
			case strings.Contains(cn, "fk_sale") || strings.Contains(cn, "sale"):
				return "Операция невозможна: запись связана с продажами."
			case strings.Contains(cn, "fk_receipt") || strings.Contains(cn, "receipt"):
				return "Операция невозможна: запись связана с поступлениями."
			case strings.Contains(cn, "fk_order") || strings.Contains(cn, "order"):
				return "Операция невозможна: запись связана с заказами."
			case strings.Contains(cn, "user"):
				return "Пользователь связан с операциями в системе и не может быть удалён."
			case strings.Contains(cn, "customer"):
				return "Клиент связан с заказами и не может быть удалён."
			default:
				return "Операция невозможна: есть связанные данные в базе."
			}
		case "23505":
			return "Такая запись уже существует."
		case "P0001", "P0002":
			if msg != "" {
				return msg
			}
		}
		if msg != "" && !strings.HasPrefix(strings.ToUpper(msg), "ERROR:") {
			return msg
		}
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "violates foreign key constraint") {
		if strings.Contains(lower, "fk_order_becomes_sale") {
			return "Нельзя удалить продажу: по ней оформлен заказ."
		}
		return "Операция невозможна: есть связанные данные в базе."
	}
	return "Не удалось выполнить операцию. Попробуйте ещё раз или обратитесь к администратору."
}

func stockErrorMessage(msg string) string {
	if !(strings.Contains(msg, "Недостаточно товара") || strings.Contains(msg, "Продажа отклонена")) {
		return ""
	}
	re := regexp.MustCompile(`(?:остаток|осталось)=?(-?\d+).*?(?:требуется|запрошено)=?(\d+)`)
	m := re.FindStringSubmatch(msg)
	if len(m) == 3 {
		return fmt.Sprintf("Недостаточно товара на складе: осталось %s шт., запрошено %s шт.", m[1], m[2])
	}
	return "Недостаточно товара на складе."
}
