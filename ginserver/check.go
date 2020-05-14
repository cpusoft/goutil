package ginserver

import (
	"errors"
	"strconv"
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// UserAuthMiddleware  User authorization middleware
func UserLoginMiddleware(userIdUrls map[uint64][]string, skipper ...SkipperFunc) gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(skipper) > 0 && skipper[0](c) {
			c.Next()
			return
		}
		var uuid string
		session := sessions.Default(c)
		if t := c.GetHeader(TOKEN_KEY); t != "" {
			userInfo, ok := ParseToken(t)
			if !ok {
				ResponseFail(c, errors.New("token invalid"), nil)
				return
			}
			exptimestamp, _ := strconv.ParseInt(userInfo["exp"], 10, 64)
			exp := time.Unix(exptimestamp, 0)
			ok = exp.After(time.Now())
			if !ok {
				ResponseFail(c, errors.New("token expired"), nil)
				return
			}
			uuid = userInfo["uuid"]
		}
		if uuid != "" {
			belogs.Debug("uuid:", uuid)
			belogs.Debug("uuid session:", session.Get(uuid))

			userid := session.Get(uuid)
			if userid == "" {
				ResponseFail(c, errors.New("User is not login"), nil)
				return
			}
			uid := userid.(uint64)
			c.Set(USER_UUID_Key, uuid)
			c.Set(USER_ID_Key, uid)
			var SessionContent SessionContent
			Urls := userIdUrls[uid] //db.QueryUserAuthValidById(uid)
			SessionContent.Urls = Urls
			JsonUrls := jsonutil.MarshalJson(SessionContent)
			authUrl := string(strconv.Itoa(int(uid))) + AUTH_URL
			if session.Get(authUrl) == nil {
				session.Set(authUrl, JsonUrls)
				belogs.Debug("JsonUrls:", JsonUrls)
				session.Save()
			}
		}
		if uuid == "" {
			ResponseFail(c, errors.New("User is not login"), nil)
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
			ResponseFail(c, errors.New("User has not been assigned permissions"), nil)
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
			ResponseFail(c, errors.New("No access"), nil)
			return
		}
		c.Next()
	}
}
