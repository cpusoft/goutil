package ginserver

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	belogs "github.com/cpusoft/goutil/belogs"
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

	postForm := c.PostForm("name")
	file, err := c.FormFile("file")
	belogs.Debug("ReceiveFile():dir:", dir, "  postForm:", postForm, "   file:", file)
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

	/*
		belogs.Debug("ReceiveFiles(): receiveDir:", receiveDir)
		r := c.Request
		defer r.Body.Close()

		reader, err := r.MultipartReader()
		if err != nil {
			belogs.Error("ReceiveFiles(): err:", err)
			return "", err
		}

		part, err := reader.NextPart()
		if err == io.EOF || part == nil {
			belogs.Error("ReceiveFiles(): NextPart fail:", err)
			return "", errors.New("NextPart is empty")
		}
		if !strings.HasSuffix(dir, string(os.PathSeparator)) {
			dir = dir + string(os.PathSeparator)
		}
		receiveFile = dir + osutil.Base(part.FileName())
		form := strings.TrimSpace(part.FormName())
		belogs.Debug("ReceiveFiles():part.FileName:", part.FileName(),
			"   part.FormName:", part.FormName(),
			"   receiveFile:", receiveFile, "   form:", form)
		if part.FileName() == "" { // this is FormData
			data, _ := ioutil.ReadAll(part)
			ioutil.WriteFile(receiveFile, data, 0644)
		} else { // This is FileData
			dst, _ := os.Create(receiveFile)
			defer dst.Close()
			io.Copy(dst, part)
		}

		return receiveFile, nil
	*/
}
