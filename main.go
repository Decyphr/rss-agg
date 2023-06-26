package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Decyphr/rss-agg/internal/database"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	fmt.Println("Hello, World!")

	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == ""{
		log.Fatal("PORT is not found in the environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not found in the environment variables")
	}

	conn, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		DB: database.New(conn),
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: 	[]string{"https://*", "http://*"},
		AllowedMethods: 	[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: 	[]string{"*"},
		ExposedHeaders: 	[]string{"Link"},
		AllowCredentials: false,
		MaxAge: 					300,
	}))

	// /v1
	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerErr)

	// Users
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	v1Router.Post("/users", apiCfg.handlerCreateUser)

	// Feeds
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr: ":" + portString,
	}

	log.Printf("Server starting on port %v", portString)

	srv.ListenAndServe()
	
	fmt.Println("Port:", portString)
}