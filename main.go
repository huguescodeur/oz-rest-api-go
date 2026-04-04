package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/huguescodeur/oz-rest-api-go/internal/app"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/config"
)

// @title           o'z api
// @version         1.0
// @termsOfService  http://swagger.io/terms/
// @description     API de gestion de commerce multi-tenants et multi-boutiques/magasins construite en Go.

// @contact.name   Support API
// @contact.email  goliyaohugues@gmail.com

//@host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" suivi d'un espace et du token JWT.
func main() {
	config.LoadConfig()

	// ? Connection MySQL
	cfg := mysql.Config{
		User:                 os.Getenv("DB_USER"),
		Passwd:               os.Getenv("DB_PASS"),
		Net:                  "tcp",
		Addr:                 os.Getenv("DB_ADDR"),
		DBName:               os.Getenv("DB_NAME"),
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// ? Initialisons notre App
	myApp := app.Init(db)

	// ? Démarrer et écouter le serveur
	fmt.Println("Server start http://localhost:8080")

	http.ListenAndServe(":8080", myApp.Routes())

}
