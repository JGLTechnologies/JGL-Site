package main

import (
	"JGLSite/api"
	"JGLSite/utils"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func jna(c *gin.Context) {
	c.HTML(http.StatusOK, "jna", gin.H{})
}

func jnau(c *gin.Context) {
	data, err := api.ReadJNAData()
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to read announcement data")
		return
	}
	exp, _ := strconv.Atoi(c.PostForm("expire"))
	n := api.Announcement{c.PostForm("title"), c.PostForm("body"), time.Now().Unix(), time.Now().Unix() + int64(exp*3600)}
	data.Announcements = append(data.Announcements, n)
	if err := api.WriteJNAData(data); err != nil {
		c.String(http.StatusInternalServerError, "Unable to save announcement")
		return
	}
	if len(data.Emails) > 0 {
		if err := utils.SendEmail(data.Emails, n.Title, n.Body); err != nil {
			c.String(http.StatusBadGateway, "Announcement saved, but the email could not be sent: %v", err)
			return
		}
	}
	c.String(http.StatusOK, "Success")
}

func getJNAEmails(c *gin.Context) {
	data, err := api.ReadJNAData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to read email list"})
		return
	}
	c.JSON(http.StatusOK, data.Emails)
}

func addJNAEmail(c *gin.Context) {
	address, err := mail.ParseAddress(strings.TrimSpace(c.PostForm("email")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enter a valid email address"})
		return
	}

	data, err := api.ReadJNAData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to read email list"})
		return
	}
	for _, existing := range data.Emails {
		if strings.EqualFold(existing, address.Address) {
			c.JSON(http.StatusConflict, gin.H{"error": "email address is already in the list"})
			return
		}
	}
	data.Emails = append(data.Emails, address.Address)
	if err := api.WriteJNAData(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save email address"})
		return
	}
	c.JSON(http.StatusCreated, data.Emails)
}

func removeJNAEmail(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email address is required"})
		return
	}

	data, err := api.ReadJNAData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to read email list"})
		return
	}

	emails := make([]string, 0, len(data.Emails))
	found := false
	for _, existing := range data.Emails {
		if strings.EqualFold(existing, email) {
			found = true
			continue
		}
		emails = append(emails, existing)
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "email address was not found"})
		return
	}

	data.Emails = emails
	if err := api.WriteJNAData(data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save email list"})
		return
	}
	c.JSON(http.StatusOK, data.Emails)
}
