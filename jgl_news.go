package main

import (
	"JGLSite/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

func jna(c *gin.Context) {
	c.HTML(http.StatusOK, "jna", gin.H{})
}

func jnau(c *gin.Context) {
	exp, _ := strconv.Atoi(c.PostForm("expire"))
	now := time.Now().Unix()
	a := utils.Announcement{
		Title:  c.PostForm("title"),
		Body:   c.PostForm("body"),
		Time:   now,
		Expire: now + int64(exp*3600),
	}
	utils.DB.Create(&a)
	var emails []string
	utils.DB.Model(&utils.Email{}).Pluck("email", &emails)
	if len(emails) > 0 {
		if err := utils.SendEmail(emails, a.Title, a.Body); err != nil {
			log.Printf("JGL News email failed: %v", err)
			c.JSON(http.StatusFailedDependency, gin.H{
				"error":   "announcement saved, but the email could not be sent",
				"details": err.Error(),
			})
			return
		}
	}
	c.String(http.StatusOK, "Success")
}

func getJNAEmails(c *gin.Context) {
	var emails []string
	utils.DB.Model(&utils.Email{}).Pluck("email", &emails)
	c.JSON(http.StatusOK, emails)
}

func addJNAEmail(c *gin.Context) {
	address, err := mail.ParseAddress(strings.TrimSpace(c.PostForm("email")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enter a valid email address"})
		return
	}

	email := utils.Email{Email: address.Address}

	result := utils.DB.Create(&email)
	switch {
	case errors.Is(result.Error, gorm.ErrDuplicatedKey):
		c.JSON(http.StatusConflict, gin.H{
			"error": "email address already exists",
		})
		return

	case result.Error != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to save email address",
		})
		return
	}
	var emails []string
	utils.DB.Model(&utils.Email{}).Pluck("email", &emails)
	c.JSON(http.StatusCreated, emails)
}

func removeJNAEmail(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email address is required"})
		return
	}

	utils.DB.Where("email = ?", email).Delete(&utils.Email{})
	var emails []string
	utils.DB.Model(&utils.Email{}).Pluck("email", &emails)
	c.JSON(http.StatusOK, emails)
}

func jnu(c *gin.Context) {
	user := os.Getenv("sshuser")
	password := os.Getenv("sshpass")
	host := "jgltv"

	// Configure client
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
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
