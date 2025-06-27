package main

import (
	"JGLSite/api"
	"JGLSite/test"
	"JGLSite/utils"
	"context"
	"errors"
	"github.com/JGLTechnologies/SimpleFiles"
	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/jucardi/go-streams/v2/streams"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var store *persist.MemoryStore
var Projects []*api.Project

const port string = ":81"

func main() {
	godotenv.Load("/var/www/.env")
	defer utils.GetDB().Close()
	store = persist.NewMemoryStore(time.Hour)

	// Create HTML templates
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
	r.AddFromFiles("kbs", "go web files/kbs.html", "go web files/base.html")

	// Router config
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	router.HTMLRender = r
	router.Use(gin.Logger())
	router.HandleMethodNotAllowed = true
	f, _ := SimpleFiles.New("cloudflare_ips.txt")
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
	router.GET("/", cache.CacheByRequestPath(store, time.Minute*10), home)
	router.GET("/home", cache.CacheByRequestPath(store, time.Hour*24), home)
	router.GET("/projects", cache.CacheByRequestPath(store, time.Minute), projects)
	router.GET("/contact", cache.CacheByRequestPath(store, time.Hour*24), contact)
	router.GET("/KeyboardSoundPlayer", cache.CacheByRequestPath(store, time.Hour*24), kbs)
	router.GET("/logo.png", cache.CacheByRequestPath(store, time.Hour*24), logo)
	router.GET("/favicon.ico", cache.CacheByRequestPath(store, time.Hour*24), favicon)
	router.GET("/ksp_logo.png", cache.CacheByRequestPath(store, time.Hour*24), kspLogo)

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
		apiGroup.POST("/contact", utils.GetMW(time.Second, 1), reqIDMiddleware, api.Contact)
		//apiGroup.POST("/custom-bot", utils.GetMW(time.Second, 1), reqIDMiddleware, api.CustomBot)
		apiGroup.GET("/error", cache.CacheByRequestURI(store, time.Hour*24), api.GetErr)
	}

	router.NoRoute(noRoute)
	router.NoMethod(noMethod)
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}
	// Load projects
	p, _ := api.Projects()
	Projects = p
	go func() {
		for {
			time.Sleep(time.Minute * 10)
			p, _ := api.Projects()
			Projects = p
		}
	}()
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

func projects(c *gin.Context) {
	data := gin.H{}
	var html template.HTML
	if len(Projects) < 1 {
		data["projects"] = template.HTML("<p>Projects could not be loaded.</p>")
	} else {
		f := func(project *api.Project) {
			project.Description = strings.ReplaceAll(project.Description, "[project]", "")
			html += template.HTML("<p class=\"lead fw-normal text-muted mb-0\">\n<br/>\n<span style='color: var(--bs-dark);'>" + project.Name + ":</span>\n<br/><span style=\"position: relative; left: 10px;\">Description: " + project.Description + "</span>\n<br/><span style='position: relative; left: 10px; top: 7px;'>Github URL: <a\nhref=https://github.com/JGLTechnologies/" + project.Name + " >click</a></span>\n</p>")
		}
		stream := streams.From[*api.Project](Projects).Filter(
			func(p *api.Project) bool {
				return strings.Contains(p.Description, "[project]") && !p.Private && p.Name != "JGL-Site"
			})
		if stream.Count() < 1 {
			data["projects"] = template.HTML("<p>There are no current projects.</p>")
		} else {
			stream.ForEach(f)
			data["projects"] = html
		}
	}
	c.HTML(200, "projects", data)
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
			"message": "The method you used is not allowed.",
		})
	}
}
