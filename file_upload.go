package hermes

import (
	// "errors"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

/**
* File Uploader
*
* handle file uploading process
* @param 	gin.Context		a gin Context input
* @param 	string 			name of file input field , like image , avatar , file and etc...
* @return	string 			file uploaded path on success
* @return	error 			error if fails
* @author	Mostafa Solati <Mostafa.solati@gmail.com>
 */
func Upload(c *gin.Context, inputName, savePath string) (string, error) {
	//generate unique filename
	uniquefid := uuid.NewV4().String()
	r := c.Request
	r.ParseMultipartForm(32 << 20)
	//get file content
	file, handler, err := r.FormFile(inputName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	//categorize folders by content-type
	contentType := handler.Header.Get("Content-Type") + "/"

	year, month, _ := time.Now().Date()
	var path_ string = "./upload/"
	if savePath != "" {
		path_ = savePath
	}
	parts := strings.Split(contentType, "/")
	path_ += parts[0] + "/"
	//categorize folders by year and then month

	path_ += strconv.Itoa(year) + "/" + strconv.Itoa(int(month)) + "/"
	err = os.MkdirAll(path_, 0777)
	if err != nil {
		return "", err
	}
	//create new file with

	f, err := os.OpenFile(path_+uniquefid+path.Ext(handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	//create content of upload file to the new file with

	io.Copy(f, file)
	//return name of file
	return path_ + uniquefid + path.Ext(handler.Filename), nil
}
