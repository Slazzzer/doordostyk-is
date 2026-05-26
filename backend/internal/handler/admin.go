package handler

import (
	"net/http"
	"strconv"

	"github.com/doordostyk/api/internal/middleware"
	"github.com/doordostyk/api/internal/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ---- users ----

type createUserReq struct {
	FullName string `json:"full_name" binding:"required"`
	Login    string `json:"login"     binding:"required,min=3"`
	Role     string `json:"role"      binding:"required"`
	Password string `json:"password"  binding:"required,min=8"`
}

type customerReq struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) AdminListUsers(c *gin.Context) {
	rows, err := h.pool.Query(c, `SELECT user_id, user_full_name, user_login, user_role FROM "user" ORDER BY user_id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Login, &u.Role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, u)
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) AdminCreateUser(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != RoleAdmin && req.Role != RoleSeller && req.Role != RoleStoreman {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	if req.Role == RoleSeller {
		var sellers int
		if err := h.pool.QueryRow(c, `SELECT COUNT(*) FROM "user" WHERE user_role = $1`, RoleSeller).Scan(&sellers); err == nil && sellers > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "в системе уже есть продавец"})
			return
		}
	}
	if !isStrongPassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пароль: минимум 8 символов, латиница верх/низ, цифра и спецсимвол"})
		return
	}
	b, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var id int
	err := h.pool.QueryRow(c, `
		INSERT INTO "user"(user_full_name, user_login, user_role, user_password_hash)
		VALUES ($1,$2,$3,$4) RETURNING user_id`,
		req.FullName, req.Login, req.Role, string(b)).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user_id": id})
}

func (h *Handler) AdminDeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	cl := middleware.CurrentClaims(c)
	if cl != nil && cl.Subject == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нельзя удалить собственную учётную запись"})
		return
	}
	var role string
	err := h.pool.QueryRow(c, `SELECT user_role FROM "user" WHERE user_id=$1`, id).Scan(&role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}
	if role == RoleStoreman {
		c.JSON(http.StatusBadRequest, gin.H{"error": "учётную запись кладовщика удалять нельзя"})
		return
	}
	_, err = h.pool.Exec(c, `DELETE FROM "user" WHERE user_id=$1`, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- customers (admin only) ----

func (h *Handler) AdminListCustomers(c *gin.Context) {
	rows, err := h.pool.Query(c, `
		SELECT customer_id, customer_full_name, customer_email, customer_phone_number
		FROM customer ORDER BY customer_id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Customer{}
	for rows.Next() {
		var x model.Customer
		if err := rows.Scan(&x.ID, &x.FullName, &x.Email, &x.Phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, x)
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) AdminCreateCustomer(c *gin.Context) {
	var req customerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !phoneRE.MatchString(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "телефон в формате +7XXXXXXXXXX (10 цифр)"})
		return
	}
	if !isStrongPassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пароль: минимум 8 символов, латиница верх/низ, цифра и спецсимвол"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var id int
	err := h.pool.QueryRow(c, `
		INSERT INTO customer(customer_full_name, customer_email, customer_phone_number, customer_password_hash)
		VALUES ($1, LOWER($2), $3, $4) RETURNING customer_id`,
		req.FullName, req.Email, req.Phone, string(hash)).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"customer_id": id})
}

func (h *Handler) AdminUpdateCustomer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req customerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !phoneRE.MatchString(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "телефон в формате +7XXXXXXXXXX (10 цифр)"})
		return
	}
	if !isStrongPassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пароль: минимум 8 символов, латиница верх/низ, цифра и спецсимвол"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	_, err := h.pool.Exec(c, `
		UPDATE customer
		SET customer_full_name=$1, customer_email=LOWER($2), customer_phone_number=$3, customer_password_hash=$4
		WHERE customer_id=$5`,
		req.FullName, req.Email, req.Phone, string(hash), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) AdminDeleteCustomer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := h.pool.Exec(c, `DELETE FROM customer WHERE customer_id=$1`, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- categories ----

type categoryReq struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req categoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id int
	err := h.pool.QueryRow(c,
		`INSERT INTO category(category_name, category_description) VALUES ($1,$2) RETURNING category_id`,
		req.Name, req.Description).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category_id": id})
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req categoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.pool.Exec(c,
		`UPDATE category SET category_name=$1, category_description=$2 WHERE category_id=$3`,
		req.Name, req.Description, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := h.pool.Exec(c, `DELETE FROM category WHERE category_id=$1`, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- products ----

type productReq struct {
	CategoryID    int      `json:"category_id" binding:"required"`
	Name          string   `json:"name"        binding:"required"`
	Description   *string  `json:"description"`
	Dimensions    *string  `json:"dimensions"`
	PurchasePrice *float64 `json:"purchase_price"`
	RetailPrice   *float64 `json:"retail_price"`
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id int
	err := h.pool.QueryRow(c, `
		INSERT INTO product(category_id, product_name, product_description, product_dimensions,
		                    product_purchase_price, product_retail_price)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING product_id`,
		req.CategoryID, req.Name, req.Description, req.Dimensions,
		req.PurchasePrice, req.RetailPrice).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"product_id": id})
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.pool.Exec(c, `
		UPDATE product SET category_id=$1, product_name=$2, product_description=$3,
		    product_dimensions=$4, product_purchase_price=$5, product_retail_price=$6
		WHERE product_id=$7`,
		req.CategoryID, req.Name, req.Description, req.Dimensions,
		req.PurchasePrice, req.RetailPrice, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := h.pool.Exec(c, `DELETE FROM product WHERE product_id=$1`, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- suppliers ----

type supplierReq struct {
	Name    string  `json:"name"    binding:"required"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
}

func (h *Handler) ListSuppliers(c *gin.Context) {
	rows, err := h.pool.Query(c,
		`SELECT s.supplier_id, s.organization_name, s.supplier_address, s.supplier_phone_number,
		    NOT EXISTS (SELECT 1 FROM receipt r WHERE r.supplier_id = s.supplier_id) AS can_delete
		 FROM supplier s ORDER BY s.organization_name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []model.Supplier{}
	for rows.Next() {
		var s model.Supplier
		if err := rows.Scan(&s.ID, &s.Name, &s.Address, &s.Phone, &s.CanDelete); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, s)
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateSupplier(c *gin.Context) {
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id int
	err := h.pool.QueryRow(c, `
		INSERT INTO supplier(organization_name, supplier_address, supplier_phone_number)
		VALUES ($1,$2,$3) RETURNING supplier_id`,
		req.Name, req.Address, req.Phone).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"supplier_id": id})
}

func (h *Handler) UpdateSupplier(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.pool.Exec(c, `
		UPDATE supplier SET organization_name=$1, supplier_address=$2, supplier_phone_number=$3
		WHERE supplier_id=$4`,
		req.Name, req.Address, req.Phone, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteSupplier(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := h.pool.Exec(c, `DELETE FROM supplier WHERE supplier_id=$1`, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": mapDBError(err)})
		return
	}
	c.Status(http.StatusNoContent)
}
