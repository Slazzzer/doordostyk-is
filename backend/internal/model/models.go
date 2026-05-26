package model

import "time"

type Category struct {
	ID          int     `json:"category_id"`
	Name        string  `json:"category_name"`
	Description *string `json:"category_description,omitempty"`
	CanDelete   bool    `json:"can_delete"`
}

type Supplier struct {
	ID      int     `json:"supplier_id"`
	Name    string  `json:"organization_name"`
	Address *string `json:"supplier_address,omitempty"`
	Phone   *string `json:"supplier_phone_number,omitempty"`
	CanDelete bool  `json:"can_delete"`
}

type Customer struct {
	ID       int     `json:"customer_id"`
	FullName string  `json:"customer_full_name"`
	Email    *string `json:"customer_email,omitempty"`
	Phone    *string `json:"customer_phone_number,omitempty"`
}

type User struct {
	ID       int    `json:"user_id"`
	FullName string `json:"user_full_name"`
	Login    string `json:"user_login"`
	Role     string `json:"user_role"`
}

type Product struct {
	ID            int      `json:"product_id"`
	CategoryID    int      `json:"category_id"`
	CategoryName  string   `json:"category_name,omitempty"`
	Name          string   `json:"product_name"`
	Description   *string  `json:"product_description,omitempty"`
	Dimensions    *string  `json:"product_dimensions,omitempty"`
	PurchasePrice *float64 `json:"product_purchase_price,omitempty"`
	RetailPrice   *float64 `json:"product_retail_price,omitempty"`
	Balance       int      `json:"balance,omitempty"`
	CanDelete     bool     `json:"can_delete"`
}

type Receipt struct {
	ID            int       `json:"receipt_id"`
	SupplierID    int       `json:"supplier_id"`
	SupplierName  string    `json:"supplier_name,omitempty"`
	UserID        int       `json:"user_id"`
	UserName      string    `json:"user_name,omitempty"`
	ProductID     int       `json:"product_id"`
	ProductName   string    `json:"product_name,omitempty"`
	Date          time.Time `json:"receipt_date"`
	Quantity      int       `json:"receipt_quantity"`
	PurchasePrice float64   `json:"receipt_purchase_price"`
}

type Sale struct {
	ID          int       `json:"sale_id"`
	OrderID     *int      `json:"order_id,omitempty"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	UserID      int       `json:"user_id"`
	UserName    string    `json:"user_name,omitempty"`
	Date        time.Time `json:"sale_date"`
	Quantity    int       `json:"sale_quantity"`
	Price       float64   `json:"sale_price"`
	Amount      float64   `json:"sale_amount,omitempty"`
}

type Order struct {
	ID           int       `json:"order_id"`
	CustomerID   int       `json:"customer_id"`
	CustomerName string    `json:"customer_name,omitempty"`
	ProductID    int       `json:"product_id"`
	ProductName  string    `json:"product_name,omitempty"`
	SaleID       *int      `json:"sale_id,omitempty"`
	Date         time.Time `json:"order_date"`
	Quantity     int       `json:"order_quantity"`
	Status       string    `json:"order_status"`
}

type StockItem struct {
	ProductID    int      `json:"product_id"`
	ProductName  string   `json:"product_name"`
	Dimensions   *string  `json:"product_dimensions,omitempty"`
	RetailPrice  *float64 `json:"product_retail_price,omitempty"`
	CategoryID   int      `json:"category_id"`
	CategoryName string   `json:"category_name"`
	ReceivedQty  int      `json:"received_qty"`
	SoldQty      int      `json:"sold_qty"`
	Balance      int      `json:"balance"`
}

type SalesReportRow struct {
	SaleID       int       `json:"sale_id"`
	Date         time.Time `json:"sale_date"`
	CategoryID   int       `json:"category_id"`
	CategoryName string    `json:"category_name"`
	ProductID    int       `json:"product_id"`
	ProductName  string    `json:"product_name"`
	Dimensions   *string   `json:"product_dimensions,omitempty"`
	Quantity     int       `json:"sale_quantity"`
	Price        float64   `json:"sale_price"`
	Amount       float64   `json:"sale_amount"`
	SellerName   string    `json:"seller_name"`
}

type ReceiptsReportRow struct {
	ReceiptID    int       `json:"receipt_id"`
	Date         time.Time `json:"receipt_date"`
	SupplierID   int       `json:"supplier_id"`
	SupplierName string    `json:"supplier_name"`
	ProductID    int       `json:"product_id"`
	ProductName  string    `json:"product_name"`
	CategoryName string    `json:"category_name"`
	Quantity     int       `json:"receipt_quantity"`
	Price        float64   `json:"receipt_purchase_price"`
	Amount       float64   `json:"receipt_amount"`
}
