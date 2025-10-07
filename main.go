package main

import (
	"JGLSite/api"
	"JGLSite/test"
	"JGLSite/utils"
	"context"
	"errors"
	"fmt"
	"github.com/JGLTechnologies/SimpleFiles"
	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var store *persist.MemoryStore

const port string = ":81"
const cacheTime = time.Minute * 5

func main() {
	// Load env variables and setup databases
	godotenv.Load("/var/www/.env")
	defer utils.GetDB().Close()
	store = persist.NewMemoryStore(time.Minute)

	// Create HTML templates
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "go web files/home.html", "go web files/base.html")
	r.AddFromFiles("client-error", "go web files/client_error.html")
	r.AddFromFiles("contact", "go web files/contact.html", "go web files/base.html")
	r.AddFromFiles("status", "go web files/status.html")
	r.AddFromFiles("contact-thank-you", "go web files/thank-you.html")
	r.AddFromFiles("jna", "go web files/jna.html")
	r.AddFromFiles("contact-limit", "go web files/limit.html")
	r.AddFromFiles("contact-captcha", "go web files/captcha.html")
	r.AddFromFiles("contact-bl", "go web files/bl.html")
	r.AddFromFiles("contact-spam", "go web files/spam.html")
	r.AddFromFiles("error", "go web files/error.html")
	r.AddFromFiles("bmi-home", "go web files/bmi/build/index.html")
	r.AddFromFiles("kbs", "go web files/kbs.html", "go web files/base.html")

	// Router config
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	router.HTMLRender = r
	router.Use(gin.Logger())
	router.RedirectFixedPath = true
	router.RedirectTrailingSlash = true
	router.HandleMethodNotAllowed = true
	f, _ := SimpleFiles.New("cloudflare_ips.txt", nil)
	s, _ := f.ReadString()
	ips := strings.Split(s, "\n")
	ips = append(ips, "127.0.0.1")
	router.SetTrustedProxies(ips)
	router.ForwardedByClientIP = true
	router.RemoteIPHeaders = []string{"X-Forwarded-For"}

	// Error handler
	router.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		err = strings.Split(err.(error).Error(), "runtime/debug.Stack()")[0]
		if utils.StartsWith(c.Request.URL.String(), "/api") {
			c.AbortWithStatusJSON(500, gin.H{"error": err})
		} else {
			id, _ := uuid.NewRandom()
			errStruct := &utils.Err{Message: err.(string), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id.String(), IP: c.ClientIP(), Path: c.Request.URL.String()}
			utils.Pool.Submit(func() {
				utils.DB.Create(errStruct)
			})
			c.HTML(500, "error", gin.H{"id": errStruct.ID})
			c.AbortWithStatus(500)
		}
	}))

	// Cache static files
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Next()
	})

	// Favicon files
	router.Static("/favicon", "./static/favicon")

	// Routes
	router.GET("/jnu", gin.BasicAuth(map[string]string{"jgl": os.Getenv("pass")}), jnu)
	router.GET("/jna", gin.BasicAuth(map[string]string{"jgl": os.Getenv("pass")}), jna)
	router.GET("/jnau", utils.AllowCors, jnau)
	router.GET("/", cache.CacheByRequestPath(store, cacheTime), home)
	router.GET("/home", cache.CacheByRequestPath(store, cacheTime), home)
	router.GET("/ksp_land_down", cache.CacheByRequestPath(store, cacheTime), kspLandDown)
	router.GET("/contact", cache.CacheByRequestPath(store, cacheTime), contact)
	router.GET("/keyboardsoundplayer", cache.CacheByRequestPath(store, cacheTime), ksp)
	router.GET("/keyboardsoundplayeryoutube", cache.CacheByRequestPath(store, cacheTime), kspYoutube)
	router.GET("/keyboardsoundplayerstore", cache.CacheByRequestPath(store, cacheTime), kspStore)
	router.GET("/robots.txt", cache.CacheByRequestPath(store, cacheTime), func(c *gin.Context) {
		c.File("static/robots.txt")
	})
	router.GET("/favicon.ico", cache.CacheByRequestPath(store, cacheTime), favicon)
	router.GET("/keyboardsoundplayer/vm_exe", cache.CacheByRequestPath(store, cacheTime), func(c *gin.Context) {
		c.File("static/voicemeeterprosetup.exe")
	})
	router.GET("/logo.png", cache.CacheByRequestPath(store, cacheTime), logo)
	router.GET("/ksp_logo.png", cache.CacheByRequestPath(store, cacheTime), kspLogo)
	router.GET("/domain_ownership_verification", func(c *gin.Context) {
		c.String(200, "This domain is owned and managed by JGL Technologies LLC. Email gluca@jgltechnologies for more info.")
	})

	// Testing Group
	testGroup := router.Group("/test")
	{
		testGroup.GET("/bmi", cache.CacheByRequestPath(store, cacheTime), test.BMIHome)
		testGroup.GET("/bmi/static/main.js", cache.CacheByRequestPath(store, cacheTime), test.BMIJS)
	}

	// API Group
	apiGroup := router.Group("/api")
	apiGroup.Use(utils.AllowCors)
	{
		apiGroup.GET("/bot/status", cache.CacheByRequestPath(store, time.Second*5), api.BotStatus)
		apiGroup.GET("/jna", api.JNA)
		apiGroup.GET("/bot/info", cache.CacheByRequestPath(store, time.Second*5), api.BotInfo)
		apiGroup.POST("/traffic", cache.CacheByRequestPath(store, time.Second*5), api.CFProxy)
		apiGroup.POST("/contact", utils.GetMW(time.Second, 1), utils.ReqIDMiddleware, api.Contact)
		apiGroup.GET("/error", cache.CacheByRequestURI(store, cacheTime), api.GetErr)
	}

	// 404 and 405 Handling
	router.NoRoute(noRoute)
	router.NoMethod(noMethod)

	// Server Config
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func jnu(c *gin.Context) {
	user := os.Getenv("sshuser")
	password := os.Getenv("sshpass")
	host := "192.168.1.173:22"

	// Configure client
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		c.String(500, fmt.Sprintf("Error: %v", err))
		return
	}
	defer client.Close()

	// Create a new session
	session, err := client.NewSession()
	if err != nil {
		c.String(500, fmt.Sprintf("Error: %v", err))
		return
	}
	defer session.Close()

	// Run a command on the remote host
	err = session.Start("bash -c 'export DISPLAY=:0; export XAUTHORITY=/home/pi/.Xauthority; sudo pkill firefox-esr; sudo xhost +; sudo unclutter -display :0 -idle 0 -root & firefox-esr --kiosk /var/www/drive/jglnews.html &'")
	if err != nil {
		c.String(500, fmt.Sprintf("Error: %v", err))
		return
	}
	c.String(200, "Success")
}

// KeyboardSoundPlayer

func kspYoutube(c *gin.Context) {
	c.Redirect(301, "https://youtu.be/GeKuPdFSxAM")
}

func kspStore(c *gin.Context) {
	c.Redirect(301, "https://apps.microsoft.com/detail/9pfsjgvshm0l?hl=en-US&gl=US")
}

func ksp(c *gin.Context) {
	c.HTML(200, "kbs", gin.H{})
}

func kspLandDown(c *gin.Context) {
	c.File("go web files/ksp_landing_download.html")
}

// Files

func favicon(c *gin.Context) {
	c.File("static/favicon/favicon.ico")
}

func logo(c *gin.Context) {
	c.File("static/logo.png")
}

func kspLogo(c *gin.Context) {
	c.File("static/ksp_logo.png")
}

// Main website pages

func home(c *gin.Context) {
	c.HTML(200, "home", gin.H{})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

// JGL News

func jna(c *gin.Context) {
	c.HTML(200, "jna", gin.H{})
}

func jnau(c *gin.Context) {
	if c.GetHeader("Pass") != os.Getenv("pass") {
		c.String(403, "Incorrect Password")
		return
	}
	f, _ := SimpleFiles.New("jna.json", nil)
	s, _ := f.ReadString()
	if s == "" {
		f.WriteString("[]")
	}
	var announcements []api.Announcement
	f.ReadJSON(&announcements)
	exp, _ := strconv.Atoi(c.GetHeader("Expire"))
	n := api.Announcement{c.GetHeader("Title"), c.GetHeader("Body"), time.Now().Unix(), time.Now().Unix() + int64(exp*3600)}
	announcements = append(announcements, n)
	f.WriteJSON(announcements)
	c.String(200, "Success")
}

// Error handlers

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
	if c.Request.Method == http.MethodOptions {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		reqHdrs := c.GetHeader("Access-Control-Request-Headers")
		if reqHdrs == "" {
			reqHdrs = "Content-Type, Authorization, Key, Pass"
		}
		c.Header("Access-Control-Allow-Headers", reqHdrs)

		c.AbortWithStatus(http.StatusNoContent) // 204 and STOP
		return
	}

	if strings.HasPrefix(c.Request.URL.Path, "/api") {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
		return
	}
	c.HTML(http.StatusMethodNotAllowed, "status", gin.H{
		"code":    "405",
		"message": "The method you used is not allowed.",
	})
}
