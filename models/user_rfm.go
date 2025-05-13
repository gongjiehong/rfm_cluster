package models

import (
	"fmt"
	"math"
	"rfm_cluster/pkg/clusters"
	"rfm_cluster/pkg/kmeans"
	"rfm_cluster/pkg/silhouette"
	"slices"
)

type UserRFM struct {
	UserID            uint64  `json:"user_id"`
	Nickname          string  `json:"nickname"`
	Birthday          string  `json:"birthday"`
	Gender            int8    `json:"gender"`
	RecencyOriginal   float64 `json:"recency_original"`
	FrequencyOriginal float64 `json:"frequency_original"`
	MonetaryOriginal  float64 `json:"monetary_original"`
	RecencyWeighted   float64 `json:"recency_weighted"`
	FrequencyWeighted float64 `json:"frequency_weighted"`
	MonetaryWeighted  float64 `json:"monetary_weighted"`
}

type DataIndicators struct {
	Mean     float64 `json:"mean"`
	Stddev   float64 `json:"stddev"`
	Variance float64 `json:"variance"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
}

type RMFDataIndicators struct {
	R DataIndicators `json:"r"`
	F DataIndicators `json:"f"`
	M DataIndicators `json:"m"`
}

func ProcessRMFDataIndicators(dataCollection []*UserRFM) RMFDataIndicators {
	rArray := []float64{}
	fArray := []float64{}
	mArray := []float64{}

	for _, data := range dataCollection {
		rArray = append(rArray, data.RecencyOriginal)
		fArray = append(fArray, data.FrequencyOriginal)
		mArray = append(mArray, data.MonetaryOriginal)
	}

	result := RMFDataIndicators{
		R: DataIndicators{Mean: Mean(rArray), Stddev: Stddev(rArray), Variance: Variance(rArray), Min: slices.Min(rArray), Max: slices.Max(rArray)},
		F: DataIndicators{Mean: Mean(fArray), Stddev: Stddev(fArray), Variance: Variance(fArray), Min: slices.Min(fArray), Max: slices.Max(fArray)},
		M: DataIndicators{Mean: Mean(mArray), Stddev: Stddev(mArray), Variance: Variance(mArray), Min: slices.Min(mArray), Max: slices.Max(mArray)},
	}

	return result
}

func ProcessData(dataCollection []*UserRFM) (clusters.Observations, []silhouette.KScore, int, float64, error) {
	var observations clusters.Observations = processRealRFMData(dataCollection)

	// 构建kmeans
	km, err := kmeans.NewWithOptions(0.01, nil)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	// 计算kmeans的得分和分组
	scores, estimate, score, err := silhouette.EstimateK(observations, 8, km)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return observations, scores, estimate, score, nil
}

func processRealRFMData(dataCollection []*UserRFM) clusters.Observations {
	var d clusters.Observations

	// barrier := len(dataCollection) / 5

	// rsorted := make([]*UserRFM, len(dataCollection))
	// fsorted := make([]*UserRFM, len(dataCollection))
	// msorted := make([]*UserRFM, len(dataCollection))

	// copy(rsorted, dataCollection)
	// copy(fsorted, dataCollection)
	// copy(msorted, dataCollection)
	// slices.SortFunc(rsorted, func(a, b *UserRFM) int {
	// 	if a.RecencyOriginal < b.RecencyOriginal {
	// 		return -1
	// 	} else if a.RecencyOriginal > b.RecencyOriginal {
	// 		return 1
	// 	} else {
	// 		return 0
	// 	}
	// })

	// slices.SortFunc(fsorted, func(a, b *UserRFM) int {
	// 	if a.FrequencyOriginal < b.FrequencyOriginal {
	// 		return -1
	// 	} else if a.FrequencyOriginal > b.FrequencyOriginal {
	// 		return 1
	// 	} else {
	// 		return 0
	// 	}
	// })

	// slices.SortFunc(msorted, func(a, b *UserRFM) int {
	// 	if a.MonetaryOriginal < b.MonetaryOriginal {
	// 		return -1
	// 	} else if a.MonetaryOriginal > b.MonetaryOriginal {
	// 		return 1
	// 	} else {
	// 		return 0
	// 	}
	// })

	// rbarrier := make([]float64, 4)
	// fbarrier := make([]float64, 4)
	// mbarrier := make([]float64, 4)
	// for i := 1; i < 5; i++ {
	// 	rbarrier[i-1] = rsorted[i*barrier].RecencyOriginal
	// 	fbarrier[i-1] = fsorted[i*barrier].FrequencyOriginal
	// 	mbarrier[i-1] = msorted[i*barrier].MonetaryOriginal
	// }

	// fmt.Println(rbarrier)
	// fmt.Println(fbarrier)
	// fmt.Println(mbarrier)

	// for _, row := range dataCollection {
	// 	var r, f, m float64 = 0.0, 0.0, 0.0
	// 	if row.RecencyOriginal <= rbarrier[0] {
	// 		r = 5
	// 	} else if row.RecencyOriginal > rbarrier[0] && row.RecencyOriginal <= rbarrier[1] {
	// 		r = 4
	// 	} else if row.RecencyOriginal > rbarrier[1] && row.RecencyOriginal <= rbarrier[2] {
	// 		r = 3
	// 	} else if row.RecencyOriginal > rbarrier[2] && row.RecencyOriginal <= rbarrier[3] {
	// 		r = 2
	// 	} else if row.RecencyOriginal > rbarrier[3] {
	// 		r = 1
	// 	} else {
	// 		r = 1
	// 	}

	// 	if row.FrequencyOriginal <= fbarrier[0] {
	// 		f = 1
	// 	} else if row.FrequencyOriginal > fbarrier[0] && row.FrequencyOriginal <= fbarrier[1] {
	// 		f = 2
	// 	} else if row.FrequencyOriginal > fbarrier[1] && row.FrequencyOriginal <= fbarrier[2] {
	// 		f = 3
	// 	} else if row.FrequencyOriginal > fbarrier[2] && row.FrequencyOriginal <= fbarrier[3] {
	// 		f = 4
	// 	} else if row.FrequencyOriginal > fbarrier[3] {
	// 		f = 5
	// 	} else {
	// 		f = 5
	// 	}

	// 	if row.MonetaryOriginal <= mbarrier[0] {
	// 		m = 1
	// 	} else if row.MonetaryOriginal > mbarrier[0] && row.MonetaryOriginal <= mbarrier[1] {
	// 		m = 2
	// 	} else if row.MonetaryOriginal > mbarrier[1] && row.MonetaryOriginal <= mbarrier[2] {
	// 		m = 3
	// 	} else if row.MonetaryOriginal > mbarrier[2] && row.MonetaryOriginal <= mbarrier[3] {
	// 		m = 4
	// 	} else if row.MonetaryOriginal > mbarrier[3] {
	// 		m = 5
	// 	} else {
	// 		m = 5
	// 	}

	for _, row := range dataCollection {
		var r, f, m float64 = 0.0, 0.0, 0.0
		if row.RecencyOriginal == -1 {
			r = 0
		} else if row.RecencyOriginal <= 1 {
			r = 5
		} else if row.RecencyOriginal > 1 && row.RecencyOriginal <= 7 {
			r = 4
		} else if row.RecencyOriginal > 7 && row.RecencyOriginal <= 31 {
			r = 3
		} else if row.RecencyOriginal > 31 && row.RecencyOriginal <= 93 {
			r = 2
		} else if row.RecencyOriginal > 93 {
			r = 1
		}

		if row.FrequencyOriginal == 1 {
			f = 1
		} else if row.FrequencyOriginal >= 2 && row.FrequencyOriginal <= 4 {
			f = 2
		} else if row.FrequencyOriginal >= 5 && row.FrequencyOriginal <= 7 {
			f = 3
		} else if row.FrequencyOriginal >= 8 && row.FrequencyOriginal <= 10 {
			f = 4
		} else if row.FrequencyOriginal >= 11 {
			f = 5
		}

		if row.MonetaryOriginal <= 6.99 {
			m = 1
		} else if row.MonetaryOriginal > 6.99 && row.MonetaryOriginal <= 14.99 {
			m = 2
		} else if row.MonetaryOriginal > 14.99 && row.MonetaryOriginal <= 39.99 {
			m = 3
		} else if row.MonetaryOriginal > 39.99 && row.MonetaryOriginal <= 71.88 {
			m = 4
		} else if row.MonetaryOriginal > 71.88 {
			m = 5
		}

		row.RecencyWeighted = r
		row.FrequencyWeighted = f
		row.MonetaryWeighted = m

		d = append(d, row)
	}

	fmt.Printf("%d data points\n", len(d))

	return d
}

// clusters.Observation 协议实现
func (c *UserRFM) Coordinates() clusters.Coordinates {
	return []float64{c.RecencyWeighted, c.FrequencyWeighted, c.MonetaryWeighted}
}

func (c *UserRFM) Distance(p2 clusters.Coordinates) float64 {
	var r float64
	for i, v := range c.Coordinates() {
		r += math.Pow(v-p2.Coordinates()[i], 2)
	}
	return r
}
