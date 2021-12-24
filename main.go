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

var store *persist.MemoryStore

func main() {
	godotenv.Load("../.env")
	gin.SetMode(gin.ReleaseMode)
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "go web files/home.html", "go web files/base.html")
	r.AddFromFiles("contact", "go web files/contact.html", "go web files/base.html")
	r.AddFromFiles("status", "go web files/status.html")
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
	store = persist.NewMemoryStore(time.Hour)

	server.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.HTML(500, "contact-error", gin.H{"error": fmt.Sprintf("%s", err)})
		c.AbortWithStatus(500)
	}))
	server.Use(utils.LoggerWithConfig(gin.LoggerConfig{}))
	server.SetTrustedProxies([]string{"192.168.1.252", "127.0.0.1", "192.168.1.1"})

	server.GET("/", cache.CacheByRequestPath(store, time.Minute*10), home)
	server.GET("/home", cache.CacheByRequestPath(store, time.Minute*10), home)
	server.GET("/contact", cache.CacheByRequestPath(store, time.Hour*24), contact)
	server.GET("/logo.png", cache.CacheByRequestPath(store, time.Hour*24), logo)
	server.GET("/favicon.ico", cache.CacheByRequestPath(store, time.Hour*24), favicon)

	testGroup := server.Group("/test")
	{
		testGroup.GET("/bmi", test.BMIHome)
		testGroup.GET("/bmi/calc", test.BMICalc)
	}

	apiGroup := server.Group("/api")
	{
		apiGroup.GET("/bot/status", cache.CacheByRequestPath(store, time.Minute), api.BotStatus)
		apiGroup.GET("/bot/info", cache.CacheByRequestPath(store, time.Hour), api.BotInfo)
		apiGroup.GET("/versions", cache.CacheByRequestPath(store, time.Minute*10), api.Versions)
		apiGroup.GET("/downloads", cache.CacheByRequestPath(store, time.Minute*10), api.Downloads)
		apiGroup.POST("/contact", utils.GetMW(1, 1), api.Contact)
	}

	server.NoRoute(noRoute)
	server.NoMethod(noMethod)
	go updateVersionsAndDownloads()
	time.Sleep(time.Second * 3)
	if err := server.Run(":81"); err != nil {
		log.Fatalln(err)
	}
}

func updateVersionsAndDownloads() {
	var dpys string
	var aiohttplimiter string
	var grl string
	var pmrl string

	for {
		dpys = utils.GetPythonLibDownloads("dpys")
		aiohttplimiter = utils.GetPythonLibDownloads("aiohttp-ratelimiter")
		pmrl = utils.GetNPMLibDownloads("precise-memory-rate-limit")
		grl = utils.GetGoLibDownloads("GinRateLimit")
		list := []string{dpys, aiohttplimiter, grl, pmrl}
		store.Set("versions", utils.Versions(), -1)
		store.Set("downloads", map[string]string{
			"dpys":                      dpys,
			"aiohttp-ratelimiter":       aiohttplimiter,
			"precise-memory-rate-limit": pmrl,
			"GinRateLimit":              grl,
			"total":                     api.GetTotal(list),
		}, -1)
		time.Sleep(time.Minute * 10)
	}
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

	store.Get("downloads", &downloads)
	store.Get("versions", &versions)

	c.HTML(200, "home", gin.H{
		"dpys_downloads":           downloads["dpys"],
		"aiohttplimiter_downloads": downloads["aiohttp-ratelimiter"],
		"grl_downloads":            downloads["GinRateLimit"],
		"pmrl_downloads":           downloads["precise-memory-rate-limit"],
		"dpys_version":             versions["dpys"],
		"aiohttplimiter_version":   versions["aiohttp-ratelimiter"],
		"grl_version":              versions["GinRateLimit"],
		"pmrl_version":             versions["precise-memory-rate-limit"],
	})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

func noRoute(c *gin.Context) {
	c.HTML(404, "status", gin.H{
		"code":    "404",
		"message": "The page you requested does not exist.",
	})
}

func noMethod(c *gin.Context) {
	c.HTML(405, "status", gin.H{
		"code":    "405",
		"message": "The method you used is not allowed.",
	})
}
