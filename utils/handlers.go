package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"time"
	"strings"
	"github.com/shirou/gopsutil/process"
	"strconv"
	"os"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"github.com/satori/go.uuid"
	"io"
	"log"
)

type User struct {
	Name     string `form:"username"`
	Password string `form:"password"`
}

type FileList struct {
	Files []string `json:"files"`
}

type ManageProcess struct {
	Pid        string `form:"pid"`
	Operation  string `form:"operation"`
	CreateTime string `form:"createTime"`
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
}

func logErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func LoginHandler_post(c *gin.Context) {
	var user User
	err := c.ShouldBind(&user)
	checkErr(err)
	if user.Name != Username || user.Password != Password {
		c.String(http.StatusUnauthorized, "failed")
	} else {
		//c.SetCookie(CookieName, CookieValue, 0, "/", "", false, true)
		//session := sessions.Default(c)
		//session.Set("login", "true")
		//session.Save()
		c.String(http.StatusOK, Token)
	}
}

func SystemInfoHandler_post(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"log_info": GetLog_Info(Logfile),
		"cpu_info": GetCpu_Info(),
		"sys_info": GetSys_Info(),
		"mem_info": GetMem_Info(),
		"swap_info": GetSwap_Info(),
		"disk_info": GetDisk_Info(),
		"network_info": GetNetwork_Info(),
	})
}

var wsupgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024,CheckOrigin: func(r *http.Request) bool {
	return true
},}

func SystemInfoHandler_ws(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	checkErr(err)
	ticker := time.NewTicker(time.Second * 3)
	defer func() {
		ticker.Stop()
	}()
	err = conn.WriteJSON(gin.H{"log_info": GetLog_Info(Logfile),
		"cpu_info": GetCpu_Info(),
		"sys_info": GetSys_Info(),
		"mem_info": GetMem_Info(),
		"swap_info": GetSwap_Info(),
		"disk_info": GetDisk_Info(),
		"network_info": GetNetwork_Info(),
	})
	if err != nil {
		//log.Println("websocket disconnect")
		return
	}
	for range ticker.C {
		//log.Println("websocket ok")
		err := conn.WriteJSON(gin.H{"log_info": GetLog_Info(Logfile),
			"cpu_info": GetCpu_Info(),
			"sys_info": GetSys_Info(),
			"mem_info": GetMem_Info(),
			"swap_info": GetSwap_Info(),
			"disk_info": GetDisk_Info(),
			"network_info": GetNetwork_Info(),
		})
		if err != nil {
			//log.Println("websocket disconnect")
			break
		}
	}
}

func GetProcessInfoHandler_post(c *gin.Context) {
	info, err := GetProcess_Info()
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"ProcessInfo": info,
		})
	} else {
		c.String(http.StatusBadRequest, fmt.Sprint(err))
	}
}

func ManageProcessInfoHandler_post(c *gin.Context) {
	var mp ManageProcess
	err := c.ShouldBind(&mp)
	if err == nil {
		pid, err := strconv.Atoi(mp.Pid)
		if err == nil {
			if b, err := process.PidExists(int32(pid)); b && err == nil {
				pro, err := process.NewProcess(int32(pid))
				if err == nil {
					createTime, err := pro.CreateTime()
					if err == nil && fmt.Sprint(createTime) == mp.CreateTime {
						if mp.Operation == "suspend" {
							err = pro.Suspend()
							if err == nil {
								c.String(http.StatusOK, mp.Pid+" succeed")
							} else {
								c.String(http.StatusInternalServerError, fmt.Sprint(err))
							}
						} else if mp.Operation == "resume" {
							err = pro.Resume()
							if err == nil {
								c.String(http.StatusOK, mp.Pid+" succeed")
							} else {
								c.String(http.StatusInternalServerError, fmt.Sprint(err))
							}
						} else if mp.Operation == "terminate" {
							err = pro.Terminate()
							if err == nil {
								c.String(http.StatusOK, mp.Pid+" succeed")
							} else {
								c.String(http.StatusInternalServerError, fmt.Sprint(err))
							}
						} else if mp.Operation == "kill" {
							err = pro.Kill()
							if err == nil {
								c.String(http.StatusOK, mp.Pid+" succeed")
							} else {
								c.String(http.StatusInternalServerError, fmt.Sprint(err))
							}
						} else {
							c.String(http.StatusBadRequest, "invalid operation")
						}
					} else {
						c.String(http.StatusBadRequest, "create time not match")
					}
				} else {
					c.String(http.StatusBadRequest, fmt.Sprint(err))
				}
			} else {
				c.String(http.StatusBadRequest, "process not exist")
			}
		} else {
			c.String(http.StatusBadRequest, fmt.Sprint(err))
		}
	} else {
		c.String(http.StatusBadRequest, fmt.Sprint(err))
	}
}

var wsupgrader2 = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024,CheckOrigin: func(r *http.Request) bool {
	return true
},}

func ProcessInfoHandler_ws(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader2.Upgrade(w, r, nil)
	checkErr(err)
	ticker := time.NewTicker(time.Second * 3)
	defer func() {
		ticker.Stop()
	}()

	info, err := GetProcess_Info()
	if err == nil {
		err := conn.WriteJSON(gin.H{
			"ProcessInfo": info,
		})
		if err != nil {
			//log.Println("websocket disconnect")
			conn.Close()
			return
		}
	}

	go func() {
		for {
			t, msg, err := conn.ReadMessage()
			//log.Println(err)
			if err == nil {
				s := string(msg[:])
				info := strings.Split(s, " ")
				pid, err := strconv.Atoi(info[0])
				if err == nil {
					if b, err := process.PidExists(int32(pid)); b && err == nil {
						pro, err := process.NewProcess(int32(pid))
						if err == nil {
							createTime, err := pro.CreateTime()
							if err == nil && fmt.Sprint(createTime) == info[2] {
								if info[1] == "suspend" {
									err = pro.Suspend()
									if err == nil {
										conn.WriteMessage(t, []byte(info[0]+" succeed"))
									} else {
										conn.WriteMessage(t, []byte(fmt.Sprint(err)))
									}
								} else if info[1] == "resume" {
									err = pro.Resume()
									if err == nil {
										conn.WriteMessage(t, []byte(info[0]+" succeed"))
									} else {
										conn.WriteMessage(t, []byte(fmt.Sprint(err)))
									}
								} else if info[1] == "terminate" {
									err = pro.Terminate()
									if err == nil {
										conn.WriteMessage(t, []byte(info[0]+" succeed"))
									} else {
										conn.WriteMessage(t, []byte(fmt.Sprint(err)))
									}
								} else if info[1] == "kill" {
									err = pro.Kill()
									if err == nil {
										conn.WriteMessage(t, []byte(info[0]+" succeed"))
									} else {
										conn.WriteMessage(t, []byte(fmt.Sprint(err)))
									}
								} else {
									conn.WriteMessage(t, []byte("invalid operation"))
								}
							} else {
								conn.WriteMessage(t, []byte("create time not match"))
							}
						} else {
							conn.WriteMessage(t, []byte(fmt.Sprint(err)))
						}
					} else {
						conn.WriteMessage(t, []byte("process not exist"))
					}
				} else {
					conn.WriteMessage(t, []byte(fmt.Sprint(err)))
				}
			} else {
				if fmt.Sprint(err) == "websocket: close 1001 (going away)" {
					return
				}
				conn.WriteMessage(t, []byte(fmt.Sprint(err)))
			}

		}
	}()
	for range ticker.C {
		//log.Println("websocket ok")
		//log.Println(gin.H{
		//	"ProcessInfo":GetProcess_Info(),
		//})
		info, err := GetProcess_Info()
		if err == nil {
			err := conn.WriteJSON(gin.H{
				"ProcessInfo": info,
			})
			if err != nil {
				//log.Println("websocket disconnect")
				conn.Close()
				break
			}
		}
	}
}

func PathHandler_get(c *gin.Context) {
	path := c.DefaultQuery("path", "/")
	var dirs []DirItem
	var files []FileItem
	var writable bool
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			dirs = make([]DirItem, 0)
			files = make([]FileItem, 0)
			writable = false
		}
	} else {
		if !stat.IsDir() {
			dirs = make([]DirItem, 0)
			files = make([]FileItem, 0)
			writable = false
		} else {
			allfiles, err := ioutil.ReadDir(path)
			if err == nil {
				dirs = GetDirs(path, allfiles)
				files = GetFiles(path, allfiles)
				writable = unix.Access(path, unix.W_OK) == nil
			} else {
				dirs = make([]DirItem, 0)
				files = make([]FileItem, 0)
				writable = false
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"path": path, "writable": writable, "dirs": dirs, "files": files})
}

func DownloadHandler_get(c *gin.Context) {
	path := c.DefaultQuery("name", "")
	if len(path) == 0 {
		c.String(http.StatusBadRequest, "invalid url")
	} else {
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename=files.zip")
		c.Header("Content-Type", "application/octet-stream")
		defer func() {
			os.Remove(path)
		}()
		c.File(path)
	}
}

func DownloadHandler_post(c *gin.Context) {
	var filelist FileList
	err := c.BindJSON(&filelist)
	if err == nil {
		for i, file := range filelist.Files {
			path, err := url.QueryUnescape(file)
			if err != nil {
				logErr(err)
				c.String(http.StatusBadRequest, fmt.Sprint(err))
				return
			}
			filelist.Files[i] = path
		}
		fmt.Println(filelist.Files)
		if len(filelist.Files) == 0 {
			logErr(err)
			c.String(http.StatusBadRequest, fmt.Sprint(err))
			return
		}
		openfiles := make([]*os.File, 0)
		for _, filename := range filelist.Files {
			f, err := os.Open(filename)
			defer f.Close()
			if err != nil {
				logErr(err)
				c.String(http.StatusBadRequest, fmt.Sprint(err))
				return
			}
			openfiles = append(openfiles, f)
		}
		ex, err := os.Executable()
		if err != nil {
			logErr(err)
			c.String(http.StatusBadRequest, fmt.Sprint(err))
			return
		}
		expath := filepath.Dir(ex)
		zippath := filepath.Join(expath, uuid.Must(uuid.NewV4()).String()+".zip")
		err = Compress(openfiles, zippath)
		if err != nil {
			logErr(err)
			os.Remove(zippath)
			c.String(http.StatusBadRequest, fmt.Sprint(err))
			return
		}
		c.String(http.StatusOK, "ok "+zippath)
	} else {
		logErr(err)
		c.String(http.StatusBadRequest, fmt.Sprint(err))
	}
}

func UploadHandler_post(c *gin.Context) {
	path:=c.DefaultPostForm("path","")
	if len(path) == 0 {
		c.String(http.StatusBadRequest, "invalid path")
	} else {
		file, header, err := c.Request.FormFile("files")
		logErr(err)
		filename := header.Filename
		out, err := os.Create(filepath.Join(path, filename))
		logErr(err)
		defer out.Close()
		_, err = io.Copy(out, file)
		logErr(err)
		c.JSON(http.StatusOK, gin.H{"name": filename})
	}
}

func DeleteHandler_post(c *gin.Context) {
	var filelist FileList
	err := c.BindJSON(&filelist)
	if err == nil {
		for i, file := range filelist.Files {
			path, err := url.QueryUnescape(file)
			if err != nil {
				logErr(err)
				c.String(http.StatusBadRequest, fmt.Sprint(err))
				return
			}
			filelist.Files[i] = path
		}
		if len(filelist.Files) == 0 {
			logErr(err)
			c.String(http.StatusBadRequest, fmt.Sprint(err))
			return
		}
		for _, file := range filelist.Files {
			err = os.RemoveAll(file)
			if err != nil {
				logErr(err)
				c.String(http.StatusBadRequest, fmt.Sprint(err))
				return
			}
		}
		c.String(http.StatusOK, "ok")
	} else {
		logErr(err)
		c.String(http.StatusBadRequest, fmt.Sprint(err))
	}
}

func PingHandler_post(c *gin.Context)  {
	c.String(http.StatusOK,"ok")
}
