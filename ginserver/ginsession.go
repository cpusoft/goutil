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
	if len(key) == 0 || value == nil {
		return nil
	}
	s := sessions.Default(c)
	s.Set(key, jsonutil.MarshalJson(value))
	return s.Save()
}
func SaveUserToSession(c *gin.Context, ginUserModel *GinUserModel) error {
	return SaveToSession(c, "ginuser", ginUserModel)
}
func GetFromSession(c *gin.Context, key string, value interface{}) error {
	if len(key) == 0 {
		return nil
	}

	s := sessions.Default(c)
	if s == nil {
		return nil
	}
	v := s.Get(key)
	if v == nil {
		return nil
	}
	return jsonutil.UnmarshalJson(v.(string), &value)
}
func GetUserFromSession(c *gin.Context, ginUserModel *GinUserModel) error {
	return GetFromSession(c, "ginuser", ginUserModel)
}
func DeleteInSession(c *gin.Context, key string) error {
	if len(key) == 0 {
		return nil
	}

	s := sessions.Default(c)
	s.Delete(key)
	return s.Save()
}

func DeleteUserInSession(c *gin.Context) error {
	return DeleteInSession(c, "ginuser")
}
