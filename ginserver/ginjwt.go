package ginserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/jwtutil"
	"github.com/gin-gonic/gin"
)

// 自定义Context Key类型，避免命名冲突
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
			// 模糊错误信息，避免泄露实现细节
			ResponseFail(c, errors.New("authentication failed"), nil)
			c.Abort()
			return
		}
		belogs.Debug("jwtAuthMiddleware(): authHeader:", authHeader)

		// 健壮解析Bearer Token：忽略大小写、处理多余空格
		parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != strings.ToLower(JWT_HEADER_PREFIX_BEARER) {
			belogs.Error("jwtAuthMiddleware(): invalid Authorization format, authHeader:", authHeader)
			// 模糊错误信息
			ResponseFail(c, errors.New("authentication failed"), nil)
			c.Abort()
			return
		}
		tokenString := parts[1]
		belogs.Debug("jwtAuthMiddleware(): tokenString:", tokenString)

		// 校验JWT密钥有效性，避免空密钥
		jwtSecret := conf.String("jwt::secret")
		if jwtSecret == "" {
			belogs.Error("jwtAuthMiddleware(): JWT secret is empty, config key: jwt::secret")
			ResponseFail(c, errors.New("authentication failed"), nil)
			c.Abort()
			return
		}

		customClaims, err := jwtutil.ParseToken(tokenString, jwtSecret)
		if err != nil {
			belogs.Error("jwtAuthMiddleware(): ParseToken fail, tokenString:", tokenString, err)
			// 模糊错误信息，不暴露Token解析的具体失败原因
			ResponseFail(c, errors.New("authentication failed"), nil)
			c.Abort()
			return
		}
		belogs.Debug("jwtAuthMiddleware(): customClaims:", jsonutil.MarshalJson(customClaims))
		c.Set(string(JWT_CTX_CustomClaims_Infos), customClaims.Infos)
		c.Next()
	}
}

func SetToContextWithValue(c *gin.Context) context.Context {
	m, exists := c.Get(string(JWT_CTX_CustomClaims_Infos))
	if !exists {
		belogs.Error("SetToContextWithValue(): get JWT_CTX_CustomClaims_Infos from gin.Context fail, JWT_CTX_CustomClaims_Infos:", JWT_CTX_CustomClaims_Infos)
		// 使用请求上下文作为父上下文，而非空上下文
		return context.Background()
	}
	// 日志仅输出关键标识，不泄露完整Claims内容
	belogs.Debug("SetToContextWithValue(): get JWT_CTX_CustomClaims_Infos success, data exists")

	authHeader := c.GetHeader(JWT_HEADER_AUTHORIZATION)
	belogs.Debug("SetToContextWithValue(): get Authorization header", "value", authHeader)

	// 正确设置Context值（链式传递，避免覆盖）
	// 核心：基于同一个Background()，依次设置两个Key
	ctx := context.WithValue(context.Background(), JWT_CTX_CustomClaims_Infos, m)
	ctx = context.WithValue(ctx, JWT_HEADER_AUTHORIZATION, authHeader)
	return ctx
}

func GetCustomClaims(ctx context.Context) map[string]interface{} {
	cc := ctx.Value(JWT_CTX_CustomClaims_Infos)
	if cc == nil {
		belogs.Error("GetCustomClaims(): get JWT_CTX_CustomClaims_Infos from context fail",
			"JWT_CTX_CustomClaims_Infos:", JWT_CTX_CustomClaims_Infos)
		return make(map[string]interface{})
	}
	m, ok := cc.(map[string]interface{})
	if !ok {
		// 日志不输出完整cc内容，避免泄露敏感信息
		belogs.Error("GetCustomClaims(): assert to map[string]interface{} fail, type:", fmt.Sprintf("%T", cc))
		return make(map[string]interface{})
	}
	return m
}

func GetAuthHeader(ctx context.Context) string {
	authHeader := ctx.Value(JWT_HEADER_AUTHORIZATION)
	if authHeader == nil {
		belogs.Error("GetAuthHeader(): get JWT_HEADER_AUTHORIZATION from context fail",
			"JWT_HEADER_AUTHORIZATION", JWT_HEADER_AUTHORIZATION)
		return ""
	}
	authHeaderStr, ok := authHeader.(string)
	if !ok {
		// 日志仅输出类型，不输出完整内容
		belogs.Error("GetAuthHeader(): assert to string fail, type:", fmt.Sprintf("%T", authHeader))
		return ""
	}
	return authHeaderStr
}
