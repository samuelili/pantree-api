package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"pantree/api/db"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

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

type RecipeRequest struct {
	Name        string             `json:"name" binding:"required"`
	Description string             `json:"description"`
	Steps       []string           `json:"steps" binding:"required"`
	Allergens   []string           `json:"allergens" binding:"required"`
	CookingTime string             `json:"cookingtime" binding:"required"`
	ServingSize string             `json:"servingsize" binding:"required"`
	ImagePath   string             `json:"imagepath"`
	Ingredients []RecipeIngredient `json:"ingredients" binding:"required"`
}

type RecipeIngredient struct {
	Ingredient        string `json:"ingredient" binding:"required"`
	Quantity          string `json:"quantity" binding:"required"`
	AuthorUnitType    string `json:"authorunittype" binding:"required"`
	AuthorMeasureType string `json:"authormeasuretype" binding:"required"`
}

func createRecipe(c *gin.Context) {
	var request RecipeRequest

	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
		return
	}

	newRecipe := db.CreateRecipeParams{
		CreatorID: &userUuid,
		Name:      request.Name,
		Steps:     request.Steps,
		Allergens: request.Allergens,
		ImagePath: pgtype.Text{String: request.ImagePath, Valid: true},
	}

	if request.Description != "" {
		newRecipe.Description = pgtype.Text{
			String: request.Description,
			Valid:  true,
		}
	}

	newRecipe.CookingTime, err = decimal.NewFromString(request.CookingTime)
	if err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid cooking time")
		return
	}

	newRecipe.ServingSize, err = decimal.NewFromString(request.ServingSize)
	if err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid serving size")
		return
	}

	ctx := c.Request.Context()

	tx, err := conn.Begin(ctx)

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Failed to start transaction")
		return
	}

	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)

	createdRecipe, err := qtx.CreateRecipe(ctx, newRecipe)

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Could not create recipe")
		return
	}

	c.IndentedJSON(http.StatusCreated, createdRecipe)

	requestIngredients := request.Ingredients

	for _, ingredient := range requestIngredients {

		recipeIngredient := db.CreateRecipeIngredientParams{
			RecipeID:          createdRecipe.ID,
			AuthorUnitType:    db.UnitType(ingredient.AuthorUnitType),
			AuthorMeasureType: db.MeasureType(ingredient.AuthorMeasureType),
		}

		ingredientId, err := uuid.Parse(ingredient.Ingredient)
		if err != nil {
			sendError(c, http.StatusBadRequest, err, "Invalid ingredient UUID")
			return
		}
		recipeIngredient.IngredientID = ingredientId

		recipeIngredient.Quantity, err = decimal.NewFromString(ingredient.Quantity)
		if err != nil {
			sendError(c, http.StatusBadRequest, err, "Invalid Quantity")
			return
		}

		_, err = qtx.CreateRecipeIngredient(ctx, recipeIngredient)
		if err != nil {
			sendError(c, http.StatusInternalServerError, err, "Could not insert ingredient")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		sendError(c, http.StatusInternalServerError, err, "Transaction failed")
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Recipe created successfully"})
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

func favoriteRecipe(c *gin.Context) {
	var recipeId uuid.UUID

	if err := c.BindJSON(&recipeId); err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
		return
	}

	err = queries.AddFavorite(c, db.AddFavoriteParams{
		UserID:   userUuid,
		RecipeID: recipeId,
	})

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Could not favorite recipe")
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Recipe favorited sucessfully"})

}

func unfavoriteRecipe(c *gin.Context) {
	var recipeId uuid.UUID

	if err := c.BindJSON(&recipeId); err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
		return
	}

	err = queries.RemoveFavorite(c, db.RemoveFavoriteParams{
		UserID:   userUuid,
		RecipeID: recipeId,
	})

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Could not unfavorite recipe")
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Recipe unfavorited sucessfully"})

}

func getFavorites(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
		return
	}

	favorites, err := queries.GetFavorites(c, userUuid)

	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Could not fetch favorites")
		return
	}

	if len(favorites) == 0 {
		favorites = []uuid.UUID{}
	}

	c.IndentedJSON(http.StatusOK, favorites)
}

func registerRecipeRoutes(router *gin.RouterGroup) {
	router.GET("/get", getRecipes)
	router.POST("/create", createRecipe)
	router.POST("/update", updateRecipe)
	router.POST("/favorite", favoriteRecipe)
	router.POST("/unfavorite", unfavoriteRecipe)
	router.GET("/getfavorites", getFavorites)
}
