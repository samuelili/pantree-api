package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"
)

/**
inspired by https://ieeexplore.ieee.org/abstract/document/10404196
*/

var (
	identityKey = "id"
)

type login struct {
	Email string `form:"email" json:"email" binding:"required,email"`
	Otp   string `form:"otp" json:"otp" binding:"required,len=6"`
	Hash  string `form:"hash" json:"hash" binding:"required"`
}

type JwtUser struct {
	Id string `form:"id" json:"id" binding:"required"`
}

func authenticator() func(c *gin.Context) (interface{}, error) {

	return func(c *gin.Context) (interface{}, error) {
		var loginVals login
		if err := c.ShouldBind(&loginVals); err != nil {
			return "", jwt.ErrMissingLoginValues
		}
		email := loginVals.Email
		hashStr := loginVals.Hash
		otp := loginVals.Otp

		expectedHash, err := hex.DecodeString(hashStr)
		if err != nil {
			return nil, jwt.ErrFailedAuthentication
		}

		// look through the next 5 timed OTPs
		match := false
		for off := 0; off < EXPR_OFF; off++ {
			testedHash := hashOtp(loginVals.Email, otp, off)
			if hmac.Equal(testedHash, expectedHash) {
				match = true
				break
			}
		}

		if !match {
			return nil, jwt.ErrFailedAuthentication
		}

		// success!
		user, err := createOrGetNewUser(c, email)

		if err != nil {
			return nil, jwt.ErrFailedAuthentication
		}

		return &JwtUser{
			Id: user.ID.String(),
		}, nil
	}
}

func authorizator() func(data interface{}, c *gin.Context) bool {
	return func(data interface{}, c *gin.Context) bool {
		// just verify that data is of type *User for now
		log.Println("asdf", data)
		if _, ok := data.(*JwtUser); ok {
			return true
		}
		return false
	}
}

func unauthorized() func(c *gin.Context, code int, message string) {
	return func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func payloadFunc() func(data interface{}) jwt.MapClaims {
	return func(data interface{}) jwt.MapClaims {
		if v, ok := data.(*JwtUser); ok {
			return jwt.MapClaims{
				identityKey: v.Id,
			}
		}
		return jwt.MapClaims{}
	}
}

func initParams() *jwt.GinJWTMiddleware {

	return &jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: payloadFunc(),

		Authenticator: authenticator(),
		Unauthorized:  unauthorized(),
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}
}

func handleNoRoute() func(c *gin.Context) {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	}
}

var secret string

func generateRandomString(length int, charset string) string {
	// Seed the random number generator.
	rand.Seed(time.Now().UnixNano())

	var result strings.Builder

	for i := 0; i < length; i++ {
		// Get a random index from the charset.
		randomIndex := rand.Intn(len(charset))

		// Append the random character.
		result.WriteByte(charset[randomIndex])
	}

	return result.String()
}

func generateOtp() string {
	const charset = "0123456789"
	return generateRandomString(6, charset)
}

const TIME_INT int64 = 60 * 1000
const EXPR_OFF int = 5

func getTimeStep(offset int) int64 {
	now := time.Now().UnixMilli()
	curr := now - now%TIME_INT
	return curr + int64(offset)*TIME_INT
}

func hashOtp(email string, otp string, offset int) []byte {
	hash := hmac.New(sha1.New, []byte(secret))

	hash.Write([]byte(email))
	hash.Write([]byte(otp))
	hash.Write([]byte{byte(getTimeStep(offset))})

	return hash.Sum(nil)
}

func getHashString(hash []byte) string {
	return hex.EncodeToString(hash[:])
}

type RequestOtpParams struct {
	Email string `json:"email" binding:"required,email"`
}

func requestOtp(c *gin.Context) {
	var params RequestOtpParams

	if err := c.BindJSON(&params); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	otp := generateOtp()
	hash := hashOtp(params.Email, otp, EXPR_OFF-1)
	hashStr := getHashString(hash)

	if cfg.Server.SendMail {
		sendEmail(params.Email, "pantree: Your One-Time Password is...", "Your One-Time Password is "+otp+". This will expire in 5 minutes.")
	} else {
		log.Println("DEV: OTP is ", otp, " for email ", params.Email)
	}

	log.Println("OTP Requested, sending hash to user: ", hashStr)

	c.JSON(200, gin.H{
		"hash": hashStr,
	})
}

type LoginParams struct {
	Email string `json:"email" binding:"required,email"`
	Otp   string `json:"otp" binding:"required,len=6"`
	Hash  string `json:"hash" binding:"required"`
}

func handlerMiddleware(authMiddleware *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(context *gin.Context) {
		errInit := authMiddleware.MiddlewareInit()
		if errInit != nil {
			log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
		}
	}
}

func getUserId(c *gin.Context) (uuid.UUID, error) {
	claims := jwt.ExtractClaims(c)
	idStr := claims[identityKey].(string)

	parsedUuid, err := uuid.Parse(idStr)

	if err != nil {
		return parsedUuid, err
	}

	return parsedUuid, nil
}

func getPgtypeUuid(uuid uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}
}

func registerAuth(engine *gin.Engine) *jwt.GinJWTMiddleware {
	secret = os.Getenv("SECRET")

	// register jwt middleware
	authMiddleware, err := jwt.New(initParams())
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	engine.Use(handlerMiddleware(authMiddleware))

	engine.NoRoute(authMiddleware.MiddlewareFunc(), handleNoRoute())

	engine.POST("/requestOtp", requestOtp)
	engine.POST("/login", authMiddleware.LoginHandler)
	auth := engine.Group("/auth", authMiddleware.MiddlewareFunc())
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)

	return authMiddleware
}
