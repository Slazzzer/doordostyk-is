package handler

import (
	"net/http"
	"strings"

	"github.com/doordostyk/api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type userLoginReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type customerLoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type customerRegisterReq struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) LoginUser(c *gin.Context) {
	var req userLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var (
		id       int
		hash     string
		role     string
		fullName string
	)
	err := h.pool.QueryRow(c, `
		SELECT user_id, user_password_hash, user_role, user_full_name
		FROM "user" WHERE user_login=$1`, req.Login).Scan(&id, &hash, &role, &fullName)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}
	tok, err := middleware.GenerateToken(h.cfg.JWTSecret, middleware.SubjectTypeUser, id, role, fullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":     tok,
		"user_id":   id,
		"role":      role,
		"full_name": fullName,
		"type":      "user",
	})
}

func (h *Handler) LoginCustomer(c *gin.Context) {
	var req customerLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	var (
		id       int
		hash     string
		fullName string
	)
	err := h.pool.QueryRow(c, `
		SELECT customer_id, COALESCE(customer_password_hash,''), customer_full_name
		FROM customer WHERE LOWER(customer_email)=$1`, email).Scan(&id, &hash, &fullName)
	if err == pgx.ErrNoRows || hash == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}
	tok, err := middleware.GenerateToken(h.cfg.JWTSecret, middleware.SubjectTypeCustomer, id, "", fullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":     tok,
		"user_id":   id,
		"full_name": fullName,
		"type":      "customer",
	})
}

func (h *Handler) RegisterCustomer(c *gin.Context) {
	var req customerRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)
	if !phoneRE.MatchString(phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "телефон в формате +7XXXXXXXXXX (10 цифр)"})
		return
	}
	if !isStrongPassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пароль: минимум 8 символов, латиница верх/низ, цифра и спецсимвол"})
		return
	}
	b, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var id int
	err = h.pool.QueryRow(c, `
		INSERT INTO customer(customer_full_name, customer_email, customer_phone_number, customer_password_hash)
		VALUES ($1, $2, NULLIF($3,''), $4)
		RETURNING customer_id`,
		req.FullName, email, phone, string(b)).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tok, _ := middleware.GenerateToken(h.cfg.JWTSecret, middleware.SubjectTypeCustomer, id, "", req.FullName)
	c.JSON(http.StatusCreated, gin.H{
		"token":     tok,
		"user_id":   id,
		"full_name": req.FullName,
		"type":      "customer",
	})
}

func (h *Handler) Me(c *gin.Context) {
	cl := middleware.CurrentClaims(c)
	if cl == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":   cl.Subject,
		"type":      cl.SubjectType,
		"role":      cl.Role,
		"full_name": cl.Name,
	})
}
