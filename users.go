package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"pantree/api/db"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func createOrGetNewUser(ctx context.Context, email string) (*db.User, error) {
	user, err := queries.GetUser(ctx, db.GetUserParams{
		Email: pgtype.Text{
			String: email,
			Valid:  true,
		},
	})

	if err != nil {
		log.Println("Failed to get new user, likely does not exist:\n\t", err)

		// create user if it doesn't exist
		user, err = queries.CreateUser(ctx, db.CreateUserParams{
			Email:       email,
			Name:        email,
			PrefMeasure: "metric",
		})

		if err != nil {
			log.Println("Failed to create new user", err)
			return nil, err
		}

		log.Println("Created new user: ", user)
	}

	return &user, nil
}

func uploadUserImage(c *gin.Context) {
	// read in image
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		sendError(c, http.StatusBadRequest, err, "image")
	}
	defer file.Close()

	// get user id
	userUuid, err := getUserId(c)
	if err != nil {
		sendError(c, http.StatusUnauthorized, err, "Could not determine user")
		return
	}

	// maintain one profile picture per user at any given time in S3 bucket
	user, err := queries.GetUser(c, db.GetUserParams{
		ID: &userUuid,
	})

	if user.ProfilePic.Valid == true {
		err = deleteS3(user.ProfilePic.String)
		if err != nil {
			log.Println("Could not delete image from S3 bucket", err)
			return
		}
	}

	// create image key
	imageKey := fmt.Sprintf(
		"users/%s%s",
		userUuid.String(),
		path.Ext(header.Filename),
	)

	// update image key in db
	err = queries.UpdateUser(c, db.UpdateUserParams{
		ID:         userUuid,
		ProfilePic: pgtype.Text{String: imageKey, Valid: true},
	})

	if err != nil {
		log.Println("Could not save profile picture to db", err)
		return
	}

	// upload image to s3
	err = uploadS3(imageKey, header.Header.Get("Content-Type"), file)
	if err != nil {
		sendError(c, http.StatusInternalServerError, err, "Failed to upload image")
		return
	}

	c.JSON(http.StatusOK, gin.H{"imageKey": imageKey})
}

func handleMe(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
		return
	}

	log.Printf("Getting user %s\n", userUuid)
	user, err := queries.GetUser(c, db.GetUserParams{
		ID: &userUuid,
	})

	if err != nil {
		log.Println("Could not get user: ", err)
		c.JSON(500, gin.H{
			"message": "Could not get user.",
		})
		return
	}

	c.JSON(200, user)
}

type UpdateUserRequest struct {
	Email       string `form:"email" json:"email"`
	Name        string `form:"name" json:"name"`
	PrefMeasure string `form:"prefMeasure" json:"prefMeasure"`
}

func handleUpdateMe(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
		c.JSON(500, gin.H{
			"message": "Could not get user UUID",
			"error":   err,
		})
		return
	}

	var request UpdateUserRequest
	if err := c.BindJSON(&request); err != nil {
		log.Println("Could not read JSON:\n", err)
		c.JSON(400, gin.H{
			"message": "Could not read JSON",
			"error":   err,
		})
		return
	}

	var params = db.UpdateUserParams{
		ID: userUuid,
	}

	if request.Email != "" {
		params.Email = pgtype.Text{
			String: request.Email,
			Valid:  true,
		}
	}

	if request.Name != "" {
		params.Name = pgtype.Text{
			String: request.Name,
			Valid:  true,
		}
	}

	if request.PrefMeasure != "" {
		params.PrefMeasure = db.NullMeasureType{
			MeasureType: db.MeasureType(request.PrefMeasure),
			Valid:       true,
		}
	}

	err = queries.UpdateUser(c, params)

	if err != nil {
		log.Println("Unable to update user"+userUuid.String()+": \n", err)
		c.JSON(500, gin.H{
			"message": "Unable to update user",
			"error":   err,
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Success",
	})
}

func registerUserRoutes(router *gin.RouterGroup) {
	router.GET("me", handleMe)
	router.POST("updateMe", handleUpdateMe)
	router.POST("uploadImage", uploadUserImage)
}
