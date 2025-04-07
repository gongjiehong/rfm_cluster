package models

import (
	"math"
)

func Mean(v []float64) float64 {
	var res float64 = 0
	var n int = len(v)
	for i := range n {
		res += v[i]
	}
	return res / float64(n)
}

func Variance(v []float64) float64 {
	var res float64 = 0
	var m = Mean(v)
	var n int = len(v)
	for i := range n {
		res += (v[i] - m) * (v[i] - m)
	}
	return res / float64(n-1)
}

func Stddev(v []float64) float64 {
	return math.Sqrt(Variance(v))
}
