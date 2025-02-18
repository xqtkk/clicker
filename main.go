package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	score       int
	progress    int
	autoClicker bool
	mutex       sync.Mutex
)

func main() {
	r := gin.Default()

	r.GET("/score", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		c.JSON(http.StatusOK, gin.H{"score": score, "progress": progress, "autoClicker": autoClicker})
	})

	r.POST("/click", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		score++
		progress = score / 10 // Example: Progress increases every 10 clicks
		c.JSON(http.StatusOK, gin.H{"score": score, "progress": progress})
	})

	r.POST("/buy-autoclicker", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		if score >= 50 && !autoClicker { // Auto-clicker costs 50 points
			score -= 50
			autoClicker = true
			go startAutoClicker()
		}
		c.JSON(http.StatusOK, gin.H{"score": score, "progress": progress, "autoClicker": autoClicker})
	})

	r.Static("/static", "./static")

	fmt.Println("Server running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func startAutoClicker() {
	for autoClicker {
		time.Sleep(1 * time.Second)
		mutex.Lock()
		score++
		progress = score / 10
		mutex.Unlock()
	}
}
