package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	SubjectTypeUser     = "user"     // employee
	SubjectTypeCustomer = "customer" // client
)

type Claims struct {
	Subject     int    `json:"sub_id"`
	SubjectType string `json:"sub_type"`
	Role        string `json:"role"` // for employees
	Name        string `json:"name"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, subjectType string, id int, role, name string) (string, error) {
	claims := Claims{
		Subject:     id,
		SubjectType: subjectType,
		Role:        role,
		Name:        name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "doordostyk",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		claims := &Claims{}
		tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(secret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("claims", claims)
		c.Set("user_id", claims.Subject)
		c.Set("user_type", claims.SubjectType)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireRole допускает только сотрудников указанных ролей.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("claims")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no claims"})
			return
		}
		cl := v.(*Claims)
		if cl.SubjectType != SubjectTypeUser {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "employees only"})
			return
		}
		for _, r := range roles {
			if cl.Role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
	}
}

func RequireCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("claims")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no claims"})
			return
		}
		cl := v.(*Claims)
		if cl.SubjectType != SubjectTypeCustomer {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "customers only"})
			return
		}
		c.Next()
	}
}

func CurrentClaims(c *gin.Context) *Claims {
	v, ok := c.Get("claims")
	if !ok {
		return nil
	}
	return v.(*Claims)
}
