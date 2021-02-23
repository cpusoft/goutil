package ginserver

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

const (
	TOKEN_KEY     = "X-Token"  //Page token key name
	USER_ID_Key   = "X-USERID" //Page user ID key name
	USER_UUID_Key = "X-UUID"   //UUID key name
	AUTH_URL      = "authurl"  //session key name
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
