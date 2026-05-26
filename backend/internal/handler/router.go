package handler

import (
	"time"

	"github.com/doordostyk/api/internal/config"
	"github.com/doordostyk/api/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	RoleAdmin    = "administrator"
	RoleSeller   = "seller"
	RoleStoreman = "storekeeper"
)

type Handler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

func NewRouter(pool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	h := &Handler{pool: pool, cfg: cfg}
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login/user", h.LoginUser)
			auth.POST("/login/customer", h.LoginCustomer)
			auth.POST("/register", h.RegisterCustomer)
			auth.GET("/me", middleware.AuthRequired(cfg.JWTSecret), h.Me)
		}

		cat := api.Group("/catalog")
		{
			cat.GET("/categories", h.ListCategories)
			cat.GET("/products", h.ListProducts)
			cat.GET("/products/:id", h.GetProduct)
		}

		secured := api.Group("/")
		secured.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			cust := secured.Group("/customer")
			cust.Use(middleware.RequireCustomer())
			{
				cust.POST("/orders", h.CreateOrderByCustomer)
				cust.GET("/orders/my", h.MyOrders)
				cust.PATCH("/orders/:id", h.UpdateCustomerOrder)
				cust.DELETE("/orders/:id", h.DeleteCustomerOrder)
			}

			seller := secured.Group("/seller")
			seller.Use(middleware.RequireRole(RoleSeller, RoleAdmin))
			{
				seller.GET("/orders", h.ListOrders)
				seller.POST("/orders/:id/execute", h.ExecuteOrder)
				seller.POST("/orders/:id/reject", h.RejectOrder)
				seller.POST("/sales", h.CreateSale)
				seller.GET("/sales", h.ListSales)
				seller.GET("/reports/sales", h.ReportSales)
			}

			storeman := secured.Group("/storeman")
			storeman.Use(middleware.RequireRole(RoleStoreman, RoleAdmin))
			{
				storeman.POST("/receipts", h.CreateReceipt)
				storeman.GET("/receipts", h.ListReceipts)
				storeman.GET("/stock", h.Stock)
				storeman.GET("/reports/receipts", h.ReportReceipts)
			}

			admin := secured.Group("/admin")
			admin.Use(middleware.RequireRole(RoleAdmin))
			{
				admin.GET("/users", h.AdminListUsers)
				admin.POST("/users", h.AdminCreateUser)
				admin.DELETE("/users/:id", h.AdminDeleteUser)

				admin.GET("/customers", h.AdminListCustomers)
				admin.POST("/customers", h.AdminCreateCustomer)
				admin.PATCH("/customers/:id", h.AdminUpdateCustomer)
				admin.DELETE("/customers/:id", h.AdminDeleteCustomer)

				admin.POST("/categories", h.CreateCategory)
				admin.PATCH("/categories/:id", h.UpdateCategory)
				admin.DELETE("/categories/:id", h.DeleteCategory)

				admin.POST("/products", h.CreateProduct)
				admin.PATCH("/products/:id", h.UpdateProduct)
				admin.DELETE("/products/:id", h.DeleteProduct)

				admin.GET("/suppliers", h.ListSuppliers)
				admin.POST("/suppliers", h.CreateSupplier)
				admin.PATCH("/suppliers/:id", h.UpdateSupplier)
				admin.DELETE("/suppliers/:id", h.DeleteSupplier)

				admin.GET("/dashboard", h.Dashboard)
			}

			// Suppliers list also for storeman (need it for receipts form)
			secured.GET("/suppliers", middleware.RequireRole(RoleStoreman, RoleSeller, RoleAdmin), h.ListSuppliers)
			// Stock balance also visible to seller / admin
			secured.GET("/seller/stock", middleware.RequireRole(RoleSeller, RoleAdmin), h.Stock)
		}
	}

	return r
}
