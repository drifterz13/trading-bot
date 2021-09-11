package utils

import (
	"math"
)

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func GetPriceRatio(price float64, comparer float64) float64 {
	r := (100 / comparer) * price
	return r - 100
}
