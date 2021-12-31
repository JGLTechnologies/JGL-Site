package main

import (
	"JGLSite/api"
	"JGLSite/test"
	"JGLSite/utils"
	"fmt"
	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"time"
)

var store *persist.MemoryStore

func main() {
	godotenv.Load("/var/www/.env")
	gin.SetMode(gin.ReleaseMode)
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "go web files/home.html", "go web files/base.html")
	r.AddFromFiles("projects", "go web files/projects.html", "go web files/base.html")
	r.AddFromFiles("client-error", "go web files/client_error.html")
	r.AddFromFiles("contact", "go web files/contact.html", "go web files/base.html")
	r.AddFromFiles("status", "go web files/status.html")
	r.AddFromFiles("contact-thank-you", "go web files/thank-you.html")
	r.AddFromFiles("contact-limit", "go web files/limit.html")
	r.AddFromFiles("contact-captcha", "go web files/captcha.html")
	r.AddFromFiles("contact-bl", "go web files/bl.html")
	r.AddFromFiles("error", "go web files/error.html")
	r.AddFromFiles("bmi-home", "go web files/bmi/build/index.html")
	router := gin.New()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	router.Use(cors.New(config))
	router.HTMLRender = r
	store = persist.NewMemoryStore(time.Hour)

	router.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.HTML(500, "error", gin.H{"error": fmt.Sprintf("%s", err)})
		c.AbortWithStatus(500)
	}))
	router.Use(utils.LoggerWithConfig(gin.LoggerConfig{}))
	router.SetTrustedProxies([]string{"192.168.1.252", "127.0.0.1", "192.168.1.1"})
	router.HandleMethodNotAllowed = true

	router.GET("/", cache.CacheByRequestPath(store, time.Minute*10), home)
	router.GET("/bot", func(c *gin.Context) {
		c.String(200, "Bot webpage is coming soon.")
	})
	router.GET("/home", cache.CacheByRequestPath(store, time.Hour*24), home)
	router.GET("/projects", cache.CacheByRequestPath(store, time.Hour*24), projects)
	router.GET("/contact", cache.CacheByRequestPath(store, time.Hour*24), contact)
	router.GET("/logo.png", cache.CacheByRequestPath(store, time.Hour*24), logo)
	router.GET("/favicon.ico", cache.CacheByRequestPath(store, time.Hour*24), favicon)

	static := router.Group("/static")
	{
		static.GET("/projects.js", cache.CacheByRequestPath(store, time.Hour*24), projectsJS)
	}

	testGroup := router.Group("/test")
	{
		testGroup.GET("/bmi", cache.CacheByRequestPath(store, time.Hour*24), test.BMIHome)
		testGroup.GET("/bmi/static/main.js", cache.CacheByRequestPath(store, time.Hour*24), test.BMIJS)
	}

	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/bot/status", cache.CacheByRequestPath(store, time.Minute), api.BotStatus)
		apiGroup.GET("/bot/info", cache.CacheByRequestPath(store, time.Hour), api.BotInfo)
		apiGroup.GET("/downloads", cache.CacheByRequestPath(store, time.Minute*10), api.Downloads)
		apiGroup.POST("/contact", utils.GetMW(1, 1), api.Contact)
		apiGroup.GET("/projects", cache.CacheByRequestPath(store, time.Minute*10), api.Projects)
	}

	router.NoRoute(noRoute)
	router.NoMethod(noMethod)
	srv := &http.Server{
		Addr:         ":81",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}

func favicon(c *gin.Context) {
	c.File("static/favicon.ico")
}

func logo(c *gin.Context) {
	c.File("static/logo.png")
}

func home(c *gin.Context) {
	c.HTML(200, "home", gin.H{})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

func projects(c *gin.Context) {
	c.HTML(200, "projects", gin.H{})
}

func projectsJS(c *gin.Context) {
	c.File("static/projects.js")
}

func noRoute(c *gin.Context) {
	if utils.StartsWith(c.Request.URL.String(), "/api") {
		c.JSON(404, gin.H{"error": "Not Found"})
	} else {
		c.HTML(404, "status", gin.H{
			"code":    "404",
			"message": "The page you requested does not exist.",
		})
	}
}

func noMethod(c *gin.Context) {
	if utils.StartsWith(c.Request.URL.String(), "/api") {
		c.JSON(405, gin.H{"error": "Method Not Allowed"})
	} else {
		c.HTML(405, "status", gin.H{
			"code":    "405",
			"message": "The page you used is not allowed.",
		})
	}
}
