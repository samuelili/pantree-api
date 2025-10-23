package main

import (
	"log"
	"pantree/api/db"

	"github.com/gin-gonic/gin"
)

func _handleAll(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	log.Printf("Getting all ingredients %s\n", userUuid)
	ingredients, err := queries.GetIngredients(c, getPgtypeUuid(userUuid))

	if err != nil {
		log.Println("Could not get ingredients: ", err)
		c.JSON(500, gin.H{
			"message": "Could not get ingredients.",
		})
		return
	}

	if ingredients == nil {
		ingredients = []db.Ingredient{}
	}

	c.JSON(200, ingredients)
}

type AddIngredientRequest struct {
	Name           string      `json:"name" binding:"required"`
	Unit           db.UnitType `json:"unit" binding:"required"`
	StorageLoc     db.LocType  `json:"storageLoc" binding:"required"`
	IngredientType db.GrocType `json:"ingredientType" binding:"required"`
	ImagePath      string      `json:"imagePath"`
}

func _handleAddIngredient(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	log.Printf("Getting  %s\n", userUuid)
	ingredients, err := queries.GetIngredients(c, getPgtypeUuid(userUuid))

	var request AddIngredientRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("Failed to bind JSON:", err)
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}

	if err != nil {
		log.Println("Could not get ingredients: ", err)
		c.JSON(500, gin.H{
			"message": "Could not get ingredients.",
		})
		return
	}

	if ingredients == nil {
		ingredients = []db.Ingredient{}
	}

	c.JSON(200, ingredients)
}

func registerIngredientRoutes(router *gin.RouterGroup) {
	router.GET("all", _handleAll)
	router.POST("add", _handleAddIngredient)
}
