package ginserver

import (
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/uuidutil"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

// sessionId wil save to cookis, session value will save to memory ;
// maxAge(seconds): 30 * 60 ;
// secure: only use cookie on https ;
// httpOnly: only in http, cannot in js ;
func InitSessionInMem(engine *gin.Engine, cookieName string, maxAge int, secure, httpOnly bool) {

	memStore := memstore.NewStore([]byte(uuidutil.GetUuid()))
	memStore.Options(sessions.Options{
		Path:     "/",
		Domain:   "/",
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
	engine.Use(sessions.Sessions(cookieName, memStore))

}

func SaveToSession(c *gin.Context, key string, value interface{}) error {
	s := sessions.Default(c)
	s.Set(key, jsonutil.MarshalJson(value))
	return s.Save()
}
func SaveUserToSession(c *gin.Context, ginUserModel *GinUserModel) error {
	return SaveToSession(c, "ginuser", ginUserModel)
}
func GetFromSession(c *gin.Context, key string, value interface{}) error {
	s := sessions.Default(c)
	v := s.Get(key)
	return jsonutil.UnmarshalJson(v.(string), &value)
}
func GetUserFromSession(c *gin.Context, ginUserModel *GinUserModel) error {
	return GetFromSession(c, "ginuser", ginUserModel)
}
func DeleteInSession(c *gin.Context, key string) {
	s := sessions.Default(c)
	s.Delete(key)
	s.Save()
}
