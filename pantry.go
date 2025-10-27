package main

import (
	"log"
	"pantree/api/db"

	"github.com/gin-gonic/gin"
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
		UserID: getPgtypeUuid(userUuid),
	})

	if err != nil {
		log.Println("Could not get pantry:", err)
		c.JSON(500, gin.H{
			"message": "Could not get pantry.",
		})
	}

	if pantry == nil {
		pantry = []db.Userpantryview{}
	}

	c.JSON(200, pantry)
}

// /**
//  * /newIngredient
//  */
// type NewIngredientRequest struct {
// 	Name           string      `json:"name" binding:"required"`
// 	Unit           db.UnitType `json:"unit" binding:"required"`
// 	StorageLoc     db.LocType  `json:"storageLoc" binding:"required"`
// 	IngredientType db.GrocType `json:"ingredientType" binding:"required"`
// 	ImagePath      string      `json:"imagePath"`
// }

// func _handleAddPantryItem(c *gin.Context) {
// 	userUuid, err := getUserId(c)

// 	if err != nil {
// 		log.Println("Unable to get user UUID: \n", err)
// 	}

// 	var request NewIngredientRequest
// 	if err := c.BindJSON(&request); err != nil {
// 		log.Println("Invalid request body: \n", err)
// 		c.JSON(400, gin.H{
// 			"message": "Invalid request body.",
// 		})
// 		return
// 	}

// 	log.Printf("Creating new ingredient for user %s\n", userUuid)

// 	newIngredient, err := queries.CreateIngredient(c, db.CreateIngredientParams{
// 		UserID:         getPgtypeUuid(userUuid),
// 		Name:           request.Name,
// 		Unit:           request.Unit,
// 		StorageLoc:     request.StorageLoc,
// 		IngredientType: request.IngredientType,
// 		ImagePath:      getPgtypeText(request.ImagePath),
// 	})

// 	if err != nil {
// 		log.Println("Could not create ingredient: ", err)
// 		c.JSON(500, gin.H{
// 			"message": "Could not create ingredient.",
// 		})
// 		return
// 	}

// 	log.Printf("Successfully created ingredient: %v\n", newIngredient)
// 	c.JSON(200, newIngredient)
// }

func registerPantryRoutes(router *gin.RouterGroup) {
	router.GET("/getPantry", _handleGetPantry)
	// router.GET("/getEntries", _handleGetIngredients)
	// router.POST("/newIngredient", _handleNewIngredient)
}
