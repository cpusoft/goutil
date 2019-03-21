package httpserver

import (
	"net/http"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/go-json-rest/rest"
)

// setup Http Server, listen on port
func ListenAndServe(port string, router *rest.App) {

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

// setup Https Server, listen on port. need crt and key files
func ListenAndServeTLS(port string, crtFile string, keyFile string, router *rest.App) {

	api := rest.NewApi()
	MyAccessProdStack := rest.AccessProdStack
	MyAccessProdStack[0] = &rest.AccessLogApacheMiddleware{
		Logger: belogs.GetLogger("access"),
		Format: rest.CombinedLogFormat,
	}
	api.Use(MyAccessProdStack...)
	api.SetApp(*router)
	//belogs.Emergency(http.ListenAndServe(port, api.MakeHandler()))
	belogs.Emergency(http.ListenAndServeTLS(port, crtFile, keyFile, api.MakeHandler()))
}
