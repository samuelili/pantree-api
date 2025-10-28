package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"net/http"

	"pantree/api/db"

	"github.com/jackc/pgx/v5"

	"github.com/gin-gonic/gin"
)

var ctx context.Context
var conn *pgx.Conn
var queries *db.Queries

func sendError(c *gin.Context, errorCode int, err error, message string) {
	log.Println(message, ": ", err)
	c.IndentedJSON(errorCode, gin.H{"message": message, "error": err.Error()})
}

func bing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "bong",
	})
}

func main() {
	// read config
	readConfig()

	// db startup
	ctx = context.Background()

	err := loadAws(&ctx)
	if err != nil {
		log.Fatal("Error loading AWS", err)
		os.Exit(1)
	}

	// disable ssl_mode = verify_full for testing
	_conn, err := pgx.Connect(ctx, fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", cfg.Database.User, cfg.Database.Password, cfg.Database.DBName))
	conn = _conn
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	queries = db.New(conn)

	router := gin.Default()

	middleware := registerAuth(router)

	router.GET("/bing", bing)

	api := router.Group("/api", middleware.MiddlewareFunc())

	users := api.Group("/users")
	registerUserRoutes(users)

	ingredients := api.Group("/ingredients")
	registerIngredientsRoutes(ingredients)

	pantry := api.Group("/pantry")
	registerPantryRoutes(pantry)

	recipes := api.Group("/recipes")
	registerRecipeRoutes(recipes)

	router.Run(fmt.Sprintf("%s:%s", cfg.Server.Broadcast, cfg.Server.Port))
}
