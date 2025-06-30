package test

import (
	"github.com/gin-gonic/gin"
)

func BMIHome(c *gin.Context) {
	c.HTML(200, "bmi-home", gin.H{})
}

func BMIJS(c *gin.Context) {
	c.File("go web files/bmi/build/static/js/main.js")
}
