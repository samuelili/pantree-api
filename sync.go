package main

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"pantree/api/db"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type SyncState struct {
	Items []db.Useritem `json:"items" binding:"required"`
}

type SyncOperations struct {
	ItemsToAdd    []db.Useritem `json:"itemsToAdd"`
	ItemsToUpdate []db.Useritem `json:"itemsToUpdate"`
	ItemsToDelete []db.Useritem `json:"itemsToDelete"`
}

func _computeSyncStateHash(syncState SyncState) (uint32, error) {
	// convert SyncState to bytes
	buf := new(bytes.Buffer)

	// write for items
	for _, item := range syncState.Items {
		err := binary.Write(buf, binary.BigEndian, item.ID)
		if err != nil {
			return 0, err
		}

		err = binary.Write(buf, binary.BigEndian, item.LastModified.UnixMilli())
		if err != nil {
			return 0, err
		}
	}

	hash := fnv.New32a()
	hash.Write(buf.Bytes())

	return hash.Sum32(), nil
}

func _pullDbSyncState(c *gin.Context) (SyncState, error) {
	userUuid, err := getUserId(c)
	if err != nil {
		return SyncState{}, err
	}

	// get items

	items, err := queries.GetUserItems(c, &userUuid)

	if err != nil {
		return SyncState{}, err
	}

	if items == nil {
		items = []db.Useritem{}
	}

	var syncState SyncState
	syncState.Items = items

	return syncState, nil
}

type SyncRequest struct {
	Items        []db.Useritem `json:"items" binding:"required"`
	LastSyncTime time.Time     `json:"lastSyncTime" binding:"required"`
}

func sync(c *gin.Context) {
	userUuid, err := getUserId(c)
	if err != nil {
		sendError(c, 400, err, "Invalid user")
		return
	}

	// on device
	var request SyncRequest

	if err := c.BindJSON(&request); err != nil {
		sendError(c, 400, err, "Invalid request parameters")
		return
	}

	for _, item := range request.Items {
		_, err := queries.UpsertUserItem(c, db.UpsertUserItemParams(item))

		if err != nil && err != pgx.ErrNoRows {
			sendError(c, 500, err, "Unable to upsert item")
			return
		}
	}

	toSyncItems, err := queries.GetUserItemsSinceTime(c, db.GetUserItemsSinceTimeParams{
		UserID:       &userUuid,
		LastModified: request.LastSyncTime,
	})

	if err != nil {
		sendError(c, 500, err, "Unable to get user items since time")
		return
	}

	if toSyncItems == nil {
		toSyncItems = []db.Useritem{}
	}

	c.JSON(200, toSyncItems)
}

func syncStateHash(c *gin.Context) {
	syncState, err := _pullDbSyncState(c)

	if err != nil {
		sendError(c, 500, err, "Unable to pull sync state from database")
		return
	}

	hash, err := _computeSyncStateHash(syncState)

	if err != nil {
		sendError(c, 500, err, "Unable to compute sync state hash")
		return
	}

	c.JSON(200, gin.H{
		"hash": hash,
	})
}

func registerSyncRoutes(router *gin.RouterGroup) {
	router.POST("/sync", sync)
	router.GET("/hash", syncStateHash)
}
