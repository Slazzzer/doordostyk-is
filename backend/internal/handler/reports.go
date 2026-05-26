package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/doordostyk/api/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ReportSales(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	categoryID := c.Query("category_id")

	args := []any{}
	sb := strings.Builder{}
	sb.WriteString(`SELECT sale_id, sale_date, category_id, category_name, product_id, product_name,
		product_dimensions, sale_quantity, sale_price, sale_amount, seller_name
		FROM v_sales_by_category WHERE 1=1 `)
	if from != "" {
		args = append(args, from)
		sb.WriteString(" AND sale_date >= $" + strconv.Itoa(len(args)))
	}
	if to != "" {
		args = append(args, to)
		sb.WriteString(" AND sale_date <= $" + strconv.Itoa(len(args)))
	}
	if categoryID != "" {
		args = append(args, categoryID)
		sb.WriteString(" AND category_id = $" + strconv.Itoa(len(args)))
	}
	sb.WriteString(" ORDER BY sale_date DESC, sale_id DESC")

	rows, err := h.pool.Query(c, sb.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.SalesReportRow{}
	totalQty := 0
	totalAmount := 0.0
	for rows.Next() {
		var r model.SalesReportRow
		if err := rows.Scan(&r.SaleID, &r.Date, &r.CategoryID, &r.CategoryName,
			&r.ProductID, &r.ProductName, &r.Dimensions,
			&r.Quantity, &r.Price, &r.Amount, &r.SellerName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		totalQty += r.Quantity
		totalAmount += r.Amount
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{
		"rows":         out,
		"total_qty":    totalQty,
		"total_amount": totalAmount,
	})
}

func (h *Handler) ReportReceipts(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	supplierID := c.Query("supplier_id")
	categoryID := c.Query("category_id")

	args := []any{}
	sb := strings.Builder{}
	sb.WriteString(`SELECT r.receipt_id, r.receipt_date, r.supplier_id, s.organization_name,
		r.product_id, p.product_name, c.category_name, r.receipt_quantity,
		r.receipt_purchase_price, (r.receipt_quantity*r.receipt_purchase_price)::DECIMAL(14,2)
		FROM receipt r
		JOIN supplier s ON s.supplier_id=r.supplier_id
		JOIN product  p ON p.product_id=r.product_id
		JOIN category c ON c.category_id=p.category_id WHERE 1=1 `)
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
	if categoryID != "" {
		args = append(args, categoryID)
		sb.WriteString(" AND p.category_id = $" + strconv.Itoa(len(args)))
	}
	sb.WriteString(" ORDER BY r.receipt_date DESC, r.receipt_id DESC")

	rows, err := h.pool.Query(c, sb.String(), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.ReceiptsReportRow{}
	totalQty := 0
	totalAmount := 0.0
	for rows.Next() {
		var r model.ReceiptsReportRow
		if err := rows.Scan(&r.ReceiptID, &r.Date, &r.SupplierID, &r.SupplierName,
			&r.ProductID, &r.ProductName, &r.CategoryName,
			&r.Quantity, &r.Price, &r.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		totalQty += r.Quantity
		totalAmount += r.Amount
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{
		"rows":         out,
		"total_qty":    totalQty,
		"total_amount": totalAmount,
	})
}

func (h *Handler) Dashboard(c *gin.Context) {
	type kv struct {
		Label string  `json:"label"`
		Value float64 `json:"value"`
	}
	var (
		todaySales  float64
		monthSales  float64
		newOrders   int
		lowStock    int
		productsCnt int
	)
	_ = h.pool.QueryRow(c,
		`SELECT COALESCE(SUM(sale_quantity*sale_price),0) FROM sale WHERE sale_date = CURRENT_DATE`).Scan(&todaySales)
	_ = h.pool.QueryRow(c,
		`SELECT COALESCE(SUM(sale_quantity*sale_price),0) FROM sale
		 WHERE sale_date >= DATE_TRUNC('month', CURRENT_DATE)`).Scan(&monthSales)
	_ = h.pool.QueryRow(c,
		`SELECT COUNT(*) FROM "order" WHERE order_status='новый'`).Scan(&newOrders)
	_ = h.pool.QueryRow(c,
		`SELECT COUNT(*) FROM v_stock_balance WHERE balance < 5`).Scan(&lowStock)
	_ = h.pool.QueryRow(c, `SELECT COUNT(*) FROM product`).Scan(&productsCnt)

	rows, _ := h.pool.Query(c, `
		SELECT category_name, COALESCE(SUM(sale_amount),0)::float8 AS amount
		FROM v_sales_by_category
		WHERE sale_date >= DATE_TRUNC('month', CURRENT_DATE)
		GROUP BY category_name ORDER BY amount DESC`)
	defer rows.Close()
	top := []kv{}
	for rows.Next() {
		var x kv
		if err := rows.Scan(&x.Label, &x.Value); err == nil {
			top = append(top, x)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sales_today":           todaySales,
		"sales_month":           monthSales,
		"new_orders":            newOrders,
		"low_stock_count":       lowStock,
		"products_count":        productsCnt,
		"top_categories_month":  top,
	})
}
