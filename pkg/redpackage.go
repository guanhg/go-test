package pkg

import (
	"math"
	"math/rand"
)

const (
	minMoney = 0.01
)

// 红包分片，每次从范围 0.01~两倍平均值 之间随机取值
func Assign(total float64, size uint) []float64 {

	var ag []float64

	if total < minMoney*float64(size) || size < 1 {
		panic("太少了")
	}

	for ; size > 1; size-- {
		max := 2 * total / float64(size) // 平均值的两倍
		money := math.Floor(rand.Float64()*max*100) / 100
		if money < minMoney {
			money = minMoney
		}
		ag = append(ag, money)
		total -= money
	}
	total = math.Round(total*100) / 100
	ag = append(ag, total)
	return ag
}
