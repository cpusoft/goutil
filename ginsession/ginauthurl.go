package ginsession

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/ginserver"
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
type checkAuthUrlsFunc func(*gin.Context) checkAuthUrlsFuncResult
type checkAuthUrlsFuncResult struct {
	Result bool   `json:"result"`
	Method string `json:"method"`
}

func checkAuthUrls(redirectUrl string, failJson string,
	checkAuthUrlsFuncs ...checkAuthUrlsFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := checkAuthUrlsFuncResult{}
		funcLen := len(checkAuthUrlsFuncs)
		if funcLen > 0 {
			result = checkAuthUrlsFuncs[0](c)
		}
		if funcLen > 0 && result.Result {
			//belogs.Debug("checkAuthUrls(): checkAuthUrlsFuncs[0](c) pass: ", checkAuthUrlsFuncs[0], "   url:", c.Request.URL.Path)
			c.Next()
			return
		}

		//belogs.Debug("checkAuthUrls(): checkAuthUrlsFuncs[0](c) unpass: ", checkAuthUrlsFuncs[0],
		//	"    redirectUrl:", redirectUrl, "  or  failJson:", failJson)
		//c.Request.Method
		if result.Method == "GET" && len(redirectUrl) > 0 {
			belogs.Error("checkAuthUrls():will redirectUrl, result.Method:", result.Method, "   redirectUrl:", redirectUrl)
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		} else if result.Method == "POST" && len(failJson) > 0 {
			belogs.Error("checkAuthUrls():will ResponseFail, result.Method:", result.Method, "   failJson:", failJson)
			ginserver.ResponseFail(c, errors.New(failJson), nil)
		} else {
			belogs.Error("checkAuthUrls():default redirectUrl, result.Method:", result.Method, "   redirectUrl:", redirectUrl)
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
		}

		c.Abort()
		return
	}
}

// if the request path contains skip url(prefix), or the roles has the ruls, or roleHasUrls is empty, skip if it contains
func skipAuthUrlsOrRoleHasAuthUrls(skipUrls []string, roleHasUrls map[uint64][]string) checkAuthUrlsFunc {
	return func(c *gin.Context) checkAuthUrlsFuncResult {
		result := checkAuthUrlsFuncResult{}

		reqPath := c.Request.URL.Path
		result.Method = c.Request.Method
		// check if in skipUrls
		//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls(): reqPath:", reqPath, "   skipUrls:", jsonutil.MarshalJson(skipUrls),
		//	"   roleHasUrls:", jsonutil.MarshalJson(roleHasUrls))
		for _, skipUrl := range skipUrls {

			// if equal
			if skipUrl == reqPath {
				//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check skipUrl, skipUrl == reqPath, pass:", reqPath)
				result.Result = true
				return result
			} else if strings.HasSuffix(skipUrl, "*") {
				//if endwith, eg: /static/*
				reg := regexp.MustCompile(skipUrl).MatchString(reqPath)
				if reg {
					//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check skipUrl,roleUrl HasSuffix (*), skipUrl,reqPath, pass:", skipUrl, reqPath)
					result.Result = true
					return result
				}
			}
		}

		// check if role has urls
		ginUserModel := GinUserModel{}
		err := GetUserFromSession(c, &ginUserModel)
		//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():GetUserFromSession reqPath:", reqPath, "  ginUserModel:", jsonutil.MarshalJson(ginUserModel))
		if err != nil || ginUserModel.Id == 0 {
			belogs.Error("skipAuthUrlsOrRoleHasAuthUrls():get ginUserModel fail or ginUserModel.Id==0, reqPath:", reqPath, " , err:", err)
			result.Result = false
			return result
		}

		if len(roleHasUrls) == 0 {
			//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():len(roleHasUrls)==0, reqPath:", reqPath)
			result.Result = true
			return result
		}
		//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls(): check roleHasUrls, reqPath:", reqPath, "   roleHasUrls:", jsonutil.MarshalJson(roleHasUrls))
		roleUrls, ok := roleHasUrls[ginUserModel.RoleId]
		if !ok {
			belogs.Error("skipAuthUrlsOrRoleHasAuthUrls(): !ok, check roleUrls, reqPath:", reqPath,
				"  ginUserModel:", jsonutil.MarshalJson(ginUserModel), "   roleUrls:", jsonutil.MarshalJson(roleUrls))
			result.Result = false
			return result
		}
		for _, roleUrl := range roleUrls {
			if roleUrl == reqPath {
				//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check roleUrls,roleUrl == reqPath, pass:", reqPath)
				result.Result = true
				return result
			} else if strings.HasSuffix(roleUrl, "*") {
				//if endwith, eg: /static/*
				reg := regexp.MustCompile(roleUrl).MatchString(reqPath)
				if reg {
					//belogs.Debug("skipAuthUrlsOrRoleHasAuthUrls():check roleUrls,roleUrl HasSuffix (*), roleUrl,reqPath, pass:", roleUrl, reqPath)
					result.Result = true
					return result
				}
			}
		}
		belogs.Error("skipAuthUrlsOrRoleHasAuthUrls():auth false, reqPath:", reqPath,
			"  ginUserModel:", jsonutil.MarshalJson(ginUserModel), "   roleUrls:", jsonutil.MarshalJson(roleUrls))
		result.Result = false
		return result
	}
}
