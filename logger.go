package hermes

import (
	//"io"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Logger struct {
	Level     int
	TheLogger *log.Logger
	LogFile   *os.File
}

func (logger *Logger) UninitLogs() {
	logger.LogFile.Close()
}

func timeFormatter(t time.Time) string {
	return strconv.Itoa(t.Year()) + "_" + strconv.Itoa(int(t.Month())) + "_" + strconv.Itoa(t.Day()) + "_" + strconv.Itoa(t.Hour()) + "_" + strconv.Itoa(t.Minute()) + "_" + strconv.Itoa(t.Second())
}
func (logger *Logger) InitLogs(path string) {
	filepath := path + "log.txt"
	logFile, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("Error in opening log file", err)
		panic(err)
	}

	st, _ := logFile.Stat()
	if st.Size() > 1024*1024 { // 1 MB
		logFile.Close()
		newpath := filepath + "." + timeFormatter(time.Now())
		err = os.Rename(filepath, newpath)
		if err != nil {
			fmt.Println("Error in renaming old log file", err)
			panic(err)
		}

		logFile, err = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			fmt.Println("Error in opening log file", err)
			panic(err)
		}
	}

	logger.TheLogger = log.New(logFile,
		"",
		log.Ldate|log.Ltime)
}

func (logger *Logger) Error(message string) {
	if logger.Level > 0 {
		txt := "ERROR " + message
		logger.TheLogger.Println(txt)
	}
}

func (logger *Logger) Warning(message string) {
	if logger.Level > 1 {
		txt := "Warning " + message
		logger.TheLogger.Println(txt)
	}

}

func (logger *Logger) Info(message string) {
	if logger.Level > 2 {
		txt := "Info " + message
		logger.TheLogger.Println(txt)
	}
}

func (logger *Logger) Trace(message string) {
	if logger.Level > 3 {
		txt := "Trace " + message
		logger.TheLogger.Println(txt)
	}
}

func (logger *Logger) LogHttpByBody(c *gin.Context, body string) {

	// txt := "HTTP Request, Method: " + c.Request.Method + " IP: " + c.ClientIP() + " Path:" + c.Request.RequestURI
	txt := c.Request.RequestURI + " "

	if logger.Level >= 5 {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			txt = txt + "empty "

		} else {
			txt = txt + token + " "
		}

	}
	if logger.Level >= 6 {
		if c.Request.RequestURI != "/upload" {
			txt = txt + strings.Replace(body, "\n", "", -1)
		}
	}
	logger.Trace(txt)
}
func (logger *Logger) LogHttp(c *gin.Context) {
	// txt := "HTTP Request, Method: " + c.Request.Method + " IP: " + c.ClientIP() + " Path:" + c.Request.RequestURI
	txt := c.Request.RequestURI + " "

	if logger.Level >= 5 {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			txt = txt + "empty "

		} else {
			txt = txt + token + " "
		}

	}
	if logger.Level >= 6 {

		reqBody := c.Request.Body

		body, err := ioutil.ReadAll(reqBody)

		if err == nil {
			rdr1 := ioutil.NopCloser(bytes.NewBuffer(body))
			c.Request.Body = rdr1

			if c.Request.RequestURI != "/upload" {
				txt = txt + strings.Replace(string(body), "\n", "", -1)
			}
		}
	}
	logger.Trace(txt)
}
func (logger *Logger) SetLevel(lvl int) {
	if lvl < 0 || lvl > 6 {
		fmt.Println("Log level is between 0 to 6")
	} else {
		logger.Level = lvl
	}
}
