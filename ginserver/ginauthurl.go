package ginserver

import (
	"errors"
	"regexp"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

func RegisterCheckAuthUrls(app *gin.Engine, skipUrls []string, roleHasUrls map[uint64][]string) {
	app.Use(checkAuthUrls(
		skipAuthUrlsOrRoleHasAuthUrls(skipUrls, roleHasUrls),
	))

}

// check Func
type checkAuthUrlsFunc func(*gin.Context) bool

func checkAuthUrls(checkAuthUrlsFuncs ...checkAuthUrlsFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(checkAuthUrlsFuncs) > 0 && checkAuthUrlsFuncs[0](c) {
			belogs.Debug("checkAuthUrls(): checkAuthUrlsFuncs[0](c) pass: ", checkAuthUrlsFuncs[0])
			c.Next()
			return
		}
		ResponseFail(c, errors.New("No auth"), nil)
		c.Abort()
		return
	}
}

// if the request path contains skip url(prefix), or the roles has the ruls, skip if it contains
func skipAuthUrlsOrRoleHasAuthUrls(skipUrls []string, roleHasUrls map[uint64][]string) checkAuthUrlsFunc {
	return func(c *gin.Context) bool {
		reqPath := c.Request.URL.Path

		// check if in skipUrls
		belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls(): reqPath:", reqPath, "   skipUrls:", skipUrls)
		for _, skipUrl := range skipUrls {

			// if equal
			if skipUrl == reqPath {
				return true
			} else if strings.HasSuffix(skipUrl, "*") {
				//if endwith, eg: /static/*
				reg := regexp.MustCompile(skipUrl).MatchString(reqPath)
				if reg {
					return true
				}
			}
		}

		// check if role has urls
		ginUserModel := GinUserModel{}
		err := GetUserFromSession(c, &ginUserModel)
		if err != nil || ginUserModel.Id == 0 {
			belogs.Error("skipAuthUrlsOrRoleHasAuthUrls(): reqPath:", reqPath, " get  ginUserModel fail")
			return false
		}

		roleUrls, ok := roleHasUrls[ginUserModel.RoleId]
		belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls(): reqPath:", reqPath, "   roleUrls:", jsonutil.MarshalJson(roleUrls))
		if !ok {
			return false
		}
		for _, roleUrl := range roleUrls {
			if roleUrl == reqPath {
				return true
			} else if strings.HasSuffix(roleUrl, "*") {
				//if endwith, eg: /static/*
				reg := regexp.MustCompile(roleUrl).MatchString(reqPath)
				if reg {
					return true
				}
			}
		}
		belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():auth false, reqPath:", reqPath)
		return false
	}
}
