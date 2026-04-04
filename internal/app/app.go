// package app

// import (
// 	"database/sql"

// 	"github.com/huguescodeur/zoro_rest_api_go/internal/services"
// 	"github.com/huguescodeur/zoro_rest_api_go/internal/store"
// 	"github.com/huguescodeur/zoro_rest_api_go/internal/transport"
// )

// type App struct {
// 	ProductHandler *transport.ProductHandler
// 	ShopHandler    *transport.ShopHandler
// 	AuthHandler    *transport.AuthHandler
// 	UserHandler    *transport.UserHandler
// }

// func Init(db *sql.DB) *App {
// 	// ? Products
// 	productStore := store.NewProductSotre(db)
// 	productService := services.NewProductService(productStore)
// 	productHandler := transport.NewProductService(productService)

// 	// ? Shop
// 	shopStore := store.NewShopStore(db)
// 	shopService := services.NewShopService(shopStore)
// 	shopHandler := transport.NewShopService(shopService)

// 	// ? Auth
// 	authStore := store.NewAuthStore(db)
// 	authService := services.NewAuthService(authStore, userStore)
// 	authHandler := transport.NewAuthHandler(authService)
// 	// ? User
// 	userStore := store.NewUserStore(db)
// 	userService := services.NewUserService(userStore, authStore)
// 	userHandler := transport.NewUserService(userService)

// 	return &App{
// 		ProductHandler: productHandler,
// 		ShopHandler:    shopHandler,
// 		AuthHandler:    authHandler,
// 		UserHandler:    userHandler,
// 	}
// }

package app

import (
	"database/sql"

	"github.com/huguescodeur/zoro_rest_api_go/internal/services"
	"github.com/huguescodeur/zoro_rest_api_go/internal/store"
	"github.com/huguescodeur/zoro_rest_api_go/internal/transport"
)

type App struct {
	ProductHandler *transport.ProductHandler
	ShopHandler    *transport.ShopHandler
	AuthHandler    *transport.AuthHandler
	UserHandler    *transport.UserHandler
	StockHandler   *transport.StockHandler
	OrderHandler   *transport.OrderHandler
}

func Init(db *sql.DB) *App {
	authStore := store.NewAuthStore(db)
	userStore := store.NewUserStore(db)
	productStore := store.NewProductSotre(db)
	shopStore := store.NewShopStore(db)
	stockStore := store.NewStockStore(db)
	orderStore := store.NewOrderStore(db)

	authService := services.NewAuthService(authStore, userStore)
	userService := services.NewUserService(userStore, authStore)
	productService := services.NewProductService(productStore)
	shopService := services.NewShopService(shopStore)
	stockService := services.NewStockService(stockStore)
	orderService := services.NewOrderService(orderStore, productStore)

	authHandler := transport.NewAuthHandler(authService)
	userHandler := transport.NewUserHandler(userService)
	productHandler := transport.NewProductHandler(productService)
	shopHandler := transport.NewShopHandler(shopService)
	stockHandler := transport.NewStockHandler(stockService)
	orderHandler := transport.NewOrderHandler(orderService)

	return &App{
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		ProductHandler: productHandler,
		ShopHandler:    shopHandler,
		StockHandler:   stockHandler,
		OrderHandler:   orderHandler,
	}
}
