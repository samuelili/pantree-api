package main

import (
	"context"
	"log"
	"pantree/api/db"

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
			Email: email,
			Name:  email,
		})

		if err != nil {
			log.Println("Failed to create new user", err)
			return nil, err
		}

		log.Println("Created new user: ", user)
	}

	return &user, nil
}

func handleMe(c *gin.Context) {
	userUuid, err := getUserId(c)

	if err != nil {
		log.Println("Unable to get user UUID: \n", err)
	}

	log.Printf("Getting user %s\n", userUuid)
	user, err := queries.GetUser(c, db.GetUserParams{
		ID: pgtype.UUID{
			Bytes: userUuid,
			Valid: true,
		},
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
	Email string `form:"email" json:"email"`
	Name  string `form:"name" json:"name"`
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

	var params db.UpdateUserParams
	params.ID = getPgtypeUuid(userUuid)

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
}
