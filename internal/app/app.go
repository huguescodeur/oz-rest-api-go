package app

import (
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/mailer"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
	"github.com/huguescodeur/oz-rest-api-go/internal/transport"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	ProductHandler   *transport.ProductHandler
	ShopHandler      *transport.ShopHandler
	AuthHandler      *transport.AuthHandler
	UserHandler      *transport.UserHandler
	StockHandler     *transport.StockHandler
	OrderHandler     *transport.OrderHandler
	CategoryHandler  *transport.CategoryHandler
	DashboardHandler *transport.DashboardHandler
	SuperHandler     *transport.SuperHandler
}

func Init(db *pgxpool.Pool) *App {
	authStore := store.NewAuthStore(db)
	userStore := store.NewUserStore(db)
	productStore := store.NewProductStore(db)
	shopStore := store.NewShopStore(db)
	stockStore := store.NewStockStore(db)
	orderStore := store.NewOrderStore(db)
	categoryStore := store.NewCategoryStore(db)
	dashboardStore := store.NewDashboardStore(db)
	superStore := store.NewSuperStore(db)

	m := mailer.New()

	authService := services.NewAuthService(authStore, userStore)
	userService := services.NewUserService(userStore, authStore)
	productService := services.NewProductService(productStore, categoryStore, shopStore, stockStore)
	shopService := services.NewShopService(shopStore)
	stockService := services.NewStockService(stockStore, userStore, m)
	orderService := services.NewOrderService(orderStore, productStore, stockService)
	categoryService := services.NewCategoryService(categoryStore)
	dashboardService := services.NewDashboardService(dashboardStore)
	superService := services.NewSuperService(superStore)

	authHandler := transport.NewAuthHandler(authService)
	userHandler := transport.NewUserHandler(userService)
	productHandler := transport.NewProductHandler(productService)
	shopHandler := transport.NewShopHandler(shopService)
	stockHandler := transport.NewStockHandler(stockService)
	orderHandler := transport.NewOrderHandler(orderService)
	categoryHandler := transport.NewCategoryHandler(categoryService)
	dashboardHandler := transport.NewDashboardHandler(dashboardService)
	superHandler := transport.NewSuperHandler(superService)

	return &App{
		AuthHandler:      authHandler,
		UserHandler:      userHandler,
		ProductHandler:   productHandler,
		ShopHandler:      shopHandler,
		StockHandler:     stockHandler,
		OrderHandler:     orderHandler,
		CategoryHandler:  categoryHandler,
		DashboardHandler: dashboardHandler,
		SuperHandler:     superHandler,
	}
}
