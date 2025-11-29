package main

import (
	"log"
	"pantree/api/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

/**
 * /getIngredients
 */
func _handleGetPantry(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	log.Printf("Getting pantry for user %s\n", userUuid)
	pantry, err := queries.GetUserPantry(c, db.GetUserPantryParams{
		UserID: &userUuid,
	})

	if err != nil {
		log.Println("Could not get pantry:", err)
		c.JSON(500, gin.H{
			"message": "Could not get pantry.",
		})
	}

	if pantry == nil {
		pantry = []db.GetUserPantryRow{}
	}

	c.JSON(200, pantry)
}

/**
 * /createItem
 */
type CreateUserItemRequest struct {
	IngredientId   uuid.UUID `json:"ingredientId" binding:"required"`
	Quantity       int64     `json:"quantity" binding:"required"`
	Price          *float64  `json:"price"`
	ExpirationDate *float64  `json:"expirationDate"`
}

func _handleAddUserItem(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
		return
	}

	var request CreateUserItemRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("Invalid request body: \n", err)
		sendError(c, 400, err, "Invalid request body.")
		return
	}

	log.Printf("Creating new user item for user %s\n", userUuid)

	// create quantity
	quantity := decimal.NewFromInt(request.Quantity)

	var price decimal.NullDecimal
	if request.Price != nil {
		price.Decimal = decimal.NewFromFloat(*request.Price)
		price.Valid = true
	}

	item, err := queries.CreateUserItemEntry(c, db.CreateUserItemEntryParams{
		UserID:       &userUuid,
		IngredientID: &request.IngredientId,
		Quantity:     quantity,
		Price:        price,
	})

	if err != nil {
		log.Println("Could not create user item: ", err)
		sendError(c, 500, err, "Could not create user item.")
		return
	}

	log.Printf("Successfully created user item: %v\n", item.ID)
	c.JSON(200, item)
}

func registerPantryRoutes(router *gin.RouterGroup) {
	router.GET("/getPantry", _handleGetPantry)
	router.POST("/createItem", _handleAddUserItem)
	// router.GET("/getEntries", _handleGetIngredients)
	// router.POST("/newIngredient", _handleNewIngredient)
}
