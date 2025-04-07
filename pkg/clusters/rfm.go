package clusters

import (
	"math"
)

type RFM struct {
	UserID uint64
	R      float64
	F      float64
	M      float64
}

func (c RFM) Coordinates() Coordinates {
	return []float64{float64(c.R), float64(c.F), float64(c.M)}
}

func (c RFM) Distance(p2 Coordinates) float64 {
	var r float64
	for i, v := range c.Coordinates() {
		r += math.Pow(v-p2.Coordinates()[i], 2)
	}
	return r
}
