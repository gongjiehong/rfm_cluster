package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"rfm_cluster/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

func Index(c *gin.Context) {
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

		rTime, err := time.Parse(time.DateOnly, row[6])
		if err != nil {
			log.Fatal(err)
		}
		recency := cast.ToFloat64((1738339200000 - rTime.UnixMilli()) / 86400000)

		temp := models.UserRFM{
			UserID:    cast.ToUint64(row[0]),
			Nickname:  row[1],
			Birthday:  cast.ToInt64(row[2]),
			Gender:    cast.ToInt8(row[3]),
			Recency:   recency,
			Frequency: cast.ToFloat64(row[4]),
			Monetary:  cast.ToFloat64(row[5]),
		}

		originalData = append(originalData, &temp)
	}

	processedData := models.ProcessRMFDataIndicators(originalData)

	renderMap := map[string]interface{}{}

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

	renderMap["processedData"] = processedDataMap

	renderMap["OriginalDataChartContent"] = models.ProcessOriginalDataChart(originalData)

	processedRFMscatter3d, originalRFMscatter3d := models.ProcessCluteredOriginalDataChart(originalData)

	renderMap["ClusteredDataChartContent"] = processedRFMscatter3d
	renderMap["ClusteredOriginalDataChartContent"] = originalRFMscatter3d

	c.HTML(200, "dash.html", renderMap)

	// c.JSON(http.StatusOK, temp)
}
