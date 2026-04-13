package test

import (
	"github.com/gin-gonic/gin"
)

const bmiJSPath = "go web files/bmi/build/static/js/main.js"

func BMIHome(c *gin.Context) {
	c.HTML(200, "bmi-home", gin.H{})
}

func BMIJS(c *gin.Context) {
	c.File(bmiJSPath)
}
