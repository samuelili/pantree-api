package main

import (
	"fmt"
	"log"
	"pantree/api/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

/**
 * /getIngredients
 */
func _handleGetIngredients(c *gin.Context) {
	log.Println("Getting ingredients")
	ingredients, err := queries.GetIngredients(c)

	if err != nil {
		log.Println("Could not get ingredients:", err)
		c.JSON(500, gin.H{
			"message": "Could not get ingredients.",
		})
	}

	if ingredients == nil {
		ingredients = []db.Ingredient{}
	}

	c.JSON(200, ingredients)
}

/**
 * /newIngredient
 */
type NewIngredientRequest struct {
	Name           string      `json:"name" binding:"required"`
	Unit           db.UnitType `json:"unit" binding:"required"`
	StorageLoc     db.LocType  `json:"storageLoc" binding:"required"`
	IngredientType db.GrocType `json:"ingredientType" binding:"required"`
	ImagePath      string      `json:"imagePath"`
}

func _handleNewIngredient(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	var request NewIngredientRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("Invalid request body: \n", err)
		c.JSON(400, gin.H{
			"message": "Invalid request body.",
		})
		return
	}

	log.Printf("Creating new ingredient for user %s", userUuid)

	newIngredient, err := queries.CreateIngredient(c, db.CreateIngredientParams{
		CreatorID:      &userUuid,
		Name:           request.Name,
		Unit:           request.Unit,
		StorageLoc:     request.StorageLoc,
		IngredientType: request.IngredientType,
		ImagePath:      getPgtypeText(request.ImagePath),
	})

	if err != nil {
		log.Println("Could not create ingredient: ", err)
		c.JSON(500, gin.H{
			"message": "Could not create ingredient.",
		})
		return
	}

	log.Printf("Successfully created ingredient: %v\n", newIngredient)
	c.JSON(200, newIngredient)
}

/**
 * /getIngredientsByIds
 */
type GetIngredientsByIdsRequest struct {
	Ids []string `json:"ids" binding:"required"`
}

func _handleGetIngredientsByIds(c *gin.Context) {
	var request GetIngredientsByIdsRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("Invalid request body: \n", err)
		sendError(c, 400, err, "Invalid request body.")
		return
	}

	log.Println("Getting ingredients by ids")
	uuids := make([]uuid.UUID, len(request.Ids))
	for i, id := range request.Ids {
		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Println("Invalid UUID: ", err)
			sendError(c, 400, err, fmt.Sprintf("Invalid UUID: %s. at %d", id, i))
			return
		}
		uuids[i] = uuid
	}

	ingredients, err := queries.GetIngredientsByIds(c, uuids)

	if err != nil {
		log.Println("Could not get ingredients by ids:", err)
		sendError(c, 500, err, "Could not get ingredients by ids.")
		return
	}

	if ingredients == nil {
		ingredients = []db.Ingredient{}
	}

	c.JSON(200, ingredients)
}

/**
 * /searchIngredients
 */
type SearchIngredientsRequest struct {
	Name string `json:"name" binding:"required"`
}

func _handleSearchIngredients(c *gin.Context) {
	var request SearchIngredientsRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("Invalid request body: \n", err)
		sendError(c, 400, err, "Invalid request body.")
		return
	}

	log.Println("Searching ingredients")
	ingredients, err := queries.SearchIngredients(c, pgtype.Text{String: request.Name, Valid: true})

	if err != nil {
		log.Println("Could not search ingredients:", err)
		sendError(c, 500, err, "Could not search ingredients.")
		return
	}

	if ingredients == nil {
		ingredients = []db.Ingredient{}
	}

	c.JSON(200, ingredients)
}

func registerIngredientsRoutes(router *gin.RouterGroup) {
	router.GET("/ingredients", _handleGetIngredients)
	router.POST("/ingredientsByIds", _handleGetIngredientsByIds)
	router.POST("/searchIngredients", _handleSearchIngredients)
	router.POST("/newIngredient", _handleNewIngredient)
}
