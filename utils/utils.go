package utils

import (
	"database/sql"
	"net/http"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gammazero/workerpool"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB   *gorm.DB
	Pool = workerpool.New(20)
)

type Err struct {
	ID      string `gorm:"primaryKey" json:"id"`
	Message string `json:"message"`
	Date    string `json:"date"`
	Path    string `json:"path"`
	IP      string `json:"ip"`
}

var ReqIDMiddleware = requestid.New(requestid.WithGenerator(func() string {
	id, _ := uuid.NewRandom()
	return id.String()
}))

func AllowCors(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin == "" {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", origin)
	}

	c.Header("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

	reqHdrs := c.GetHeader("Access-Control-Request-Headers")
	if reqHdrs == "" {
		reqHdrs = "Content-Type, Authorization, Key, Pass"
	}
	c.Header("Access-Control-Allow-Headers", reqHdrs)

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
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
	}

	if s[:swLen] == sw {
		return true
	}

	return false
}

func GetMW(rate time.Duration, limit uint) func(c *gin.Context) {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  rate,
		Limit: limit,
	})
	return ratelimit.RateLimiter(store, &ratelimit.Options{})
}
