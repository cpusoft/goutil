package ginserver

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
)

// when auth fail, will redirectUrl or failJson.
// redirectUrl: "" or  "/login"
// failJson: ""  or "no auth"
func RegisterCheckAuthUrls(app *gin.Engine,
	skipUrls []string, roleHasUrls map[uint64][]string,
	redirectUrl, failJson string) {
	app.Use(checkAuthUrls(redirectUrl, failJson,
		skipAuthUrlsOrRoleHasAuthUrls(skipUrls, roleHasUrls),
	))

}

// check Func
type checkAuthUrlsFunc func(*gin.Context) bool

func checkAuthUrls(redirectUrl string, failJson string,
	checkAuthUrlsFuncs ...checkAuthUrlsFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(checkAuthUrlsFuncs) > 0 && checkAuthUrlsFuncs[0](c) {
			//belogs.Debug("checkAuthUrls(): checkAuthUrlsFuncs[0](c) pass: ", checkAuthUrlsFuncs[0], "   url:", c.Request.URL.Path)
			c.Next()
			return
		}
		//belogs.Debug("checkAuthUrls(): checkAuthUrlsFuncs[0](c) unpass: ", checkAuthUrlsFuncs[0],
		//	"    redirectUrl:", redirectUrl, "  or  failJson:", failJson)
		if len(redirectUrl) > 0 {
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		} else if len(failJson) > 0 {
			ResponseFail(c, errors.New(failJson), nil)
		}
		c.Abort()
		return
	}
}

// if the request path contains skip url(prefix), or the roles has the ruls, skip if it contains
func skipAuthUrlsOrRoleHasAuthUrls(skipUrls []string, roleHasUrls map[uint64][]string) checkAuthUrlsFunc {
	return func(c *gin.Context) bool {
		reqPath := c.Request.URL.Path

		// check if in skipUrls
		//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls(): reqPath:", reqPath, "   skipUrls:", jsonutil.MarshalJson(skipUrls),
		//	"   roleHasUrls:", jsonutil.MarshalJson(roleHasUrls))
		for _, skipUrl := range skipUrls {

			// if equal
			if skipUrl == reqPath {
				//	belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check skipUrl, skipUrl == reqPath, pass:", reqPath)
				return true
			} else if strings.HasSuffix(skipUrl, "*") {
				//if endwith, eg: /static/*
				reg := regexp.MustCompile(skipUrl).MatchString(reqPath)
				if reg {
					//	belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check skipUrl,roleUrl HasSuffix (*), skipUrl,reqPath, pass:", skipUrl, reqPath)
					return true
				}
			}
		}

		// check if role has urls
		ginUserModel := GinUserModel{}
		err := GetUserFromSession(c, &ginUserModel)
		//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():GetUserFromSession reqPath:", reqPath, "  ginUserModel:", jsonutil.MarshalJson(ginUserModel))
		if err != nil || ginUserModel.Id == 0 {
			belogs.Error("skipAuthUrlsOrRoleHasAuthUrls():get ginUserModel fail or ginUserModel.Id==0, reqPath:", reqPath, " , err:", err)
			return false
		}

		if len(roleHasUrls) == 0 {
			belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():len(roleHasUrls)==0, reqPath:", reqPath)
			return true
		}
		//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls(): check roleHasUrls, reqPath:", reqPath, "   roleHasUrls:", jsonutil.MarshalJson(roleHasUrls))
		roleUrls, ok := roleHasUrls[ginUserModel.RoleId]
		if !ok {
			belogs.Error("skipAuthUrlsOrRoleHasAuthUrls(): !ok, check roleUrls, reqPath:", reqPath, "  roleId:", ginUserModel.RoleId,
				"   roleUrls:", jsonutil.MarshalJson(roleUrls))
			return false
		}
		for _, roleUrl := range roleUrls {
			if roleUrl == reqPath {
				//		belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check roleUrls,roleUrl == reqPath, pass:", reqPath)
				return true
			} else if strings.HasSuffix(roleUrl, "*") {
				//if endwith, eg: /static/*
				reg := regexp.MustCompile(roleUrl).MatchString(reqPath)
				if reg {
					//			belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check roleUrls,roleUrl HasSuffix (*), roleUrl,reqPath, pass:", roleUrl, reqPath)
					return true
				}
			}
		}
		belogs.Info("skipAuthUrlsOrRoleHasAuthUrls():auth false, reqPath:", reqPath,
			"  ginUserModel.RoleId:", ginUserModel.RoleId, "   roleUrls:", roleUrls)
		return false
	}
}
