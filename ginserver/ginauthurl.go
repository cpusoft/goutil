package ginserver

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func RegisterCheckAuthUrls(app *gin.Engine, skipUrls []string, roleHasUrls map[uint64][]string) {
	app.Use(checkAuthUrls(
		skipAuthUrls(skipUrls),
	))
	app.Use(checkAuthUrls(
		roleHasAuthUrls(roleHasUrls),
	))

}

// check Func
type checkAuthUrlsFunc func(*gin.Context) bool

func checkAuthUrls(checkAuthUrlsFuncs ...checkAuthUrlsFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(checkAuthUrlsFuncs) > 0 && checkAuthUrlsFuncs[0](c) {
			c.Next()
			return
		}
		ResponseFail(c, errors.New("No auth"), nil)
		c.Abort()
		return
	}
}

// if the request path contains url(prefix), skip if it contains
func skipAuthUrls(skipUrls []string) checkAuthUrlsFunc {
	return func(c *gin.Context) bool {
		path := c.Request.URL.Path
		pathLen := len(path)
		for _, p := range skipUrls {
			if pl := len(p); pathLen >= pl && path[:pl] == p {
				return true
			}
		}
		return false
	}
}

// if the request path contains url(prefix) according to roles, skip if it contains
func roleHasAuthUrls(roleHasUrls map[uint64][]string) checkAuthUrlsFunc {
	return func(c *gin.Context) bool {
		path := c.Request.URL.Path
		pathLen := len(path)

		ginUserModel := GinUserModel{}
		err := GetUserFromSession(c, &ginUserModel)
		if err != nil || ginUserModel.Id == 0 {
			return false
		}
		skipUrls, ok := roleHasUrls[ginUserModel.RoleId]
		if !ok {
			return false
		}
		for _, p := range skipUrls {
			if pl := len(p); pathLen >= pl && path[:pl] == p {
				return true
			}
		}
		return false
	}
}
