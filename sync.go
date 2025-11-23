package main

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"pantree/api/db"

	"github.com/gin-gonic/gin"
)

type SyncState struct {
	Items []db.Useritem `json:"items" binding:"required"`
	User  db.User       `json:"user"  binding:"required"`
}

type SyncOperations struct {
	ItemsToAdd    []db.Useritem `json:"itemsToAdd"`
	ItemsToUpdate []db.Useritem `json:"itemsToUpdate"`
	ItemsToDelete []db.Useritem `json:"itemsToDelete"`
	UserToUpdate  *db.User      `json:"userToUpdate"`
}

func _mergeSyncStates(remote SyncState, local SyncState) SyncOperations {
	syncOperations := SyncOperations{
		ItemsToAdd:    []db.Useritem{},
		ItemsToUpdate: []db.Useritem{},
		ItemsToDelete: []db.Useritem{},
		UserToUpdate:  nil,
	}

	// we need to isolate three SETS of items:
	// 1. new items
	// 2. deleted items
	// 3. updated items

	// to prepare, construct a set of IDs for remote and local items
	remoteItemMap := make(map[string]db.Useritem)
	for _, item := range remote.Items {
		remoteItemMap[item.ID.String()] = item
	}
	localItemMap := make(map[string]db.Useritem)
	for _, item := range local.Items {
		localItemMap[item.ID.String()] = item
	}

	// 1. added items will be local set different remote set
	for _, item := range local.Items {
		if _, exists := remoteItemMap[item.ID.String()]; !exists {
			syncOperations.ItemsToAdd = append(syncOperations.ItemsToAdd, item)
		}
	}

	// 2. deleted items will be remote set different local set
	for _, item := range remote.Items {
		if _, exists := localItemMap[item.ID.String()]; !exists {
			syncOperations.ItemsToDelete = append(syncOperations.ItemsToDelete, item)
		}
	}

	// 3. updated items will be items that are in both sets, but have different updated_at timestamps
	for _, localItem := range local.Items {
		if remoteItem, exists := remoteItemMap[localItem.ID.String()]; exists {
			if localItem.LastModified.Time.After(remoteItem.LastModified.Time) {
				syncOperations.ItemsToUpdate = append(syncOperations.ItemsToUpdate, localItem)
			}
		}
	}

	// now we process user difference. we do a simple replace for now
	// TOOD: add last modified
	syncOperations.UserToUpdate = &local.User

	return syncOperations
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

		err = binary.Write(buf, binary.BigEndian, item.LastModified.Time.UnixMilli())
		if err != nil {
			return 0, err
		}
	}

	// now write for user
	err := binary.Write(buf, binary.BigEndian, syncState.User.ID)
	if err != nil {
		return 0, err
	}

	err = binary.Write(buf, binary.BigEndian, syncState.User.LastModified.Time.UnixMilli())
	if err != nil {
		return 0, err
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

	user, err := queries.GetUser(c, db.GetUserParams{
		ID: &userUuid,
	})

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
	syncState.User = user

	return syncState, nil
}

func pull(c *gin.Context) {
	syncState, err := _pullDbSyncState(c)

	if err != nil {
		sendError(c, 500, err, "Unable to pull sync state from database")
		return
	}

	c.JSON(200, syncState)
}

func push(c *gin.Context) {
	var request SyncState
	if err := c.BindJSON(&request); err != nil {
		sendError(c, 400, err, "Invalid request parameters")
		return
	}

	// get current remote sync state
	remoteSyncState, err := _pullDbSyncState(c)
	if err != nil {
		sendError(c, 500, err, "Unable to pull sync state from database")
		return
	}

	// merge states
	syncOperations := _mergeSyncStates(remoteSyncState, request)

	// get synced operations

	c.JSON(200, syncOperations)
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
	router.GET("/pull", pull)
	router.POST("/push", push)
	router.GET("/hash", syncStateHash)
}
