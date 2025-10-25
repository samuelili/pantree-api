package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"

	"net/http"

	"pantree/api/db"

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

func getRecipes(c *gin.Context) {
	recipes, err := queries.ListRecipes(c)

	log.Default().Println("recipes: ", recipes)

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Could not fetch recipes")
		return
	}

	if len(recipes) == 0 {
		recipes = []db.Recipe{}
	}

	c.IndentedJSON(http.StatusOK, recipes)
}

func createRecipe(c *gin.Context) {
	var newRecipe db.CreateRecipeParams

	if err := c.BindJSON(&newRecipe); err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	createdRecipe, err := queries.CreateRecipe(ctx, newRecipe)

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Could not create recipe")
		return
	}

	c.IndentedJSON(http.StatusCreated, createdRecipe)

	// need to insert ingredients as well
	// initialize slice of new recipe ingredients
	// assume input is slice of structs w/ ingredient id, quantity
	// var recipeIngredients []models.RecipeIngredients

	// ctx := c.Request.Context()

	// tx, err := conn.Begin(ctx)
	// if err != nil {
	// 	sendError(c, http.StatusInternalServerError, err, "Failed to start transaction")
	// 	return
	// }

	// defer tx.Rollback(ctx)

	// qtx := queries.WithTx(tx)
	// for _, v := range recipeIngredients {

	// 	var ingredientUUID pgtype.UUID
	// 	err := ingredientUUID.Scan(v.IngredientID)

	// 	if err != nil {
	// 		sendError(c, http.StatusBadRequest, err, "Invalid ingredient id")
	// 		return
	// 	}

	// 	_, err = qtx.CreateRecipeIngredient(ctx, db.CreateRecipeIngredientParams{
	// 		RecipeID:     createdRecipe.ID,
	// 		IngredientID: ingredientUUID,
	// 		Quantity:     pgtype.Numeric{v.Quantity},
	// 	})
	// 	if err != nil {
	// 		sendError(c, http.StatusInternalServerError, err, "Could not insert ingredient")
	// 		return
	// 	}

	// }

	// if err := tx.Commit(ctx); err != nil {
	// 	sendError(c, http.StatusInternalServerError, err, "Transaction failed")
	// 	return
	// }

}

func updateRecipe(c *gin.Context) {
	var updateRecipe db.UpdateRecipeParams

	if err := c.BindJSON(&updateRecipe); err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	err := queries.UpdateRecipe(ctx, updateRecipe)

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Could not update recipe")
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Recipe updated successfully"})
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

	api.GET("/recipes", getRecipes)
	api.POST("/recipes/create", createRecipe)
	api.PUT("/recipes/update", updateRecipe)

	users := api.Group("/users")
	registerUserRoutes(users)

	ingredients := api.Group("/ingredients")
	registerIngredientRoutes(ingredients)

	router.Run(fmt.Sprintf("%s:%s", cfg.Server.Broadcast, cfg.Server.Port))
}
