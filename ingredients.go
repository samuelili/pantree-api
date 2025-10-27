package main

import (
	"log"
	"pantree/api/db"

	"github.com/gin-gonic/gin"
)

/**
 * /getIngredients
 */
func _handleGetIngredients(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	log.Printf("Getting ingredients %s\n", userUuid)
	ingredients, err := queries.GetIngredients(c, getPgtypeUuid(userUuid))

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

	log.Printf("Creating new ingredient for user %s\n", userUuid)

	newIngredient, err := queries.CreateIngredient(c, db.CreateIngredientParams{
		UserID:         getPgtypeUuid(userUuid),
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

func registerIngredientsRoutes(router *gin.RouterGroup) {
	router.GET("/ingredients", _handleGetIngredients)
	router.POST("/newIngredient", _handleNewIngredient)
}
