package utils

import (
	"database/sql"
	"github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gammazero/workerpool"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"time"
)

var client = req.C().SetTimeout(time.Second * 5)
var DB *gorm.DB
var Pool = workerpool.New(20)

type Err struct {
	ID      string `gorm:"primaryKey" json:"id"`
	Message string `json:"message"`
	Date    string `json:"date"`
	Path    string `json:"path"`
	IP      string `json:"ip"`
}

// Middleware to give each request a unique id
var ReqIDMiddleware = requestid.New(requestid.WithGenerator(func() string {
	id, _ := uuid.NewRandom()
	return id.String()
}))

// Middleware to allow CORS
func AllowCors(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin == "" {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		// reflect origin (use this if you might enable credentials later)
		c.Header("Access-Control-Allow-Origin", origin)
	}

	c.Header("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")

	// must be an explicit list — no '*'
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

	// echo what the browser requested (handles custom headers like "Key" or "Pass")
	reqHdrs := c.GetHeader("Access-Control-Request-Headers")
	if reqHdrs == "" {
		// sensible defaults if no preflight header is present
		reqHdrs = "Content-Type, Authorization, Key, Pass"
	}
	c.Header("Access-Control-Allow-Headers", reqHdrs)

	// If you ever send cookies/Authorization with credentials from JS:
	// c.Header("Access-Control-Allow-Credentials", "true")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent) // 204
		return
	}
	c.Next()
}

func GetDB() *sql.DB {
	DB, _ = gorm.Open(sqlite.Open("errors.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	DB.AutoMigrate(&Err{})
	sqlDB, _ := DB.DB()
	return sqlDB
}

func StartsWith(s string, sw string) bool {
	swLen := len(sw)
	sLen := len(s)
	if swLen > sLen {
		return false
	} else if s[:swLen] == sw {
		return true
	} else {
		return false
	}
}

// Returns the ratelimt middleware
func GetMW(rate time.Duration, limit uint) func(c *gin.Context) {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  rate,
		Limit: limit,
	})
	return ratelimit.RateLimiter(store, &ratelimit.Options{})
}
