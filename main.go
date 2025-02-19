package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	score        int
	progress     int
	autoClickers int32 // Количество купленных автокликеров
	mutex        sync.Mutex
	sseClients   = make(map[chan string]bool)
	sseMutex     sync.Mutex
)

func main() {
	r := gin.Default()

	// Получение текущего состояния
	r.GET("/score", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		c.JSON(http.StatusOK, gin.H{
			"score":        score,
			"progress":     progress,
			"autoClickers": atomic.LoadInt32(&autoClickers),
			"price":        getAutoClickerPrice(),
		})
	})

	// Клик по кнопке
	r.POST("/click", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		score++
		progress = score / 10
		broadcastScore()
		c.JSON(http.StatusOK, gin.H{"score": score, "progress": progress})
	})

	// Покупка автокликера
	r.POST("/buy-autoclicker", func(c *gin.Context) {
		mutex.Lock()
		price := getAutoClickerPrice()
		if score >= price {
			score -= price
			atomic.AddInt32(&autoClickers, 1)
			if autoClickers == 1 {
				go startAutoClicker()
			}
		}
		mutex.Unlock()
		broadcastScore()
		c.JSON(http.StatusOK, gin.H{
			"score":        score,
			"progress":     progress,
			"autoClickers": atomic.LoadInt32(&autoClickers),
			"price":        getAutoClickerPrice(),
		})
	})

	// SSE для обновлений
	r.GET("/events", func(c *gin.Context) {
		clientChan := make(chan string)
		sseMutex.Lock()
		sseClients[clientChan] = true
		sseMutex.Unlock()

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})

		sseMutex.Lock()
		delete(sseClients, clientChan)
		close(clientChan)
		sseMutex.Unlock()
	})

	r.Static("/static", "./static")

	fmt.Println("Server running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

// Возвращает текущую цену автокликера
func getAutoClickerPrice() int {
	n := atomic.LoadInt32(&autoClickers)
	return 50 * (1 << n) // 50, 100, 200, 400, 800...
}

// Запуск автокликера
func startAutoClicker() {
	for atomic.LoadInt32(&autoClickers) > 0 {
		time.Sleep(1 * time.Second)
		mutex.Lock()
		score += int(atomic.LoadInt32(&autoClickers)) // Один клик за каждый автокликер
		println(int(atomic.LoadInt32(&autoClickers)))
		progress = score / 10
		mutex.Unlock()
		broadcastScore()
	}
}

// Отправка обновлений клиентам
func broadcastScore() {
	sseMutex.Lock()
	defer sseMutex.Unlock()
	for clientChan := range sseClients {
		clientChan <- fmt.Sprintf(`{"score": %d, "progress": %d, "autoClickers": %d, "price": %d}`, score, progress, atomic.LoadInt32(&autoClickers), getAutoClickerPrice())
	}
}
