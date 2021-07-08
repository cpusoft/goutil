package ginserver

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

// decode json
func DecodeJson(c *gin.Context, v interface{}) error {
	content, err := ioutil.ReadAll(c.Request.Body)
	c.Request.Body.Close()
	if err != nil {
		return err
	}
	if len(content) == 0 {
		return errors.New("JSON payload is empty")
	}
	err = jsonutil.UnmarshalJson(string(content), v)
	if err != nil {
		return err
	}
	return nil
}

// reslut:ok/fail ;
// msg: error ;
// data: more info ;
type ResponseModel struct {
	Result string      `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
}

func Html(c *gin.Context, html string, v interface{}) {
	c.Header("Cache-Control", "no-cache")
	c.HTML(http.StatusOK, html, v)
}

func ResponseOk(c *gin.Context, v interface{}) {
	c.Header("Cache-Control", "no-cache")
	ret := ResponseModel{Result: "ok", Msg: "", Data: v}
	responseJSON(c, http.StatusOK, &ret)
}

func ResponseFail(c *gin.Context, err error, v interface{}) {
	c.Header("Cache-Control", "no-cache")
	ret := ResponseModel{Result: "fail", Msg: err.Error(), Data: v}
	responseJSON(c, http.StatusOK, &ret)
}

func responseJSON(c *gin.Context, status int, v interface{}) {
	c.JSON(status, v)
}

func ReceiveFile(c *gin.Context, dir string) (receiveFile string, err error) {
	belogs.Debug("ReceiveFile():dir:", dir)
	file, err := c.FormFile("file")
	if err != nil {
		belogs.Error("ReceiveFile(): FormFile fail:", err)
		return "", err
	}
	if !strings.HasSuffix(dir, string(os.PathSeparator)) {
		dir = dir + string(os.PathSeparator)
	}
	receiveFile = dir + file.Filename
	err = c.SaveUploadedFile(file, receiveFile)
	if err != nil {
		belogs.Error("ReceiveFile(): SaveUploadedFile fail:", receiveFile, err)
		return "", err
	}
	belogs.Info("ReceiveFile(): ok, receiveFile:", receiveFile)
	return receiveFile, nil
}
