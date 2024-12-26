package routers

import (
	"net/http"
	con "server/controller"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(Cors())
	r.Static("/web", "../web")
	// r.LoadHTMLGlob("../web/**/*")
	r.LoadHTMLFiles("../web/index.html")
	r.GET("/rsakey", con.GenerateKeyPair)
	r.POST("/rsakeyrestore", con.RestoreKey)
	r.GET("/downloadfile", con.DownloadFile)
	r.POST("/upload", con.Upload)
	r.POST("/createfile", con.CreateFile)
	r.POST("/deletefile", con.DeleteFile)
	r.POST("/getrecords", con.GetRecords)
	r.POST("/getfile", con.GetFile)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})
	return r
}
func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, x-token")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		context.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
	}
}
