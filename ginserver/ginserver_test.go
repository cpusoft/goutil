package ginserver

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRunTLSEx(t *testing.T) {
	engine := gin.New()
	engine.GET("/user/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		fmt.Println(message)
		c.String(http.StatusOK, message)
	})

	certFile := `F:\share\我的坚果云\Go\rpstir2\rpstir2-rp\conf\cert\server.crt`
	keyFile := `F:\share\我的坚果云\Go\rpstir2\rpstir2-rp\conf\cert\server.key`
	fmt.Println(certFile, keyFile)
	err := RunTLSEx(engine, ":7771", certFile, keyFile)
	fmt.Println(err)
}
