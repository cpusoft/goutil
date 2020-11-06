package httpserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
	osutil "github.com/cpusoft/goutil/osutil"
)

// result:ok/fail
type HttpResponse struct {
	Result string `json:"result"`
	Msg    string `json:"msg"`
}

// setup Http Server, listen on port
func ListenAndServe(port string, router *rest.App) {

	api := rest.NewApi()

	MyAccessProdStack := rest.AccessProdStack
	MyAccessProdStack[0] = &rest.AccessLogApacheMiddleware{
		Logger: belogs.GetLogger("access"),
		Format: rest.CombinedLogFormat,
	}
	api.Use(MyAccessProdStack...)

	//api.Use(rest.DefaultDevStack...)
	api.SetApp(*router)
	err := http.ListenAndServe(port, api.MakeHandler())
	belogs.Emergency("Start Http Server failed to start, error is ", err)
	fmt.Println("Http Server failed to start, the error is ", err)

}

// setup Https Server, listen on port. need crt and key files
func ListenAndServeTLS(port string, crtFile string, keyFile string, router *rest.App) {

	api := rest.NewApi()

	MyAccessProdStack := rest.AccessProdStack
	MyAccessProdStack[0] = &rest.AccessLogApacheMiddleware{
		Logger: belogs.GetLogger("access"),
		Format: rest.CombinedLogFormat,
	}
	api.Use(MyAccessProdStack...)

	api.Use(rest.DefaultDevStack...)
	api.SetApp(*router)
	//belogs.Emergency(http.ListenAndServe(port, api.MakeHandler()))
	err := http.ListenAndServeTLS(port, crtFile, keyFile, api.MakeHandler())
	belogs.Emergency("Start Https Server failed to start, error is ", err)
	fmt.Println("Https Server failed to start, the error is ", err)
}

// return: map[fileFormName]=fileName, such as map["file1"]="aabbccdd.txt"
func ReceiveFiles(receiveDir string, r *rest.Request) (receiveFiles map[string]string, err error) {
	//belogs.Debug("ReceiveFiles(): receiveDir:", receiveDir)
	defer r.Body.Close()

	reader, err := r.MultipartReader()
	if err != nil {
		belogs.Error("ReceiveFiles(): err:", err)
		return nil, err
	}
	receiveFiles = make(map[string]string)
	for {
		part, err := reader.NextPart()
		if err == io.EOF || part == nil {
			break
		}
		if !strings.HasSuffix(receiveDir, string(os.PathSeparator)) {
			receiveDir = receiveDir + string(os.PathSeparator)
		}
		file := receiveDir + osutil.Base(part.FileName())
		form := strings.TrimSpace(part.FormName())
		belogs.Debug("ReceiveFiles():FileName:", part.FileName(), "   FormName:", part.FormName()+"   file:", file)
		if part.FileName() == "" { // this is FormData
			data, _ := ioutil.ReadAll(part)
			ioutil.WriteFile(file, data, 0644)
		} else { // This is FileData
			dst, _ := os.Create(file)
			defer dst.Close()
			io.Copy(dst, part)
		}
		receiveFiles[form] = file
	}
	return receiveFiles, nil
}

//map[fileFormName]=fileName
func RemoveReceiveFiles(receiveFiles map[string]string) {
	if len(receiveFiles) == 0 {
		return
	}
	for _, v := range receiveFiles {
		//belogs.Debug("RemoveReceiveFiles(): form:", k, "    filename:", v)
		os.Remove(v)
	}
}
func GetOkHttpResponse() HttpResponse {
	return GetOkMsgHttpResponse("")

}
func GetOkMsgHttpResponse(msg string) HttpResponse {
	return HttpResponse{
		Result: "ok",
		Msg:    msg}
}

func GetFailHttpResponse(err error) HttpResponse {
	return HttpResponse{
		Result: "fail",
		Msg:    err.Error()}

}
