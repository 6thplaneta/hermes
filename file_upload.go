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
	uniquefid := uuid.NewV4().String()
	r := c.Request
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile(inputName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	contentType := handler.Header.Get("Content-Type") + "/"

	// if handler.Header.Get("Content-Type") != "image/jpeg" {
	// 	return "", errors.New("Image should be jpeg!")
	// }
	year, month, _ := time.Now().Date()
	var path_ string = "./upload/"
	if savePath != "" {
		path_ = savePath
	}
	parts := strings.Split(contentType, "/")
	path_ += parts[0] + "/"
	path_ += strconv.Itoa(year) + "/" + strconv.Itoa(int(month)) + "/"
	err = os.MkdirAll(path_, 0777)
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(path_+uniquefid+path.Ext(handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	io.Copy(f, file)
	return path_ + uniquefid + path.Ext(handler.Filename), nil
}
