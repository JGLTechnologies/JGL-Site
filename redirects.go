package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerRedirectRoutes(router *gin.Engine) {
	router.GET("/keyboardsoundplayeryoutube", permanentRedirect("https://www.youtube.com/watch?v=lxf4MtiYwRY"))
	router.GET("/keyboardsoundplayerstore", permanentRedirect("https://apps.microsoft.com/detail/9pfsjgvshm0l?hl=en-US&gl=US"))
	router.GET("/discord", permanentRedirect("https://discord.gg/TUUbzTa3B7"))
	router.GET("/dpys", permanentRedirect("https://github.com/JGLTechnologies/DPYS/blob/main/README.md"))
	router.GET("/dpys/src", permanentRedirect("https://github.com/JGLTechnologies/dpys"))
	router.GET("/dpys/pypi", permanentRedirect("https://pypi.org/project/dpys"))
	router.GET("/aiohttplimiter", permanentRedirect("https://github.com/JGLTechnologies/aiohttp-ratelimiter"))
	router.GET("/aiohttplimiter/pypi", permanentRedirect("https://pypi.org/project/aiohttp-ratelimiter"))
	router.GET("/precise-memory-rate-limit", permanentRedirect("https://www.npmjs.com/package/precise-memory-rate-limit"))
	router.GET("/precise-memory-rate-limit/src", permanentRedirect("https://github.com/JGLTechnologies/precise-memory-rate-limit"))
	router.GET("/gin-rate-limit", permanentRedirect("https://github.com/JGLTechnologies/gin-rate-limit"))
	router.GET("/src", permanentRedirect("https://github.com/JGLTechnologies/jgl-site"))
	router.GET("/gh", permanentRedirect("https://github.com/JGLTechnologies"))
	router.GET("/bot", permanentRedirect("/jgl-bot"))
	router.GET("/jglbot", permanentRedirect("/jgl-bot"))
	router.GET("/bot/invite", permanentRedirect("https://discord.com/api/oauth2/authorize?client_id=844976951692361800&permissions=8&scope=bot%20applications.commands"))
	router.GET("/bot/top", permanentRedirect("https://top.gg/bot/844976951692361800"))
}

func permanentRedirect(location string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, location)
	}
}
