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
		} else if row.MonetaryOriginal > 39.99 && row.MonetaryOriginal <= 59.99 {
			m = 4
		} else if row.MonetaryOriginal > 59.99 {
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
