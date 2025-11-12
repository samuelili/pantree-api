package main

import (
	"pantree/api/db"

	"github.com/gin-gonic/gin"
)

type SyncState struct {
	Items []db.Useritem `json:"items" binding:"required"`
	User  db.User       `json:"user"  binding:"required"`
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

	items, err := queries.GetUserItems(c, getPgtypeUuid(userUuid))

	if err != nil {
		sendError(c, 500, err, "Could not get pantry items")
		return
	}

	if items == nil {
		items = []db.Useritem{}
	}

	var syncState SyncState
	syncState.Items = items
	syncState.User = user

	c.JSON(200, syncState)
}

func registerSyncRoutes(router *gin.RouterGroup) {
	router.GET("/pull", pull)
}
