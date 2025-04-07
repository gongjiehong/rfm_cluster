package models

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"rfm_cluster/pkg/clusters"
	"rfm_cluster/pkg/kmeans"
	"rfm_cluster/pkg/plotter"
	"slices"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type UserRFM struct {
	UserID    uint64  `json:"user_id"`
	Nickname  string  `json:"nickname"`
	Birthday  int64   `json:"birthday"`
	Gender    int8    `json:"gender"`
	Recency   float64 `json:"recency"`
	Frequency float64 `json:"frequency"`
	Monetary  float64 `json:"monetary"`
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

var colors = []string{
	"#ff5722",
	"#ffb800",
	"#16baaa",
	"#1e9fff",
	"#a233c6",
	"#2f363c",
	"#c2c2c2",
}

func ProcessRMFDataIndicators(dataCollection []*UserRFM) RMFDataIndicators {
	rArray := []float64{}
	fArray := []float64{}
	mArray := []float64{}

	for _, data := range dataCollection {
		rArray = append(rArray, data.Recency)
		fArray = append(fArray, data.Frequency)
		mArray = append(mArray, data.Monetary)
	}

	result := RMFDataIndicators{
		R: DataIndicators{Mean: Mean(rArray), Stddev: Stddev(rArray), Variance: Variance(rArray), Min: slices.Min(rArray), Max: slices.Max(rArray)},
		F: DataIndicators{Mean: Mean(fArray), Stddev: Stddev(fArray), Variance: Variance(fArray), Min: slices.Min(fArray), Max: slices.Max(fArray)},
		M: DataIndicators{Mean: Mean(mArray), Stddev: Stddev(mArray), Variance: Variance(mArray), Min: slices.Min(mArray), Max: slices.Max(mArray)},
	}

	return result
}

func ProcessOriginalDataChart(dataCollection []*UserRFM) template.HTML {
	results := []opts.Chart3DData{}

	for _, data := range dataCollection {
		results = append(results, opts.Chart3DData{
			Value: []interface{}{data.Recency, data.Frequency, data.Monetary},
			ItemStyle: &opts.ItemStyle{
				Color: colors[0],
			},
		})
	}

	scatter3d := charts.NewScatter3D()
	// set some global options like Title/Legend/ToolTip or anything else
	scatter3d.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{Name: "Recency", Show: opts.Bool(true)}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Name: "Frequency", Show: opts.Bool(true)}),
		charts.WithZAxis3DOpts(opts.ZAxis3D{Name: "Monetary", Show: opts.Bool(true)}),
	)

	scatter3d.AddSeries("", results)

	buffer := bytes.NewBuffer([]byte{})

	scatter3d.Render(buffer)

	return template.HTML(buffer.String())
}

func ProcessCluteredOriginalDataChart(dataCollection []*UserRFM) (template.HTML, template.HTML) {
	var d clusters.Observations

	idMap := map[clusters.Observation]*UserRFM{}

	for _, row := range dataCollection {

		var r, f, m float64 = 0.0, 0.0, 0.0
		if row.Recency == -1 {
			r = 0
		} else if row.Recency <= 1 {
			r = 5
		} else if row.Recency > 1 && row.Recency <= 7 {
			r = 4
		} else if row.Recency > 7 && row.Recency <= 30 {
			r = 3
		} else if row.Recency > 30 && row.Recency <= 93 {
			r = 2
		} else if row.Recency > 93 {
			r = 1
		}

		if row.Frequency == 1 {
			f = 1
		} else if row.Frequency == 2 {
			f = 2
		} else if row.Frequency == 3 {
			f = 3
		} else if row.Frequency >= 4 && row.Frequency <= 7 {
			f = 4
		} else if row.Frequency >= 7 {
			f = 5
		}

		if row.Monetary <= 6.99 {
			m = 1
		} else if row.Monetary > 6.99 && row.Monetary <= 14.99 {
			m = 2
		} else if row.Monetary > 14.99 && row.Monetary <= 39.99 {
			m = 3
		} else if row.Monetary > 39.99 && row.Monetary <= 59.99 {
			m = 4
		} else if row.Monetary > 59.99 {
			m = 5
		}

		temp := clusters.RFM{
			UserID: row.UserID,
			R:      r,
			F:      f,
			M:      m,
		}
		d = append(d, temp)

		idMap[temp] = row

	}

	fmt.Printf("%d data points\n", len(d))

	// Partition the data points into 7 clusters
	km, err := kmeans.NewWithOptions(0.01, plotter.SimplePlotter{})
	if err != nil {
		log.Fatal(err)
	}
	clusters, _ := km.Partition(d, 3)

	processedRFM := []opts.Chart3DData{}
	originalRFM := []opts.Chart3DData{}
	for i, c := range clusters {
		fmt.Printf("Cluster: %d\n", i)
		fmt.Printf("Centered at x: %.2f y: %.2f\n", c.Center[0], c.Center[1])
		for _, o := range c.Observations {
			processedRFM = append(processedRFM, opts.Chart3DData{
				Value: []interface{}{o.Coordinates()[0], o.Coordinates()[1], o.Coordinates()[2]},
				ItemStyle: &opts.ItemStyle{
					Color: colors[i],
				},
			})

			temp := idMap[o]

			originalRFM = append(originalRFM, opts.Chart3DData{
				Value: []interface{}{temp.Recency, temp.Frequency, temp.Monetary},
				ItemStyle: &opts.ItemStyle{
					Color: colors[i],
				},
			})
		}
	}

	processedRFMscatter3d := charts.NewScatter3D()
	// set some global options like Title/Legend/ToolTip or anything else
	processedRFMscatter3d.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{Name: "Recency", Show: opts.Bool(true)}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Name: "Frequency", Show: opts.Bool(true)}),
		charts.WithZAxis3DOpts(opts.ZAxis3D{Name: "Monetary", Show: opts.Bool(true)}),
	)

	processedRFMscatter3d.AddSeries("", processedRFM)

	processedRFMscatter3dBuffer := bytes.NewBuffer([]byte{})

	processedRFMscatter3d.Render(processedRFMscatter3dBuffer)

	originalRFMscatter3d := charts.NewScatter3D()
	// set some global options like Title/Legend/ToolTip or anything else
	originalRFMscatter3d.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{Name: "Recency", Show: opts.Bool(true)}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Name: "Frequency", Show: opts.Bool(true)}),
		charts.WithZAxis3DOpts(opts.ZAxis3D{Name: "Monetary", Show: opts.Bool(true)}),
	)

	originalRFMscatter3d.AddSeries("", originalRFM)

	originalRFMscatter3dBuffer := bytes.NewBuffer([]byte{})

	originalRFMscatter3d.Render(originalRFMscatter3dBuffer)

	return template.HTML(processedRFMscatter3dBuffer.String()), template.HTML(originalRFMscatter3dBuffer.String())
}
