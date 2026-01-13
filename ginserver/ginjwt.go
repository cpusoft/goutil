package ginserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/jwtutil"
	"github.com/gin-gonic/gin"
)

const (
	JWT_HEADER_AUTHORIZATION = "Authorization"
	JWT_HEADER_PREFIX_BEARER = "Bearer"
	// same as in jwtutil.go
	// same as in zaplogs.go
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
			ResponseFail(c, errors.New("Authorization is empty"), nil)
			c.Abort()
			return
		}
		belogs.Debug("jwtAuthMiddleware(): authHeader:", authHeader)

		// 验证Authorization格式（Bearer <token>）
		var tokenString string
		_, err := fmt.Sscanf(authHeader, JWT_HEADER_PREFIX_BEARER+" %s", &tokenString)
		if err != nil {
			belogs.Error("jwtAuthMiddleware(): Sscanf tokenString fail, authHeader:", authHeader, err)
			ResponseFail(c, err, nil)
			c.Abort()
			return
		}
		belogs.Debug("jwtAuthMiddleware(): tokenString:", tokenString)

		customClaims, err := jwtutil.ParseToken(tokenString, conf.String("jwt::secret"))
		if err != nil {
			belogs.Error("jwtAuthMiddleware(): ParseToken fail, tokenString:", tokenString, err)
			ResponseFail(c, err, nil)
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
		return context.Background()
	}
	belogs.Debug("SetToContextWithValue(): get JWT_CTX_CustomClaims_Infos", "m", jsonutil.MarshalJson(m))
	authHeader := c.GetHeader(JWT_HEADER_AUTHORIZATION)
	ctx := context.WithValue(context.Background(), JWT_CTX_CustomClaims_Infos, m)
	ctx = context.WithValue(ctx, JWT_HEADER_AUTHORIZATION, authHeader)
	return ctx
}

func GetCustomClaims(ctx context.Context) map[string]interface{} {
	cc := ctx.Value(JWT_CTX_CustomClaims_Infos)
	if cc == nil {
		belogs.Error("GetCustomClaims(): get JWT_CTX_CustomClaims_Infos from gin.Context fail",
			"JWT_CTX_CustomClaims_Infos:", JWT_CTX_CustomClaims_Infos)
		return make(map[string]interface{})
	}
	m, ok := cc.(map[string]interface{})
	if !ok {
		belogs.Error("GetCustomClaims(): assert to CustomClaims fail, cc:", jsonutil.MarshalJson(cc))
		return make(map[string]interface{})
	}
	return m
}

func GetAuthHeader(ctx context.Context) string {
	authHeader := ctx.Value(JWT_HEADER_AUTHORIZATION)
	if authHeader == nil {
		belogs.Error("GetAuthHeader(): get JWT_HEADER_AUTHORIZATION from gin.Context fail",
			"JWT_HEADER_AUTHORIZATION", JWT_HEADER_AUTHORIZATION)
		return ""
	}
	authHeaderStr, ok := authHeader.(string)
	if !ok {
		belogs.Error("GetAuthHeader(): assert to string fail", "authHeader", jsonutil.MarshalJson(authHeader))
		return ""
	}
	return authHeaderStr
}
