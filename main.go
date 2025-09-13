package main

import (
	"context"
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

// func run() error {
// 	ctx := context.Background()

// 	conn, err := pgx.Connect(ctx, "user=samuel dbname=samuel sslmode=verify-full")
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close(ctx)

// 	queries := db.New(conn)

// 	// list all authors
// 	authors, err := queries.ListAuthors(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	log.Println(authors)

// 	// create an author
// 	insertedAuthor, err := queries.CreateAuthor(ctx, db.CreateAuthorParams{
// 		Name: "Brian Kernighan",
// 		Bio:  pgtype.Text{String: "Co-author of The C Programming Language and The Go Programming Language", Valid: true},
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	log.Println(insertedAuthor)

// 	// get the author we just inserted
// 	fetchedAuthor, err := queries.GetAuthor(ctx, insertedAuthor.ID)
// 	if err != nil {
// 		return err
// 	}

// 	// prints true
// 	log.Println(reflect.DeepEqual(insertedAuthor, fetchedAuthor))
// 	return nil
// }

// postAlbums adds an album from JSON received in the request body.
// func postAlbums(c *gin.Context) {
//     // var newAlbum album

//     // // Call BindJSON to bind the received JSON to
//     // // newAlbum.
//     // if err := c.BindJSON(&newAlbum); err != nil {
//     //     return
//     // }

//     // // Add the new album to the slice.
//     // albums = append(albums, newAlbum)
//     // c.IndentedJSON(http.StatusCreated, newAlbum)
// }

func sendError(c *gin.Context, errorCode int, err error, message string) {
	log.Println(message, ": ", err)
	c.IndentedJSON(errorCode, gin.H{"message": message, "error": err.Error()})
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
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
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Recipe updated successfully"})
}

func main() {
	// db startup
	ctx = context.Background()

	_conn, err := pgx.Connect(ctx, "user=samuel dbname=samuel sslmode=verify-full")
	conn = _conn
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	queries = db.New(conn)

	router := gin.Default()

	registerAuth(router)

	router.GET("/ping", ping)

	router.GET("/recipes", getRecipes)
	router.POST("/recipes/create", createRecipe)
	router.PUT("/recipes/update", updateRecipe)

	router.Run("localhost:8080")
}
