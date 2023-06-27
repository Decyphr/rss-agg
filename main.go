package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

	db := database.New(conn)
	apiCfg := apiConfig{
		DB: db,
	}

	// use go routine to start scraping
	// we do this because startScraping() is a long-running function
	go startScraping(db, 10, time.Minute)

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

	// Feed Follows
	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))

	// Posts
	v1Router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUser))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr: ":" + portString,
	}

	log.Printf("Server starting on port %v", portString)

	srv.ListenAndServe()
	
	fmt.Println("Port:", portString)
}