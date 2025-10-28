package main

import (
	"log"

	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

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
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Steps       []string `json:"steps" binding:"required"`
	Allergens   []string `json:"allergens" binding:"required"`
	CookingTime float64  `json:"cookingtime" binding:"required"`
	ServingSize float64  `json:"servingsize" binding:"required"`
	ImagePath   string   `json:"imagepath"`
}

type RecipeIngredientRequest struct {
	Ingredient        string  `json:"ingredient" binding:"required"`
	Quantity          float64 `json:"quantity" binding:"required"`
	AuthorUnitType    string  `json:"authorunittype" binding:"required"`
	AuthorMeasureType string  `json:"authormeasuretype" binding:"required"`
}

func createRecipe(c *gin.Context) {
	var request RecipeRequest

	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	var requestIngredients []RecipeIngredientRequest

	if err := c.ShouldBindBodyWith(&requestIngredients, binding.JSON); err != nil {
		sendError(c, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	userUuid, _ := getUserId(c)

	newRecipe := db.CreateRecipeParams{
		CreatorID: getPgtypeUuid(userUuid),
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

	newRecipe.CookingTime = pgtype.Numeric{}
	_ = newRecipe.CookingTime.Scan(request.CookingTime)
	newRecipe.CookingTime.Valid = true

	newRecipe.ServingSize = pgtype.Numeric{}
	_ = newRecipe.ServingSize.Scan(request.ServingSize)
	newRecipe.ServingSize.Valid = true

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

	for idx, ingredient := range requestIngredients {

		recipeIngredient := db.CreateRecipeIngredientParams{
			RecipeID:          createdRecipe.ID,
			AuthorUnitType:    db.UnitType(requestIngredients[idx].AuthorUnitType),
			AuthorMeasureType: db.MeasureType(requestIngredients[idx].AuthorMeasureType),
		}

		recipeIngredient.IngredientID = pgtype.UUID{}
		_ = recipeIngredient.IngredientID.Scan(ingredient)
		recipeIngredient.IngredientID.Valid = true

		recipeIngredient.Quantity = pgtype.Numeric{}
		_ = recipeIngredient.Quantity.Scan(requestIngredients[idx].Quantity)
		recipeIngredient.Quantity.Valid = true

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

func registerRecipeRoutes(router *gin.RouterGroup) {
	router.GET("/get", getRecipes)
	router.POST("/create", createRecipe)
	router.POST("/update", updateRecipe)
}
