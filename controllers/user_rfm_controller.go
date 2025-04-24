package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"rfm_cluster/models"
	"rfm_cluster/pkg/clusters"
	"rfm_cluster/pkg/silhouette"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

var colors = []string{
	"#ff5722",
	"#ffb800",
	"#16baaa",
	"#1e9fff",
	"#a233c6",
	"#2f363c",
	"#c2c2c2",
}

func Index(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	// k := queryParams.Get("k")
	purchaseEnd := queryParams.Get("purchase_end")

	excel, err := excelize.OpenFile("论文数据8月前-未加权.xlsx")
	if err != nil {
		c.JSON(http.StatusOK, err.Error())
		return
	}

	rows, err := excel.GetRows("Sheet1")
	if err != nil {
		c.JSON(http.StatusOK, err.Error())
		return
	}

	originalData := []*models.UserRFM{}
	for index, row := range rows {
		if index == 0 {
			continue
		}

		// if index%10 != 0 {
		// 	continue
		// }

		rTime, err := time.Parse(time.DateOnly, row[6])
		if err != nil {
			log.Fatal(err)
		}
		recency := cast.ToFloat64((cast.ToInt64(purchaseEnd) - rTime.UnixMilli()) / 86400000)

		temp := models.UserRFM{
			UserID:            cast.ToUint64(row[0]),
			Nickname:          row[1],
			Birthday:          cast.ToInt64(row[2]),
			Gender:            cast.ToInt8(row[3]),
			RecencyOriginal:   recency,
			FrequencyOriginal: cast.ToFloat64(row[4]),
			MonetaryOriginal:  cast.ToFloat64(row[5]),
		}

		originalData = append(originalData, &temp)
	}

	_, scores, estimate, _, err := models.ProcessData(originalData)
	if err != nil {
		c.JSON(http.StatusOK, err.Error())
		return
	}

	waitGroup := sync.WaitGroup{}
	renderMap := map[string]interface{}{}
	renderMap["EstimateCluters"] = estimate
	lock := sync.Mutex{}

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		processedData := models.ProcessRMFDataIndicators(originalData)

		processedDataBytes, err := json.Marshal(&processedData)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}

		processedDataMap := map[string]interface{}{}
		err = json.Unmarshal(processedDataBytes, &processedDataMap)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}

		lock.Lock()
		renderMap["processedData"] = processedDataMap
		lock.Unlock()
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		tempHTML := ProcessOriginalDataChart(originalData)
		lock.Lock()
		renderMap["OriginalDataChartContent"] = tempHTML
		lock.Unlock()
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		line, err := ProcessSilhouetteLineChart(scores)
		if err != nil {
			c.JSON(http.StatusOK, err.Error())
			return
		}

		lock.Lock()
		renderMap["CluteredSilhouette"] = line
		lock.Unlock()
	}()

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		processedRFMscatter3d, originalRFMscatter3d := ProcessCluteredAndOriginalDataChart(originalData, scores[estimate-2].Clusters)

		lock.Lock()
		renderMap["ClusteredDataChartContent"] = processedRFMscatter3d
		renderMap["ClusteredOriginalDataChartContent"] = originalRFMscatter3d
		lock.Unlock()
	}()

	waitGroup.Wait()

	c.HTML(200, "dash.html", renderMap)
}

// 绘制原始数据在3D坐标中的图表
func ProcessOriginalDataChart(dataCollection []*models.UserRFM) template.HTML {
	results := []opts.Chart3DData{}

	for _, data := range dataCollection {
		results = append(results, opts.Chart3DData{
			Value: []interface{}{data.RecencyOriginal, data.FrequencyOriginal, data.MonetaryOriginal},
			ItemStyle: &opts.ItemStyle{
				Color: colors[0],
			},
		})
	}

	scatter3d := charts.NewScatter3D()
	scatter3d.AssetsHost = "/statics/echarts/"

	// set some global options like Title/Legend/ToolTip or anything else
	scatter3d.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{Name: "Recency", Show: opts.Bool(true)}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Name: "Frequency", Show: opts.Bool(true)}),
		charts.WithZAxis3DOpts(opts.ZAxis3D{Name: "Monetary", Show: opts.Bool(true)}),
	)

	scatter3d.AddSeries("", results)

	buffer := scatter3d.RenderContent()

	return template.HTML(string(buffer))
}

func ProcessCluteredAndOriginalDataChart(dataCollection []*models.UserRFM, clusters clusters.Clusters) (template.HTML, template.HTML) {
	processedRFM := []opts.Chart3DData{}
	originalRFM := []opts.Chart3DData{}
	for i, c := range clusters {
		for _, o := range c.Observations {
			processedRFM = append(processedRFM, opts.Chart3DData{
				Value: []interface{}{o.Coordinates()[0], o.Coordinates()[1], o.Coordinates()[2]},
				ItemStyle: &opts.ItemStyle{
					Color: colors[i],
				},
			})

			rfm := o.(*models.UserRFM)
			for _, user := range dataCollection {
				if rfm.UserID == user.UserID {
					originalRFM = append(originalRFM, opts.Chart3DData{
						Value: []interface{}{user.RecencyOriginal, user.FrequencyOriginal, user.MonetaryOriginal},
						ItemStyle: &opts.ItemStyle{
							Color: colors[i],
						},
					})
				}
			}
		}
	}

	processedRFMscatter3d := charts.NewScatter3D()
	processedRFMscatter3d.AssetsHost = "/statics/echarts/"
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
	originalRFMscatter3d.AssetsHost = "/statics/echarts/"

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

func ProcessSilhouetteLineChart(scores []silhouette.KScore) (template.HTML, error) {
	line := charts.NewLine()
	line.AssetsHost = "/statics/echarts/"

	titles := []string{}
	lineData := []opts.LineData{}
	for _, score := range scores {
		titles = append(titles, fmt.Sprintf("k = %d", score.K))
		lineData = append(lineData, opts.LineData{Value: score.Score})
	}

	line.SetXAxis(titles).AddSeries("", lineData).
		SetSeriesOptions(
			charts.WithMarkPointNameTypeItemOpts(
				opts.MarkPointNameTypeItem{Name: "Maximum", Type: "max"},
				opts.MarkPointNameTypeItem{Name: "Average", Type: "average"},
				opts.MarkPointNameTypeItem{Name: "Minimum", Type: "min"},
			),
			charts.WithMarkPointStyleOpts(
				opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(true)}}),
		)

	return template.HTML(line.RenderContent()), nil
}
