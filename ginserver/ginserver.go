package ginserver

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/httpserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
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

func String(c *gin.Context, v string) {
	c.String(http.StatusOK, "%s", v)
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

// if dir=="", then will use tempDir
// outside: if receiveFile!="" then should call os.Remove(receiveFile)
func ReceiveFileAndUnmarshalJson(c *gin.Context, dir string, f interface{}) (receiveFile string, err error) {
	if len(dir) == 0 {
		dir = os.TempDir()
	}

	receiveFile, err = ReceiveFile(c, dir)
	if err != nil {
		belogs.Error("ReceiveFileAndUnmarshalJson():ReceiveFile fail, dir:", dir, err)
		return "", err
	}

	bytes, err := fileutil.ReadFileToBytes(receiveFile)
	if err != nil {
		belogs.Error("ReceiveFileAndUnmarshalJson():ReadFileToBytes fail, receiveFile:", receiveFile, err)
		return receiveFile, err
	}

	err = json.Unmarshal(bytes, &f)
	if err != nil {
		belogs.Error("ReceiveFileAndUnmarshalJson():Unmarshal fail,receiveFile :", receiveFile,
			"   content:", string(bytes), err)
		return receiveFile, err
	}
	belogs.Info("ReceiveFileAndUnmarshalJson():receiveFile:", receiveFile, "   f:", jsonutil.MarshalJson(f))
	return receiveFile, nil
}

func ReceiveFileAndPostNewUrl(c *gin.Context, newUrl string) (err error) {

	belogs.Debug("ReceiveFileAndPostNewUrl(): newUrl:", newUrl)
	fileHeader, err := c.FormFile("file")
	tmpFile, tmpDir, err := saveToTmpFile(fileHeader)
	defer func() {
		osutil.CloseAndRemoveFile(tmpFile)
		os.Remove(tmpDir)
	}()
	if err != nil {
		belogs.Error("ReceiveFileAndPostNewUrl(): saveToTmpFile fail:", err)
		return err
	}
	belogs.Info("ReceiveFileAndPostNewUrl():saveToTmpFile tmpFile:", tmpFile.Name(), "  newUrl:", newUrl)

	resp, body, err := httpclient.PostFile(newUrl, tmpFile.Name(), "file", false)
	if err != nil {
		belogs.Error("ReceiveFileAndPostNewUrl(): upload PostFileHttps failed, err:", newUrl, err)
		return err
	}
	resp.Body.Close()
	belogs.Debug("ReceiveFileAndPostNewUrl():upload body:", body)

	httpResponse := httpserver.HttpResponse{}
	jsonutil.UnmarshalJson(body, &httpResponse)
	belogs.Debug("ReceiveFileAndPostNewUrl(): upload response:", newUrl, jsonutil.MarshalJson(httpResponse))
	if httpResponse.Result == "fail" {
		belogs.Error("ReceiveFileAndPostNewUrl(): upload response failed, err:", newUrl, jsonutil.MarshalJson(httpResponse))
		return errors.New(httpResponse.Msg)

	}
	belogs.Info("ReceiveFileAndPostNewUrl(): upload ok ", fileHeader.Filename, jsonutil.MarshalJson(httpResponse))
	return nil
}

func saveToTmpFile(fileHeader *multipart.FileHeader) (tmpFile *os.File, tmpDir string, err error) {
	if fileHeader == nil {
		belogs.Error("saveToTmpFile(): fileHeader is nil")
		return nil, tmpDir, errors.New("upload file is empty")
	}
	// get file
	file, err := fileHeader.Open()
	if err != nil {
		belogs.Error("saveToTmpFile(): fileHeader.Open fail:", err)
		return nil, tmpDir, err
	}
	defer file.Close()
	belogs.Debug("saveToTmpFile(): fileHeader.Filename:", fileHeader.Filename)

	// create tmp file
	tmpDir, err = ioutil.TempDir("", "tmp-")
	if err != nil {
		belogs.Error("saveToTmpFile(): TempDir fail:", err)
		return nil, tmpDir, err
	}
	tmpFile, err = os.Create(tmpDir + string(os.PathSeparator) + fileHeader.Filename)
	if err != nil {
		belogs.Error("saveToTmpFile(): TempFile fail:", err)
		return nil, tmpDir, err
	}

	belogs.Debug("saveToTmpFile(): tmpFile:", tmpFile.Name())

	// save to tmp file
	_, err = io.Copy(tmpFile, file)
	if err != nil {
		belogs.Error("saveToTmpFile(): Copy fail:", err)
		return nil, tmpDir, err
	}

	return tmpFile, tmpDir, nil
}
