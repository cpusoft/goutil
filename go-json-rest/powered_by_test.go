package rest

import (
	"testing"

	"github.com/cpusoft/goutil/go-json-rest/test"
)

func TestPoweredByMiddleware(t *testing.T) {

	api := NewApi()

	// the middleware to test
	api.Use(&PoweredByMiddleware{
		XPoweredBy: "test",
	})

	// a simple app
	api.SetApp(AppSimple(func(w ResponseWriter, r *Request) {
		w.WriteJson(map[string]string{"Id": "123"})
	}))

	// wrap all
	handler := api.MakeHandler()

	req := test.MakeSimpleRequest("GET", "http://localhost/", nil)
	recorded := test.RunRequest(t, handler, req)
	recorded.CodeIs(200)
	recorded.ContentTypeIsJson()
	recorded.HeaderIs("X-Powered-By", "test")
}
