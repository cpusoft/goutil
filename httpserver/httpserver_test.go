package httpserver

import (
	"fmt"
	"github.com/cpusoft/go-json-rest/rest"
	"testing"

	httpserver "github.com/cpusoft/goutil/httpserver"
)

func TestListenAndServe(t *testing.T) {

	router, err := rest.MakeRouter(
		rest.Get("/hello", hello),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		httpserver.ListenAndServe(":8080", &router)
	}()
	go func() {
		httpserver.ListenAndServeTLS(":8081", `E:\Go\common-util\conf\server.crt`, `E:\Go\common-util\conf\server.key`, &router)
	}()
	select {}
}

// 主动往RP发起请求，要求获取全量数据
func hello(w rest.ResponseWriter, req *rest.Request) {
	w.WriteJson("hello world")
}
