package ginsession

import (
	"context"
	"errors"
	"fmt"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/jwtutil"
	"github.com/gin-gonic/gin"
)

const (
	JWT_HEADER_AUTHORIZATION   = "Authorization"
	JWT_HEADER_PREFIX_BEARER   = "Bearer"
	JWT_CTX_CustomClaims_Infos = "CustomClaims.Infos"
)

// ginsession.RegisterJwt(engine)
func EngineRegisterJwt(engine *gin.Engine) {
	engine.Use(jwtAuthMiddleware())
}
func RouterGroupRegisterJwt(group *gin.RouterGroup) {
	group.Use(jwtAuthMiddleware())
}

// JWT中间件：验证令牌并将用户信息存入上下文
func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取Authorization字段
		authHeader := c.GetHeader(JWT_HEADER_AUTHORIZATION)
		if authHeader == "" {
			ginserver.ResponseFail(c, errors.New("Authorization is empty"), nil)
			c.Abort()
			return
		}
		belogs.Debug("jwtAuthMiddleware(): authHeader:", authHeader)

		// 验证Authorization格式（Bearer <token>）
		var tokenString string
		_, err := fmt.Sscanf(authHeader, JWT_HEADER_PREFIX_BEARER+" %s", &tokenString)
		if err != nil {
			belogs.Error("jwtAuthMiddleware(): Sscanf tokenString fail, authHeader:", authHeader, err)
			ginserver.ResponseFail(c, err, nil)
			c.Abort()
			return
		}
		belogs.Debug("jwtAuthMiddleware(): tokenString:", tokenString)

		customClaims, err := jwtutil.ParseToken(tokenString, conf.String("jwt::secret"))
		if err != nil {
			belogs.Error("jwtAuthMiddleware(): ParseToken fail, tokenString:", tokenString, err)
			ginserver.ResponseFail(c, err, nil)
			c.Abort()
			return
		}
		belogs.Debug("jwtAuthMiddleware(): customClaims:", jsonutil.MarshalJson(customClaims))
		c.Set(JWT_CTX_CustomClaims_Infos, customClaims.Infos)
		c.Next()
	}
}

func SetToContextWithValue(c *gin.Context) context.Context {
	m, exists := c.Get(JWT_CTX_CustomClaims_Infos)
	if !exists {
		belogs.Error("SetToContextWithValue(): get JWT_CTX_CustomClaims_Infos from gin.Context fail, JWT_CTX_CustomClaims_Infos:", JWT_CTX_CustomClaims_Infos)
		return nil
	}
	belogs.Debug("SetToContextWithValue(): get JWT_CTX_CustomClaims_Infos", "m", jsonutil.MarshalJson(m))
	return context.WithValue(context.Background(), JWT_CTX_CustomClaims_Infos, m)
}

func GetCustomClaims(ctx context.Context) map[string]string {
	cc := ctx.Value(JWT_CTX_CustomClaims_Infos)
	if cc == nil {
		belogs.Error("GetCustomClaims(): get JWT_CTX_CustomClaims_Infos from gin.Context fail, JWT_CTX_CustomClaims_Infos:", JWT_CTX_CustomClaims_Infos)
		return make(map[string]string)
	}
	m, ok := cc.(map[string]string)
	if !ok {
		belogs.Error("GetCustomClaims(): assert to CustomClaims fail, cc:", jsonutil.MarshalJson(cc))
		return make(map[string]string)
	}
	return m
}
