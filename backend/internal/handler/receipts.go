package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/doordostyk/api/internal/middleware"
	"github.com/doordostyk/api/internal/model"
	"github.com/gin-gonic/gin"
)

type createReceiptReq struct {
	SupplierID    int     `json:"supplier_id"    binding:"required"`
	ProductID     int     `json:"product_id"     binding:"required"`
	Quantity      int     `json:"quantity"       binding:"required,min=1"`
	PurchasePrice float64 `json:"purchase_price" binding:"required,gte=0"`
}

func (h *Handler) CreateReceipt(c *gin.Context) {
	var req createReceiptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity > 99999 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Количество не должно превышать 99 999"})
		return
	}
	if req.PurchasePrice > 2147483647 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Цена не должна превышать 2 147 483 647"})
		return
	}
	cl := middleware.CurrentClaims(c)
	var id int
	err := h.pool.QueryRow(c, `
		INSERT INTO receipt(supplier_id, user_id, product_id, receipt_date, receipt_quantity, receipt_purchase_price)
		VALUES ($1,$2,$3,CURRENT_DATE,$4,$5) RETURNING receipt_id`,
		req.SupplierID, cl.Subject, req.ProductID, req.Quantity, req.PurchasePrice).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"receipt_id": id})
}

func (h *Handler) ListReceipts(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	supplierID := c.Query("supplier_id")

	args := []any{}
	sb := strings.Builder{}
	sb.WriteString(`SELECT r.receipt_id, r.supplier_id, s.organization_name, r.user_id, u.user_full_name,
		r.product_id, p.product_name, r.receipt_date, r.receipt_quantity, r.receipt_purchase_price
		FROM receipt r
		JOIN supplier s ON s.supplier_id=r.supplier_id
		JOIN "user"   u ON u.user_id=r.user_id
		JOIN product  p ON p.product_id=r.product_id WHERE 1=1 `)
	if from != "" {
		args = append(args, from)
		sb.WriteString(" AND r.receipt_date >= $" + strconv.Itoa(len(args)))
	}
	if to != "" {
		args = append(args, to)
		sb.WriteString(" AND r.receipt_date <= $" + strconv.Itoa(len(args)))
	}
	if supplierID != "" {
		args = append(args, supplierID)
		sb.WriteString(" AND r.supplier_id = $" + strconv.Itoa(len(args)))
	}
	sb.WriteString(" ORDER BY r.receipt_date DESC, r.receipt_id DESC")

	rows, err := h.pool.Query(c, sb.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Receipt{}
	for rows.Next() {
		var r model.Receipt
		if err := rows.Scan(&r.ID, &r.SupplierID, &r.SupplierName, &r.UserID, &r.UserName,
			&r.ProductID, &r.ProductName, &r.Date, &r.Quantity, &r.PurchasePrice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, out)
}
