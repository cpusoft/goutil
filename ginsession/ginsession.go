package ginsession

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// sessionId wil save to cookis, session value will save to memory ;//
// domain: allowed to w/r cookie, must start by '.' : .aaa.com  //
// maxAge(seconds): 30 * 60 ;   //
// secure: only use cookie on https ;  //
// httpOnly: only in http, cannot in js ; //
func InitSessionInMem(engine *gin.Engine, domain, cookieName string, maxAge int, secure, httpOnly bool) {

	store := cookie.NewStore([]byte("goutil_ginserver_cookie"))
	store.Options(sessions.Options{
		Path:     "/",
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
	engine.Use(sessions.Sessions(cookieName, store))

}

func SaveToSession(c *gin.Context, key string, value interface{}) error {
	if len(key) == 0 || value == nil {
		return nil
	}
	s := sessions.Default(c)
	belogs.Debug("SaveToSession(): key:", key, " jsonutil.MarshalJson(value):", jsonutil.MarshalJson(value))
	s.Set(key, jsonutil.MarshalJson(value))
	return s.Save()
}
func SaveUserToSession(c *gin.Context, ginUserModel *GinUserModel) error {
	err := SaveToSession(c, "ginuser", ginUserModel)

	ginUserModelTest := GinUserModel{}
	err1 := GetUserFromSession(c, &ginUserModelTest)
	belogs.Debug("SaveUserToSession(): save to session: ", jsonutil.MarshalJson(ginUserModel),
		"   get from session: ", jsonutil.MarshalJson(ginUserModelTest), err1)

	return err
}
func GetFromSession(c *gin.Context, key string, value interface{}) error {
	if len(key) == 0 {
		return nil
	}

	s := sessions.Default(c)
	//belogs.Debug("GetFromSession(): key:", key, "   c:", c, " s:", jsonutil.MarshalJson(&s))
	if s == nil {
		return nil
	}
	v := s.Get(key)
	//belogs.Debug("GetFromSession(): key:", key, "   c:", c, " s:", jsonutil.MarshalJson(&s), " v:", v)
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
