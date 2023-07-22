package test

import (
	"github.com/JGLTechnologies/SimpleFiles"
	"github.com/gin-gonic/gin"
)

func BMIHome(c *gin.Context) {
	c.HTML(200, "bmi-home", gin.H{})
}

func BMIJS(c *gin.Context) {
	f, _ := SimpleFiles.New("go web files/bmi/build/static/js/main.js")
	s, _ := f.ReadString()
	c.Header("Content-Type", "application/javascript")
	c.String(200, s)
}
