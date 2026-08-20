package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func requireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("pass") == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "login is not configured"})
			return
		}

		if isLoggedIn(c) {
			c.Next()
			return
		}

		next := c.Request.URL.RequestURI()
		c.Redirect(http.StatusSeeOther, "/login?next="+urlQueryEscape(next))
		c.Abort()
	}
}

// APILogin rejects unauthenticated API requests without redirecting to an HTML page.
func apiLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("pass") == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "login is not configured"})
			return
		}
		if !isLoggedIn(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.Next()
	}
}

func isLoggedIn(c *gin.Context) bool {
	secret := os.Getenv("pass")
	if secret == "" {
		return false
	}
	cookie, err := c.Cookie(loginCookieName)
	return err == nil && hmac.Equal([]byte(cookie), []byte(loginCookieValue(secret)))
}

func showLogin(c *gin.Context) {
	renderLogin(c, http.StatusOK, "")
}

func verifyLogin(c *gin.Context) {
	next := safeNextPath(c.PostForm("next"))
	secret := os.Getenv("pass")
	if secret == "" {
		renderLogin(c, http.StatusServiceUnavailable, "Login is not configured on the server.")
		return
	}

	fileHeader, err := c.FormFile("key")
	if err != nil {
		renderLogin(c, http.StatusBadRequest, "Choose a .txt key file.")
		return
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".txt") {
		renderLogin(c, http.StatusBadRequest, "The key must be a .txt file.")
		return
	}
	if fileHeader.Size > maxLoginKeyFileSize {
		renderLogin(c, http.StatusRequestEntityTooLarge, "The key file is too large.")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		renderLogin(c, http.StatusBadRequest, "The key file could not be read.")
		return
	}
	defer file.Close()

	key, err := io.ReadAll(io.LimitReader(file, maxLoginKeyFileSize+1))
	if err != nil || len(key) > maxLoginKeyFileSize {
		renderLogin(c, http.StatusBadRequest, "The key file could not be read.")
		return
	}
	if subtle.ConstantTimeCompare(key, []byte(secret)) != 1 {
		renderLogin(c, http.StatusUnauthorized, "The key file is not valid.")
		return
	}

	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(loginCookieName, loginCookieValue(secret), int(loginCookieMaxAge.Seconds()), "/", "", secure, true)
	c.Redirect(http.StatusSeeOther, next)
}

func renderLogin(c *gin.Context, status int, message string) {
	next := c.Query("next")
	if next == "" {
		next = c.PostForm("next")
	}
	c.HTML(status, "login", gin.H{
		"error": message,
		"next":  safeNextPath(next),
	})
}

func loginCookieValue(secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("JGL-Site login verification"))
	return hex.EncodeToString(mac.Sum(nil))
}

func safeNextPath(next string) string {
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return "/"
	}
	return next
}

func urlQueryEscape(value string) string {
	// QueryEscape's only purpose here is to preserve a protected URL and its query.
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
