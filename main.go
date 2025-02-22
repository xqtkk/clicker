package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	score             int
	progress          int
	autoClickers      int32 // Количество купленных автокликеров
	clickPower        int   = 1
	clickUpgradePrice int   = 200
	mutex             sync.Mutex
	sseClients        = make(map[chan string]bool)
	sseMutex          sync.Mutex

	achievements = map[string]bool{
		"Первый клик!":                         false,
		"🏅 Новичок → Набрать 100 очков.":       false,
		"🔥 Клик-мастер → Набрать 1000 очков.":  false,
		"👑 Легенда → Набрать 1 000 000 очков.": false,
	}
)

func main() {
	r := gin.Default()

	// Получение текущего состояния
	r.GET("/score", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		c.JSON(http.StatusOK, gin.H{
			"score":             score,
			"progress":          progress,
			"autoClickers":      atomic.LoadInt32(&autoClickers),
			"autoClickerPrice":  getAutoClickerPrice(),
			"clickUpgradePrice": clickUpgradePrice,
			"achievements":      achievements,
		})
	})

	// Клик по кнопке
	r.POST("/click", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		criticalClickPower := clickPower
		critical := false
		if rand.Float32() < 0.1 {
			criticalClickPower *= 2
			critical = true
		}

		score += clickPower

		checkAchievements()
		broadcastScore(critical)
		c.JSON(http.StatusOK, gin.H{"score": score, "critical": critical, "clickPower": clickPower})
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
		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":        score,
			"autoClickers": atomic.LoadInt32(&autoClickers),
			"price":        getAutoClickerPrice(),
		})
	})

	r.POST("/buy-click-upgrade", func(c *gin.Context) {
		mutex.Lock()
		if score >= clickUpgradePrice {
			score -= clickUpgradePrice
			clickPower *= 2
			clickUpgradePrice *= 4
		}

		mutex.Unlock()
		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":        score,
			"clickPower":   clickPower,
			"clickUpgrade": clickUpgradePrice,
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
		close(clientChan)
		delete(sseClients, clientChan)
		sseMutex.Unlock()
	})

	r.Static("/static", "./static")

	fmt.Println("Server running on http://localhost:4000")
	if err := r.Run(":4000"); err != nil {
		log.Fatal(err)
	}
}

// Возвращает текущую цену автокликера
func getAutoClickerPrice() int {
	n := atomic.LoadInt32(&autoClickers)
	return 25 * (1 << n) // 50, 100, 200, 400, 800...
}

// Запуск автокликера
func startAutoClicker() {
	for atomic.LoadInt32(&autoClickers) > 0 {
		time.Sleep(1 * time.Second)
		mutex.Lock()
		score += int(atomic.LoadInt32(&autoClickers)) // Один клик за каждый автокликер
		progress = score / 10
		mutex.Unlock()
		broadcastScore(false)
	}
}

// Отправка обновлений клиентам
func broadcastScore(critical bool) {
	sseMutex.Lock()
	defer sseMutex.Unlock()

	// Создаём структуру с данными
	data := map[string]interface{}{
		"score":             score,
		"progress":          progress,
		"autoClickers":      atomic.LoadInt32(&autoClickers),
		"clickPower":        clickPower,
		"autoClickerPrice":  getAutoClickerPrice(),
		"clickUpgradePrice": clickUpgradePrice,
		"achievements":      achievements,
		"critical":          critical,
	}

	// Преобразуем в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Ошибка кодирования JSON:", err)
		return
	}

	// fmt.Println("Отправка JSON:", string(jsonData)) // Проверяем что уходит клиентам

	for clientChan := range sseClients {
		clientChan <- string(jsonData)
	}
}

func checkAchievements() {
	if score >= 1 && !achievements["Первый клик!"] {
		achievements["Первый клик!"] = true
	}
	if score >= 100 && !achievements["🏅 Новичок → Набрать 100 очков."] {
		achievements["🏅 Новичок → Набрать 100 очков."] = true
	}
	if score >= 1000 && !achievements["🔥 Клик-мастер → Набрать 1000 очков."] {
		achievements["🔥 Клик-мастер → Набрать 1000 очков."] = true
	}
	if score >= 1000000 && !achievements["👑 Легенда → Набрать 1 000 000 очков."] {
		achievements["👑 Легенда → Набрать 1 000 000 очков."] = true
	}
}
