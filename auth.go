package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"math/rand"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

/**
inspired by https://ieeexplore.ieee.org/abstract/document/10404196
*/

var (
	identityKey = "id"
	port        string
)

// User demo
type User struct {
	Email string
}

func identityHandler() func(c *gin.Context) interface{} {
	return func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		return &User{
			Email: claims[identityKey].(string),
		}
	}
}

type login struct {
	Email string `form:"email" json:"email" binding:"required,email"`
	Otp   string `form:"otp" json:"otp" binding:"required,len=6"`
	Hash  string `form:"hash" json:"hash" binding:"required"`
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

		expectedHash := hashOtp(email, otp)
		providedHash, err := hex.DecodeString(hashStr)
		if err != nil {
			// c.JSON(400, gin.H{"message": "Invalid hash format", "error": err.Error()})
			return nil, jwt.ErrFailedAuthentication
		}

		if !hmac.Equal(expectedHash, providedHash) {
			// c.JSON(401, gin.H{"message": "Invalid OTP"})
			return nil, jwt.ErrFailedAuthentication
		}

		return &User{
			Email: email,
		}, nil
	}
}

func authorizator() func(data interface{}, c *gin.Context) bool {
	return func(data interface{}, c *gin.Context) bool {
		// just verify that data is of type *User for now
		if _, ok := data.(*User); ok {
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
		if v, ok := data.(*User); ok {
			return jwt.MapClaims{
				identityKey: v.Email,
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

		IdentityHandler: identityHandler(),
		Authenticator:   authenticator(),
		Authorizator:    authorizator(),
		Unauthorized:    unauthorized(),
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",
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

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(identityKey)
	c.JSON(200, gin.H{
		"userID":   claims[identityKey],
		"userName": user.(*User).Email,
		"text":     "Hello World.",
	})
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

func generateSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"

	return generateRandomString(32, charset)
}

func hashOtp(email string, otp string) []byte {
	hash := hmac.New(sha1.New, []byte(secret))

	hash.Write([]byte(email))
	hash.Write([]byte(otp))

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
	hash := hashOtp(params.Email, otp)
	hashStr := getHashString(hash)

	log.Println("REMOVE ME: OTP is ", otp, " for email ", params.Email)

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

var handler *jwt.GinJWTMiddleware

func registerAuth(engine *gin.Engine) {
	// TODO: rotate this secret
	secret = generateSecret()

	// register jwt middleware
	authMiddleware, err := jwt.New(initParams())
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	handler = authMiddleware

	engine.Use(handlerMiddleware(authMiddleware))

	// engine.POST("/login", _login)

	engine.NoRoute(authMiddleware.MiddlewareFunc(), handleNoRoute())

	engine.POST("/requestOtp", requestOtp)
	engine.POST("/login", authMiddleware.LoginHandler)
	auth := engine.Group("/auth", authMiddleware.MiddlewareFunc())
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.GET("/hello", helloHandler)
}
