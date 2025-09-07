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
	"github.com/gin-contrib/requestid"
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

func main() {
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

	reqIDMiddleware := requestid.New(requestid.WithGenerator(func() string {
		id, _ := uuid.NewRandom()
		return id.String()
	}))

	// Routes
	router.GET("/jnu", gin.BasicAuth(map[string]string{"jgl": os.Getenv("pass")}), jnu)
	router.GET("/jna", gin.BasicAuth(map[string]string{"jgl": os.Getenv("pass")}), jna)
	router.GET("/jnau", AllowCors, jnau)
	router.GET("/", cache.CacheByRequestPath(store, cacheTime), home)
	router.GET("/home", cache.CacheByRequestPath(store, cacheTime), home)
	router.GET("/contact", cache.CacheByRequestPath(store, cacheTime), contact)
	router.GET("/KeyboardSoundPlayer", cache.CacheByRequestPath(store, cacheTime), kbs)
	router.GET("/robots.txt", cache.CacheByRequestPath(store, cacheTime), func(c *gin.Context) {
		c.File("static/robots.txt")
	})
	router.GET("/KeyboardSoundPlayer/vm_exe", cache.CacheByRequestPath(store, cacheTime), func(c *gin.Context) {
		c.File("static/voicemeeterprosetup.exe")
	})
	router.GET("/logo.png", cache.CacheByRequestPath(store, cacheTime), logo)
	router.GET("/favicon.ico", cache.CacheByRequestPath(store, cacheTime), favicon)
	router.GET("/ksp_logo.png", cache.CacheByRequestPath(store, cacheTime), kspLogo)
	router.GET("/domain_ownership_verification", func(c *gin.Context) {
		c.String(200, "This domain is Owned and Managed by JGL Technologies LLC. Email gluca@jgltechnologies for more info.")
	})

	testGroup := router.Group("/test")
	{
		testGroup.GET("/bmi", cache.CacheByRequestPath(store, cacheTime), test.BMIHome)
		testGroup.GET("/bmi/static/main.js", cache.CacheByRequestPath(store, cacheTime), test.BMIJS)
	}

	apiGroup := router.Group("/api")
	apiGroup.Use(AllowCors)
	{
		apiGroup.GET("/bot/status", cache.CacheByRequestPath(store, time.Second*5), api.BotStatus)
		apiGroup.GET("/jna", api.JNA)
		apiGroup.GET("/bot/info", cache.CacheByRequestPath(store, time.Second*5), api.BotInfo)
		apiGroup.POST("/traffic", cache.CacheByRequestPath(store, time.Second*5), api.CFProxy)
		apiGroup.POST("/contact", utils.GetMW(time.Second, 1), reqIDMiddleware, api.Contact)
		apiGroup.GET("/error", cache.CacheByRequestURI(store, cacheTime), api.GetErr)
	}

	router.NoRoute(noRoute)
	router.NoMethod(noMethod)
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
	err = session.Start("bash -c 'sudo pkill firefox-esr ; DISPLAY=:0 firefox-esr --kiosk /var/www/drive/jglnews.html & disown'")
	if err != nil {
		c.String(500, fmt.Sprintf("Error: %v", err))
		return
	}
	c.String(200, "Success")
}

func kbs(c *gin.Context) {
	c.HTML(200, "kbs", gin.H{})
}

func favicon(c *gin.Context) {
	c.File("static/favicon.ico")
}

func logo(c *gin.Context) {
	c.File("static/logo.png")
}

func kspLogo(c *gin.Context) {
	c.File("static/ksp_logo.png")
}

func home(c *gin.Context) {
	c.HTML(200, "home", gin.H{})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

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
	n := api.Announcement{c.GetHeader("Title"), c.GetHeader("Body"), time.Now().Unix(), time.Now().Unix() + int64(exp*86400)}
	announcements = append(announcements, n)
	f.WriteJSON(announcements)
	c.String(200, "Success")
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
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}
	if utils.StartsWith(c.Request.URL.String(), "/api") {
		c.JSON(405, gin.H{"error": "Method Not Allowed"})
	} else {
		c.HTML(405, "status", gin.H{
			"code":    "405",
			"message": "The method you used is not allowed.",
		})
	}
}
