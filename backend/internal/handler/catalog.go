package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/doordostyk/api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) ListCategories(c *gin.Context) {
	rows, err := h.pool.Query(c, `
		SELECT c.category_id, c.category_name, c.category_description,
			NOT EXISTS (SELECT 1 FROM product p WHERE p.category_id = c.category_id) AS can_delete
		FROM category c
		ORDER BY c.category_name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Category{}
	for rows.Next() {
		var x model.Category
		if err := rows.Scan(&x.ID, &x.Name, &x.Description, &x.CanDelete); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, x)
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ListProducts(c *gin.Context) {
	categoryID := c.Query("category_id")
	q := strings.TrimSpace(c.Query("q"))

	args := []any{}
	sb := strings.Builder{}
	sb.WriteString(`SELECT p.product_id, p.category_id, c.category_name, p.product_name,
		p.product_description, p.product_dimensions, p.product_purchase_price, p.product_retail_price,
		COALESCE(v.balance, 0),
		NOT EXISTS (SELECT 1 FROM receipt r WHERE r.product_id = p.product_id)
		AND NOT EXISTS (SELECT 1 FROM sale s WHERE s.product_id = p.product_id)
		AND NOT EXISTS (SELECT 1 FROM "order" o WHERE o.product_id = p.product_id) AS can_delete
		FROM product p
		JOIN category c ON c.category_id = p.category_id
		LEFT JOIN v_stock_balance v ON v.product_id = p.product_id
		WHERE 1=1 `)
	if categoryID != "" {
		args = append(args, categoryID)
		sb.WriteString(" AND p.category_id = $" + strconv.Itoa(len(args)))
	}
	if q != "" {
		args = append(args, "%"+strings.ToLower(q)+"%")
		n := strconv.Itoa(len(args))
		sb.WriteString(" AND (LOWER(p.product_name) LIKE $" + n + " OR LOWER(c.category_name) LIKE $" + n + ")")
	}
	sb.WriteString(" ORDER BY c.category_name, p.product_name")

	rows, err := h.pool.Query(c, sb.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Product{}
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.CategoryName, &p.Name,
			&p.Description, &p.Dimensions, &p.PurchasePrice, &p.RetailPrice, &p.Balance, &p.CanDelete); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, p)
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) GetProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p model.Product
	err := h.pool.QueryRow(c, `
		SELECT p.product_id, p.category_id, c.category_name, p.product_name,
			p.product_description, p.product_dimensions, p.product_purchase_price, p.product_retail_price,
			COALESCE(v.balance, 0),
			NOT EXISTS (SELECT 1 FROM receipt r WHERE r.product_id = p.product_id)
			AND NOT EXISTS (SELECT 1 FROM sale s WHERE s.product_id = p.product_id)
			AND NOT EXISTS (SELECT 1 FROM "order" o WHERE o.product_id = p.product_id) AS can_delete
		FROM product p JOIN category c ON c.category_id = p.category_id
		LEFT JOIN v_stock_balance v ON v.product_id = p.product_id
		WHERE p.product_id = $1`, id).Scan(&p.ID, &p.CategoryID, &p.CategoryName, &p.Name,
		&p.Description, &p.Dimensions, &p.PurchasePrice, &p.RetailPrice, &p.Balance, &p.CanDelete)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) Stock(c *gin.Context) {
	maxBalance := c.Query("max_balance")
	categoryID := c.Query("category_id")

	args := []any{}
	sb := strings.Builder{}
	sb.WriteString(`SELECT product_id, product_name, product_dimensions, product_retail_price,
		category_id, category_name, received_qty, sold_qty, balance
		FROM v_stock_balance WHERE 1=1 `)
	if maxBalance != "" {
		args = append(args, maxBalance)
		sb.WriteString(" AND balance <= $" + strconv.Itoa(len(args)))
	}
	if categoryID != "" {
		args = append(args, categoryID)
		sb.WriteString(" AND category_id = $" + strconv.Itoa(len(args)))
	}
	sb.WriteString(" ORDER BY balance ASC, product_name")

	rows, err := h.pool.Query(c, sb.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.StockItem{}
	for rows.Next() {
		var s model.StockItem
		if err := rows.Scan(&s.ProductID, &s.ProductName, &s.Dimensions, &s.RetailPrice,
			&s.CategoryID, &s.CategoryName, &s.ReceivedQty, &s.SoldQty, &s.Balance); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, s)
	}
	c.JSON(http.StatusOK, out)
}
