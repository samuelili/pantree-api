package main

import (
	"log"
	"pantree/api/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// /getListings
func _handleGetListings(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	log.Printf("Getting listings %s\n", userUuid)
	listings, err := queries.GetItemListings(c)

	c.JSON(200, listings)
	return
}

// /addListing

type AddItemListing struct {
	Name      string      `json:"name" binding:"required"`
	UnitType  db.UnitType `json:"unitType" binding:"required"`
	CreatorId string      `json:"creatorId" binding:"required"`
}

func _handleAddListing(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	var request AddItemListing
	if err := c.BindJSON(&request); err != nil {
		log.Println("Invalid request body: \n", err)
		c.JSON(400, gin.H{
			"message": "Invalid request body.",
		})
		return
	}

	creatorUuid, err := uuid.Parse(request.CreatorId)
	if err != nil {
		log.Println("Invalid creator UUID: \n", err)
		c.JSON(400, gin.H{
			"message": "Invalid creator UUID.",
		})
		return
	}

	log.Printf("Adding listing for user %s\n", userUuid)

	newListing, err := queries.AddItemListing(c, db.AddItemListingParams{
		Name:      request.Name,
		UnitType:  request.UnitType,
		CreatorID: getPgtypeUuid(creatorUuid),
	})

	if err != nil {
		log.Println("Could not add listing: ", err)
		c.JSON(500, gin.H{
			"message": "Could not add listing.",
		})
		return
	}

	log.Printf("Successfully added listing: %v\n", newListing)
	c.JSON(200, newListing)
}

// /getEntries
func _handleGetEntries(c *gin.Context) {

}

// /addEntry
func _handleAddEntry(c *gin.Context) {

}

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
		log.Println("Invalid request body: \n", err)
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
	router.GET("/getListings", _handleGetListings)
	router.POST("/addListing", _handleAddListing)
	router.GET("/getEntries", _handleGetEntries)
	router.POST("/addEntry", _handleAddEntry)
}
