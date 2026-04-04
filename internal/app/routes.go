package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/huguescodeur/zoro_rest_api_go/docs"
	"github.com/huguescodeur/zoro_rest_api_go/internal/middlewares"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *App) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173", "https://ee5b-74-244-119-50.ngrok-free.app"},

		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},

		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "ngrok-skip-browser-warning"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api", func(r chi.Router) {

		r.Route("/v1", func(r chi.Router) {
			r.Mount("/auth", a.AuthHandler.AuthRoutes())

			r.Group(func(r chi.Router) {
				r.Use(middlewares.AuthMiddleware)

				r.Mount("/users", a.UserHandler.UserRoutes())
				r.Mount("/products", a.ProductHandler.ProductRoutes())
				r.Mount("/shops", a.ShopHandler.ShopRoutes())
				r.Mount("/stocks", a.StockHandler.StockRoutes())
				r.Mount("/orders", a.OrderHandler.OrderRoutes())
			})

		})

	})

	return r
}
