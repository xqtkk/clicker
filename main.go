package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	mutex      sync.Mutex
	sseClients = make(map[chan string]bool)
	sseMutex   sync.Mutex

	score           int // Количество очков
	totalScore      int // Общее количество очков за все время
	totalSpentScore int // Общее количество потраченных очков за все время
	clicks          int // Общее количество кликов за все время
	playedTime      int // Общее время за сеанс

	autoClicker1           int          // Количество купленных автокликеров +1 score/s
	autoClicker1Price      int = 100    // Начальная цена автокликера +1 score/s
	autoClicker10          int          // Количество купленных автокликеров +10 score/s
	autoClicker10Price     int = 800    // Начальная цена автокликера +10 score/s
	autoClicker120         int          // Количество купленных автокликеров +120 score/s
	autoClicker120Price    int = 10000  // Начальная цена автокликера +120 score/s
	autoClicker1000        int          // Количество купленных автокликеров +1000 score/s
	autoClicker1000Price   int = 75000  // Начальная цена автокликера +1000 score/s
	autoClicker5000        int          // Количество купленных автокликеров +5000 score/s
	autoClicker5000Price   int = 250000 // Начальная цена автокликера +5000 score/s
	autoClicks             int          // Общее количетсво автокликов в секунду
	clickPower             int = 1      // Начальная сила клика
	clickPowerUpgradePrice int = 20     // Начальная цена прокачки силы клика
	clickPowerUpgrades     int
	// Достижения
	achievements = map[string]bool{
		"First click!": false, // Достижение за первый клик
	}
)

func main() {
	r := gin.Default()

	// Получение текущего состояния
	r.GET("/score", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"score":                  score,
			"autoClicker1":           autoClicker1,
			"autoClicker1Price":      autoClicker1Price,
			"clickPowerUpgradePrice": clickPowerUpgradePrice,
			"clickPowerUpgrades":     clickPowerUpgrades,
			"achievements":           achievements,
			"totalScore":             totalScore,
			"totalSpentScore":        totalSpentScore,
			"clicks":                 clicks,
			"autoClicker10":          autoClicker10,
			"autoClicker10Price":     autoClicker10Price,
			"autoClicker120":         autoClicker120,
			"autoClicker120Price":    autoClicker120Price,
			"autoClicker1000":        autoClicker1000,
			"autoClicker1000Price":   autoClicker1000Price,
			"autoClicker5000":        autoClicker5000,
			"autoClicker5000Price":   autoClicker5000Price,
			"autoclicks":             autoClicks,
			"playedTime":             playedTime,
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
		totalScore += clickPower
		clicks++

		if totalScore == 1 {
			go startPlayedTime()
		}

		checkAchievements()
		broadcastScore(critical)
		c.JSON(http.StatusOK, gin.H{"score": score, "critical": critical, "clickPower": clickPower, "clickPowerUpgrades": clickPowerUpgrades, "autoclicks": autoClicks})
	})

	// Покупка автокликера 1 score/s
	r.POST("/buy-autoclicker1", func(c *gin.Context) {
		mutex.Lock()

		if score >= autoClicker1Price {
			score -= autoClicker1Price
			totalSpentScore += autoClicker1Price
			autoClicker1Price = increaseAutoClickerPrice(autoClicker1Price)
			autoClicker1++

			if autoClicker1 == 1 {
				go startAutoClickers()
			}
		}

		mutex.Unlock()

		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":             score,
			"autoClicker1":      autoClicker1,
			"autoClicker1Price": autoClicker1Price,
			"autoclicks":        autoClicks,
		})
	})

	// Покупка автокликера 10 score/s
	r.POST("/buy-autoclicker10", func(c *gin.Context) {
		mutex.Lock()

		if score >= autoClicker10Price {
			score -= autoClicker10Price
			totalSpentScore += autoClicker10Price
			autoClicker10Price = increaseAutoClickerPrice(autoClicker10Price)
			autoClicker10++
		}

		mutex.Unlock()

		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":              score,
			"autoClicker10":      autoClicker10,
			"autoClicker10Price": autoClicker10Price,
			"autoclicks":         autoClicks,
		})
	})

	// Покупка автокликера 120 score/s
	r.POST("/buy-autoclicker120", func(c *gin.Context) {
		mutex.Lock()

		if score >= autoClicker120Price {
			score -= autoClicker120Price
			totalSpentScore += autoClicker120Price
			autoClicker120Price = increaseAutoClickerPrice(autoClicker120Price)
			autoClicker120++
		}

		mutex.Unlock()

		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":               score,
			"autoClicker120":      autoClicker120,
			"autoClicker120Price": autoClicker120Price,
			"autoclicks":          autoClicks,
		})
	})

	// Покупка автокликера +1000 score/s
	r.POST("/buy-autoclicker1000", func(c *gin.Context) {
		mutex.Lock()

		if score >= autoClicker1000Price {
			score -= autoClicker1000Price
			totalSpentScore += autoClicker1000Price
			autoClicker1000Price = increaseAutoClickerPrice(autoClicker1000Price)
			autoClicker1000++
		}

		mutex.Unlock()

		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":                score,
			"autoClicker1000":      autoClicker1000,
			"autoClicker1000Price": autoClicker1000Price,
			"autoclicks":           autoClicks,
		})
	})

	// Покупка автокликера 5000 score/s
	r.POST("/buy-autoclicker5000", func(c *gin.Context) {
		mutex.Lock()

		if score >= autoClicker5000Price {
			score -= autoClicker5000Price
			totalSpentScore += autoClicker5000Price
			autoClicker5000Price = increaseAutoClickerPrice(autoClicker5000Price)
			autoClicker5000++
		}

		mutex.Unlock()

		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":                score,
			"autoClicker5000":      autoClicker5000,
			"autoClicker5000Price": autoClicker5000Price,
			"autoclicks":           autoClicks,
		})
	})

	// Покупка улучшении силы клика
	r.POST("/buy-click-power-upgrade", func(c *gin.Context) {
		mutex.Lock()
		if score >= clickPowerUpgradePrice {
			score -= clickPowerUpgradePrice
			totalSpentScore += clickPowerUpgradePrice
			clickPower *= 2
			clickPowerUpgradePrice *= 10
			clickPowerUpgrades++
		}

		mutex.Unlock()
		broadcastScore(false)
		c.JSON(http.StatusOK, gin.H{
			"score":              score,
			"clickPowerUpgrade":  clickPowerUpgradePrice,
			"clickPowerUpgrades": clickPowerUpgrades,
			"autoclicks":         autoClicks,
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

	fmt.Println("Server running on http://localhost:4000/static")
	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatal(err)
	}
}

// Запуск автокликера 1 score/s
func startAutoClickers() {
	for autoClicks >= 0 {
		time.Sleep(1 * time.Second)
		mutex.Lock()

		autoClicks = autoClicker1 + autoClicker10*10 + autoClicker1000*1000 + autoClicker120*120 + autoClicker5000*5000
		score += autoClicks // Один клик за каждый автокликер
		totalScore += autoClicks
		mutex.Unlock()
		broadcastScore(false)
	}
}

func startPlayedTime() {
	for autoClicks >= 0 {
		time.Sleep(1 * time.Second)
		mutex.Lock()

		playedTime++

		mutex.Unlock()
		broadcastScore(false)
	}
}

func increaseAutoClickerPrice(before int) int {
	return before + int(math.Round((float64(before) / 6.66666667)))
}

// Отправка обновлений клиентам
func broadcastScore(critical bool) {
	sseMutex.Lock()
	defer sseMutex.Unlock()

	// Создаём структуру с данными
	data := map[string]interface{}{
		"score":                  score,
		"autoClicker1":           autoClicker1,
		"clickPowerUpgrades":     clickPowerUpgrades,
		"autoClicker1Price":      autoClicker1Price,
		"clickPowerUpgradePrice": clickPowerUpgradePrice,
		"achievements":           achievements,
		"critical":               critical,
		"autoClicker10":          autoClicker10,
		"autoClicker10Price":     autoClicker10Price,
		"autoClicker120":         autoClicker120,
		"autoClicker120Price":    autoClicker120Price,
		"autoClicker1000":        autoClicker1000,
		"autoClicker1000Price":   autoClicker1000Price,
		"autoClicker5000":        autoClicker5000,
		"autoClicker5000Price":   autoClicker5000Price,
		"autoClicks":             autoClicks,
		"playedTime":             playedTime,
		"totalScore":             totalScore,
		"totalSpentScore":        totalSpentScore,
		"clicks":                 clicks,
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
	if score >= 1 && !achievements["First click!"] {
		achievements["First click!"] = true
	}
}
