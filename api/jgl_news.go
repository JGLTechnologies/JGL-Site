package api

import (
	"JGLSite/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

func JNA(c *gin.Context) {
	var a []utils.Announcement
	utils.DB.Model(&utils.Announcement{}).Where("expire > ?", time.Now().Unix()).Order("ID DESC").Find(&a)
	c.JSON(http.StatusOK, a)
}

func JNU(c *gin.Context) {
	user := os.Getenv("sshuser")
	password := os.Getenv("sshpass")
	host := "jgltv:22"

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
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer client.Close()

	// Create a new session
	session, err := client.NewSession()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	// Run a command on the remote host
	err = session.Start("bash -c 'export DISPLAY=:0; export XAUTHORITY=/home/pi/.Xauthority; sudo pkill firefox-esr; sudo xhost +; sudo unclutter -display :0 -idle 0 -root & firefox-esr --kiosk /var/www/drive/jglnews.html &'")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.String(200, "Success")
}
