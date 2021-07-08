package ginsession

import (
	"errors"
	"strconv"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	TOKEN_KEY     = "X-Token"  //Page token key name
	USER_ID_Key   = "X-USERID" //Page user ID key name
	USER_UUID_Key = "X-UUID"   //UUID key name
	AUTH_URL      = "authurl"  //session key name
)

type SessionContent struct {
	Urls  []string `json:"urls"`
	Uname string   `json:"uname"`
}

// UserAuthMiddleware  User authorization middleware
func UserLoginMiddleware(userIdUrls map[uint64][]string, skipper ...SkipperFunc) gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(skipper) > 0 && skipper[0](c) {
			c.Next()
			return
		}
		var uuid string
		if t := c.GetHeader(TOKEN_KEY); t != "" {
			userInfo, ok := ParseToken(t)
			if !ok {
				ginserver.ResponseFail(c, errors.New("token invalid"), nil)
				c.Abort()
				return
			}
			exptimestamp, _ := strconv.ParseInt(userInfo["exp"], 10, 64)
			exp := time.Unix(exptimestamp, 0)
			ok = exp.After(time.Now())
			if !ok {
				ginserver.ResponseFail(c, errors.New("token expired"), nil)
				c.Abort()
				return
			}
			uuid = userInfo["uuid"]
		}

		if uuid != "" {
			belogs.Debug("UserLoginMiddleware():uuid:", uuid)

			session := sessions.Default(c)
			userId := session.Get(uuid)
			belogs.Debug("uuid session:", session.Get(uuid))
			if userId == "" {
				ginserver.ResponseFail(c, errors.New("User is not login"), nil)
				c.Abort()
				return
			}
			uid := userId.(uint64)
			c.Set(USER_UUID_Key, uuid)
			c.Set(USER_ID_Key, uid)

			var SessionContent SessionContent
			Urls := userIdUrls[uid] //db.QueryUserAuthValidById(uid)
			SessionContent.Urls = Urls
			JsonUrls := jsonutil.MarshalJson(SessionContent)
			authUrl := string(strconv.Itoa(int(uid))) + AUTH_URL
			if session.Get(authUrl) == nil {
				session.Set(authUrl, JsonUrls)
				//belogs.Debug("JsonUrls:", JsonUrls)
				session.Save()
			}
		}
		if uuid == "" {
			ginserver.ResponseFail(c, errors.New("User is not login"), nil)
			c.Abort()
			return
		}
	}
}

func UserAuthMiddleware(skipper ...SkipperFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(skipper) > 0 && skipper[0](c) {
			c.Next()
			return
		}

		session := sessions.Default(c)
		var SessionContent SessionContent
		uid, _ := c.Get(USER_ID_Key)

		authUrl := string(strconv.Itoa(int(uid.(uint64)))) + AUTH_URL
		if session.Get(authUrl) == nil {
			ginserver.ResponseFail(c, errors.New("User has not been assigned permissions"), nil)
			c.Abort()
			return
		}
		authcontent := session.Get(authUrl).(string)
		jsonutil.UnmarshalJson((authcontent), &SessionContent)
		belogs.Debug("Access path", SessionContent.Urls)
		p := c.Request.URL.Path
		IsAuth := true
		for _, value := range SessionContent.Urls {
			if p == value {
				IsAuth = false
			}
		}
		if IsAuth {
			ginserver.ResponseFail(c, errors.New("No access"), nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
