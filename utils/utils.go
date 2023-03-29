package utils

import (
	"database/sql"
	"github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/imroc/req/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strconv"
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

func GetPythonLibDownloads(project string) string {
	var data map[string]interface{}
	res, err := client.R().Get("https://api.pepy.tech/api/projects/" + project)
	if err != nil || res.IsErrorState() {
		return "Not Found"
	}
	jsonErr := res.UnmarshalJson(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["total_downloads"].(float64)))
}

func GetNPMLibDownloads(project string) string {
	var date string
	date += strconv.Itoa(time.Now().Year())
	date += "-" + strconv.Itoa(int(time.Now().Month()))
	date += "-" + strconv.Itoa(time.Now().Day())
	var data map[string]interface{}
	res, err := client.R().Get("https://api.npmjs.org/downloads/point/2020-1-1:" + date + "/" + project)
	if err != nil || res.IsError() {
		return "Not Found"
	}
	jsonErr := res.UnmarshalJson(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["downloads"].(float64)))
}

func GetGoLibDownloads(project string) string {
	var data map[string]interface{}
	res, err := client.R().SetHeader("Authorization", "token "+os.Getenv("gh_token")).Get("https://api.github.com/repos/JGLTechnologies/" + project + "/traffic/clones?per=week")
	if err != nil || res.IsError() {
		return "Not Found"
	}
	jsonErr := res.UnmarshalJson(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["count"].(float64)))
}

func GetMW(rate time.Duration, limit uint) func(c *gin.Context) {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  rate,
		Limit: limit,
	})
	return ratelimit.RateLimiter(store, &ratelimit.Options{})
}
