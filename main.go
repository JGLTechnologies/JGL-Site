package main

import (
	"JGLSite/api"
	"JGLSite/test"
	"JGLSite/utils"
	"fmt"
	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"time"
)

var Store *persist.MemoryStore

func main() {
	godotenv.Load("../.env")
	gin.SetMode(gin.ReleaseMode)
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "go web files/home.html", "go web files/base.html")
	r.AddFromFiles("contact", "go web files/contact.html", "go web files/base.html")
	r.AddFromFiles("404", "go web files/404.html")
	r.AddFromFiles("bmi-home", "go web files/bmi/index.html")
	r.AddFromFiles("bmi-calc", "go web files/bmi/bmi.html")
	r.AddFromFiles("bmi-invalid", "go web files/bmi/invalid.html")
	r.AddFromFiles("contact-thank-you", "go web files/thank-you.html")
	r.AddFromFiles("contact-limit", "go web files/limit.html")
	r.AddFromFiles("contact-captcha", "go web files/captcha.html")
	r.AddFromFiles("contact-bl", "go web files/bl.html")
	r.AddFromFiles("contact-error", "go web files/error.html")
	server := gin.New()
	server.HTMLRender = r
	Store = persist.NewMemoryStore(time.Hour)

	server.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.HTML(500, "contact-error", gin.H{"error": fmt.Sprintf("%s", err)})
		c.AbortWithStatus(500)
	}))
	server.Use(utils.LoggerWithConfig(gin.LoggerConfig{}))
	server.SetTrustedProxies([]string{"192.168.1.252", "127.0.0.1", "192.168.1.1"})

	server.GET("/", cache.CacheByRequestPath(Store, time.Minute*10), home)
	server.GET("/home", cache.CacheByRequestPath(Store, time.Minute*10), home)
	server.GET("/contact", cache.CacheByRequestPath(Store, time.Hour*24), contact)
	server.GET("/logo.png", cache.CacheByRequestPath(Store, time.Hour*24), logo)
	server.GET("/favicon.ico", cache.CacheByRequestPath(Store, time.Hour*24), favicon)

	testGroup := server.Group("/test")
	{
		testGroup.GET("/bmi", test.BMIHome)
		testGroup.GET("/bmi/calc", test.BMICalc)
	}

	apiGroup := server.Group("/api")
	{
		apiMW := utils.GetMW(1, 10)
		apiGroup.GET("/bot/status", cache.CacheByRequestPath(Store, time.Minute), api.BotStatus)
		apiGroup.GET("/bot/info", cache.CacheByRequestPath(Store, time.Hour), api.BotInfo)
		apiGroup.GET("/dpys", cache.CacheByRequestPath(Store, time.Minute*10), api.DPYS)
		apiGroup.GET("/aiohttplimiter", cache.CacheByRequestPath(Store, time.Minute*10), api.AIOHTTPRateLimiter)
		apiGroup.GET("/GinRateLimit", cache.CacheByRequestPath(Store, time.Minute*10), api.GinRateLimit)
		apiGroup.GET("/precise-memory-rate-limit", cache.CacheByRequestPath(Store, time.Minute*10), api.PreciseMemoryRateLimit)
		apiGroup.GET("/versions", apiMW, versions)
		apiGroup.GET("/downloads", apiMW, downloads)
		apiGroup.POST("/contact", utils.GetMW(1, 1), api.Contact)
	}

	server.NoRoute(noRoute)
	server.NoMethod(noRoute)
	go updateVersionsAndDownloads()
	time.Sleep(time.Second * 3)
	log.Fatal(server.Run(":81"))
}

func updateVersionsAndDownloads() {
	for {
		Store.Set("versions", utils.Versions(Store), -1)
		Store.Set("downloads", map[string]string{
			"dpys":                      utils.GetPythonLibDownloads("dpys", Store),
			"aiohttp-ratelimiter":       utils.GetPythonLibDownloads("aiohttp-ratelimiter", Store),
			"precise-memory-rate-limit": utils.GetNPMLibDownloads("precise-memory-rate-limit", Store),
			"GinRateLimit":              utils.GetGoLibDownloads("GinRateLimit", Store),
		}, -1)
		time.Sleep(time.Minute * 10)
	}
}

func versions(c *gin.Context) {
	var data map[string]string
	Store.Get("versions", &data)
	c.JSON(200, data)
}

func downloads(c *gin.Context) {
	var data map[string]string
	Store.Get("downloads", &data)
	c.JSON(200, data)
}

func favicon(c *gin.Context) {
	c.File("static/favicon.ico")
}

func logo(c *gin.Context) {
	c.File("static/logo.png")
}

func home(c *gin.Context) {
	var downloads map[string]string
	var versions map[string]string

	Store.Get("downloads", &downloads)
	Store.Get("versions", &versions)

	c.HTML(200, "home", gin.H{
		"dpys_downloads":           downloads["dpys"],
		"aiohttplimiter_downloads": downloads["aiohttp-ratelimiter"],
		"grl_downloads":            downloads["GinRateLimit"],
		"pmrl_downloads":           downloads["precise-memory-rate-limit"],

		"dpys_version":           versions["dpys"],
		"aiohttplimiter_version": versions["aiohttp-ratelimiter"],
		"grl_version":            versions["GinRateLimit"],
		"pmrl_version":           versions["precise-memory-rate-limit"],
	})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

func noRoute(c *gin.Context) {
	c.HTML(404, "404", gin.H{})
}
