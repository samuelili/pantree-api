package main

import (
	"pantree/api/db"

	"github.com/gin-gonic/gin"
)

type SyncState struct {
	items db.Userpantryview `json:"items" binding:"required"`
	user  db.Useritem       `json:"user"  binding:"required"`
}

func pull(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		sendError(c, 500, err, "Unable to get user UUID")
		return
	}

	// get user

	user, err := queries.GetUser(c, db.GetUserParams{
		ID: getPgtypeUuid(userUuid),
	})

	if err != nil {
		sendError(c, 500, err, "Could not get user")
		return
	}

	// get items

	items, err := queries.GetUserPantry(c, db.GetUserPantryParams{
		UserID: getPgtypeUuid(userUuid),
	})

	if err != nil {
		sendError(c, 500, err, "Could not get pantry items")
		return
	}

	if items == nil {
		items = []db.GetUserPantryRow{}
	}

	c.JSON(200, gin.H{
		"items": items,
		"user":  user,
	})
}

func registerSyncRoutes(router *gin.RouterGroup) {
	router.GET("/pull", pull)
}
