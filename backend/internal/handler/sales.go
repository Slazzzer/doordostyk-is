package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/doordostyk/api/internal/middleware"
	"github.com/doordostyk/api/internal/model"
	"github.com/gin-gonic/gin"
)

type createSaleReq struct {
	ProductID int     `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity"   binding:"required,min=1"`
	Price     float64 `json:"price"`
}

func (h *Handler) CreateSale(c *gin.Context) {
	var req createSaleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity > 99999 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Количество не должно превышать 99 999"})
		return
	}
	cl := middleware.CurrentClaims(c)

	price := req.Price
	if price > 2147483647 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Цена не должна превышать 2 147 483 647"})
		return
	}
	if price <= 0 {
		err := h.pool.QueryRow(c, `SELECT COALESCE(product_retail_price,0) FROM product WHERE product_id=$1`,
			req.ProductID).Scan(&price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Товар не найден"})
			return
		}
		if price <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Для товара не задана розничная цена"})
			return
		}
	}

	var id int
	err := h.pool.QueryRow(c, `
		INSERT INTO sale(product_id, user_id, sale_date, sale_quantity, sale_price)
		VALUES ($1,$2,CURRENT_DATE,$3,$4) RETURNING sale_id`,
		req.ProductID, cl.Subject, req.Quantity, price).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"sale_id": id})
}

func (h *Handler) ListSales(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	args := []any{}
	sb := strings.Builder{}
	sb.WriteString(`SELECT s.sale_id, s.order_id, s.product_id, p.product_name, s.user_id, u.user_full_name,
		s.sale_date, s.sale_quantity, s.sale_price, (s.sale_quantity*s.sale_price)::DECIMAL(14,2)
		FROM sale s
		JOIN product p ON p.product_id=s.product_id
		JOIN "user"  u ON u.user_id=s.user_id WHERE 1=1 `)
	if from != "" {
		args = append(args, from)
		sb.WriteString(" AND s.sale_date >= $" + strconv.Itoa(len(args)))
	}
	if to != "" {
		args = append(args, to)
		sb.WriteString(" AND s.sale_date <= $" + strconv.Itoa(len(args)))
	}
	sb.WriteString(" ORDER BY s.sale_date DESC, s.sale_id DESC")
	rows, err := h.pool.Query(c, sb.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Sale{}
	for rows.Next() {
		var s model.Sale
		if err := rows.Scan(&s.ID, &s.OrderID, &s.ProductID, &s.ProductName,
			&s.UserID, &s.UserName, &s.Date, &s.Quantity, &s.Price, &s.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, s)
	}
	c.JSON(http.StatusOK, out)
}
