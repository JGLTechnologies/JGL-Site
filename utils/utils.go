package utils

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Err struct {
	ID      string `gorm:"primaryKey" json:"id"`
	Message string `json:"message"`
	Date    string `json:"date"`
	Path    string `json:"path"`
	IP      string `json:"ip"`
}

type Announcement struct {
	ID     uint   `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Title  string `json:"title" gorm:"not null"`
	Body   string `json:"body" gorm:"not null"`
	Time   int64  `json:"time" gorm:"not null"`
	Expire int64  `json:"expire" gorm:"not null"`
}

type Email struct {
	Email string `json:"email" gorm:"primaryKey"`
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
	c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Key")
}

func InitDB() (*sql.DB, error) {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=192.168.1.163 user=%s password=%s dbname=jgldb port=5432 sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})

	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&Err{})
	err2 := db.AutoMigrate(&Announcement{})
	err3 := db.AutoMigrate(&Email{})

	if err != nil || err2 != nil || err3 != nil {
		panic("failed to migrate database")
	}

	sqlDB, err := db.DB()

	if err != nil {
		panic("failed to get database connection")
	}

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	DB = db
	DB.Where("expire < ?", time.Now().Unix()).Delete(&Announcement{})
	return sqlDB, sqlDB.Ping()
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
