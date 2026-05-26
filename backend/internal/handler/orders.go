package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/doordostyk/api/internal/middleware"
	"github.com/doordostyk/api/internal/model"
	"github.com/gin-gonic/gin"
)

type createOrderReq struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity"   binding:"required,min=1"`
}

func (h *Handler) CreateOrderByCustomer(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity > 99999 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Количество не должно превышать 99 999"})
		return
	}
	cl := middleware.CurrentClaims(c)

	var balance int
	err := h.pool.QueryRow(c, `SELECT fn_product_balance($1)`, req.ProductID).Scan(&balance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "товар не найден"})
		return
	}
	if balance < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Недостаточно товара на складе: остаток=%d, запрошено=%d", balance, req.Quantity),
		})
		return
	}

	var id int
	err = h.pool.QueryRow(c, `
		INSERT INTO "order"(customer_id, product_id, order_date, order_quantity, order_status)
		VALUES ($1, $2, CURRENT_DATE, $3, 'новый')
		RETURNING order_id`, cl.Subject, req.ProductID, req.Quantity).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"order_id": id, "order_status": "новый"})
}

func (h *Handler) MyOrders(c *gin.Context) {
	cl := middleware.CurrentClaims(c)
	rows, err := h.pool.Query(c, `
		SELECT o.order_id, o.customer_id, '' AS customer_name, o.product_id, p.product_name,
			o.sale_id, o.order_date, o.order_quantity, o.order_status
		FROM "order" o
		JOIN product p ON p.product_id = o.product_id
		WHERE o.customer_id = $1
		ORDER BY o.order_date DESC, o.order_id DESC`, cl.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Order{}
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.CustomerName, &o.ProductID, &o.ProductName,
			&o.SaleID, &o.Date, &o.Quantity, &o.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, o)
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ListOrders(c *gin.Context) {
	status := c.Query("status")
	args := []any{}
	sb := strings.Builder{}
	sb.WriteString(`SELECT o.order_id, o.customer_id, cu.customer_full_name, o.product_id, p.product_name,
		o.sale_id, o.order_date, o.order_quantity, o.order_status
		FROM "order" o
		JOIN product  p  ON p.product_id  = o.product_id
		JOIN customer cu ON cu.customer_id = o.customer_id WHERE 1=1 `)
	if status != "" {
		args = append(args, status)
		sb.WriteString(" AND o.order_status = $" + strconv.Itoa(len(args)))
	}
	sb.WriteString(" ORDER BY o.order_date DESC, o.order_id DESC")

	rows, err := h.pool.Query(c, sb.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Order{}
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.CustomerName, &o.ProductID, &o.ProductName,
			&o.SaleID, &o.Date, &o.Quantity, &o.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, o)
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ExecuteOrder(c *gin.Context) {
	orderID, _ := strconv.Atoi(c.Param("id"))
	cl := middleware.CurrentClaims(c)
	var saleID int
	err := h.pool.QueryRow(c,
		`CALL sp_execute_order($1, $2, NULL)`,
		orderID, cl.Subject).Scan(&saleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order_id": orderID, "sale_id": saleID, "order_status": "выполнен"})
}

func (h *Handler) RejectOrder(c *gin.Context) {
	orderID, _ := strconv.Atoi(c.Param("id"))
	res, err := h.pool.Exec(c,
		`UPDATE "order" SET order_status='отклонён' WHERE order_id=$1 AND order_status='новый'`, orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "заказ уже обработан или не найден"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order_id": orderID, "order_status": "отклонён"})
}
