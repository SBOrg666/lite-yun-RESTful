package main

import (
	"github.com/gin-gonic/gin"
	"github.com/SBOrg666/lite-yun-RESTful/utils"
	"github.com/jasonlvhit/gocron"
	"time"
	"flag"
	"fmt"
	"log"
	"github.com/gin-contrib/cors"
	"github.com/satori/go.uuid"
)

func main() {
	username := flag.String("u", "", "username to login (required)")
	password := flag.String("p", "", "password to login (required)")
	port := flag.Uint("port", 8000, "port to serve")
	logfile := flag.String("l", "/var/log/pacman.log", "path of logfile")
	flag.Parse()

	if len(*username) == 0 || len(*password) == 0 {
		flag.PrintDefaults()
		return
	}

	utils.Username = *username
	utils.Password = *password
	utils.Port = *port
	utils.Logfile = *logfile

	log.Println()
	log.Println("username:" + *username)
	log.Println("password:" + *password)
	log.Println("port:" + fmt.Sprint(*port))
	log.Println("password:" + *logfile)
	log.Println()

	router := gin.Default()
	router.Use(cors.Default())

	utils.Token=uuid.Must(uuid.NewV4()).String()
	log.Println("Token:"+utils.Token)

	router.POST("/login", utils.LoginHandler_post)

	router.GET("/systemInfo", utils.CheckLoginIn(), func(c *gin.Context) {
		utils.SystemInfoHandler_ws(c.Writer, c.Request)
	})
	router.POST("/systemInfo", utils.CheckLoginIn(), utils.SystemInfoHandler_post)

	router.GET("/processInfo", utils.CheckLoginIn(), func(c *gin.Context) {
		utils.ProcessInfoHandler_ws(c.Writer, c.Request)
	})
	router.POST("/getProcessInfo", utils.CheckLoginIn(), utils.GetProcessInfoHandler_post)
	router.POST("/manageProcess", utils.CheckLoginIn(), utils.ManageProcessInfoHandler_post)

	router.GET("/path", utils.CheckLoginIn(), utils.PathHandler_get)

	DownloadGroup := router.Group("/")
	{
		DownloadGroup.GET("/download", utils.CheckLoginIn(), utils.DownloadHandler_get)
		DownloadGroup.POST("/download", utils.CheckLoginIn(), utils.DownloadHandler_post)
	}

	router.POST("/upload", utils.CheckLoginIn(), utils.UploadHandler_post)

	router.POST("/delete", utils.CheckLoginIn(), utils.DeleteHandler_post)

	router.POST("/ping",utils.PingHandler_post)

	utils.Upload_data = make([]uint64, 5)
	utils.Download_data = make([]uint64, 5)
	utils.InitUpload = 0
	utils.InitDownload = 0
	utils.Current_Month = int(time.Now().Month())
	gocron.Every(1).Day().Do(utils.UpdateNetworkData)

	router.Run(":" + fmt.Sprint(*port))
}
