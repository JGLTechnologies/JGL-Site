package test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"strconv"
	"time"
)

func BMIHome(c *gin.Context) {
	lastBMI, err := c.Cookie("BMI_LAST")
	if err != nil {
		c.HTML(200, "bmi-home", gin.H{"last": "Not Found"})
	} else {
		c.HTML(200, "bmi-home", gin.H{"last": lastBMI})
	}
}

func BMICalc(c *gin.Context) {
	var context gin.H
	feet := c.Query("heightft")
	inches := c.Query("heightin")
	weight := c.Query("weight")

	if inches == "" {
		inches = "0"
	}

	feetNum, err := strconv.ParseFloat(feet, 64)
	if err != nil {
		c.HTML(400, "bmi-invalid", gin.H{})
	}

	inchesNum, err := strconv.ParseFloat(inches, 64)
	if err != nil {
		c.HTML(400, "bmi-invalid", gin.H{})
	}

	weightNum, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		c.HTML(400, "bmi-invalid", gin.H{})
	}

	bmi := weightNum / math.Pow((feetNum*12)+inchesNum, 2) * 703

	if bmi > 24.9 {
		newWeight := 24.9 / 703 * math.Pow((feetNum*12)+inchesNum, 2)
		pounds := fmt.Sprintf("%f", math.Round(weightNum-newWeight))
		poundsNum, _ := strconv.Atoi(pounds)
		if poundsNum >= 1 {
			context = gin.H{"bmi": math.Round(bmi), "weight": "You need to loose " + pounds + "pounds to be healthy."}
		} else {
			context = gin.H{"bmi": math.Round(bmi), "weight": ""}
		}
	} else if bmi < 18.5 {
		newWeight := 18.5 / 703 * math.Pow((feetNum*12)+inchesNum, 2)
		pounds := fmt.Sprintf("%f", math.Round(newWeight-weightNum))
		poundsNum, _ := strconv.Atoi(pounds)
		if poundsNum >= 1 {
			context = gin.H{"bmi": math.Round(bmi), "weight": "You need to gain " + pounds + "pounds to be healthy."}
		} else {
			context = gin.H{"bmi": math.Round(bmi), "weight": ""}
		}
	} else {
		context = gin.H{"bmi": math.Round(bmi), "weight": ""}
	}
	maxAge := time.Date(2038, 1, 1, 0, 0, 0, 0, time.Local).Unix() - time.Now().Unix()
	c.SetCookie("BMI_LAST", fmt.Sprintf("%f", bmi), int(maxAge), "/test/bmi", "jgltechnologies.com", false, false)
	c.HTML(200, "bmi-calc", context)
}
