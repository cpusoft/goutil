package httpserver

import (
	"net/http"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"

	_ "github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/logs"
)

func ListenAndServer(port string, router *rest.App) {

	api := rest.NewApi()
	MyAccessProdStack := rest.AccessProdStack
	MyAccessProdStack[0] = &rest.AccessLogApacheMiddleware{
		Logger: belogs.GetLogger("access"),
		Format: rest.CombinedLogFormat,
	}
	api.Use(MyAccessProdStack...)
	api.SetApp(*router)
	belogs.Emergency(http.ListenAndServe(port, api.MakeHandler()))
}
