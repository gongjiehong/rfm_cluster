package main

import (
	"fmt"
	"net/http"
	"rfm_cluster/controllers"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", 80),
		Handler:           HTTPRouter(),
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    0,
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func HTTPRouter() *gin.Engine {
	engine := gin.Default()

	engine.LoadHTMLGlob("views/*")

	engine.StaticFS("/statics", http.Dir("./statics"))

	engine.GET("/", controllers.Index)

	return engine
}
