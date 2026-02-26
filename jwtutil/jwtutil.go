package jwtutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/idutil"
	"github.com/cpusoft/goutil/uuidutil"
	jwt "github.com/golang-jwt/jwt/v5"
)

// same as in jwtutil.go
// same as in zaplogs.go
const JWT_CTX_CustomClaims_Infos = "CustomClaims.Infos"

/*
Example:

	 m:=make(map[string]string)
	 m["ownerId"]="1001"
	 m["ownerName"]="beijing-user1"
	 m["opUserId"]="2002"
	 m["opUserName"]="beijing-user2"
	 m["traceId"]="550e8400-e29b-41d4-a716-446655440000"
	 m["opLogId"]="3003"

		claims := CustomJwtClaims{
				Infos:   m,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)), //过期时间
					NotBefore: jwt.NewNumericDate(time.Now()),//生效时间（立即生效）
					IssuedAt:  jwt.NewNumericDate(time.Now()),//签发时间
				},
			}
*/
// same as in zaplogs.go
type CustomClaims struct {
	Infos                map[string]interface{} `json:"infos,omitempty"` // 自定义信息
	jwt.RegisteredClaims                        // 内嵌标准的声明
}

// password: 签名密码
func GenToken(customClaims *CustomClaims, password string) (string, error) {
	// 增加入参校验，避免空指针
	if customClaims == nil {
		belogs.Error("GenToken(): customClaims is nil")
		return "", errors.New("customClaims cannot be nil")
	}
	if password == "" {
		belogs.Error("GenToken(): password is empty")
		return "", errors.New("password cannot be empty")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return token.SignedString([]byte(password))
}

func ParseToken(token string, password string) (*CustomClaims, error) {
	// 入参基础校验
	if token == "" {
		belogs.Error("ParseToken(): token is empty")
		return nil, errors.New("token cannot be empty")
	}
	if password == "" {
		belogs.Error("ParseToken(): password is empty")
		return nil, errors.New("password cannot be empty")
	}

	cTmp := &CustomClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		cTmp,
		func(token *jwt.Token) (interface{}, error) {
			// 核心修复：校验签名算法是否为预期的HS256，防止none算法伪造
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				errMsg := fmt.Sprintf("ParseToken(): unexpected signing method: %v", token.Header["alg"])
				belogs.Error("ParseToken(): ParseWithClaims token fail", errMsg)
				return nil, errors.New(errMsg)
			}
			return []byte(password), nil
		},
	)

	if err != nil {
		belogs.Error("ParseToken(): ParseWithClaims fail:", err)
		return nil, err
	}
	claims, ok := parsedToken.Claims.(*CustomClaims)
	if !ok {
		errMsg := "ParseToken(): claims type assertion failed"
		belogs.Error(errMsg, "ok:", ok)
		return nil, errors.New(errMsg)
	}

	if !parsedToken.Valid {
		belogs.Error("ParseToken(): is not valid:", parsedToken)
		return nil, errors.New("token is invalid")
	}
	return claims, nil
}

// m := make(map[string]interface{})
// m["traceId"] = "ginserver-background-traceId"
// m["opLogId"] = "ginserver-background-opLogId"
func NewCtxWithValue(m map[string]interface{}) context.Context {
	if m == nil {
		m = make(map[string]interface{})
	}
	return context.WithValue(context.Background(), JWT_CTX_CustomClaims_Infos, m)
}
func NewCtxWithValueOfParentCtx(parentCtx context.Context, m map[string]interface{}) context.Context {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	// 防止空map导致后续取值panic
	if m == nil {
		m = make(map[string]interface{})
	}
	return context.WithValue(parentCtx, JWT_CTX_CustomClaims_Infos, m)
}
func NewCtxWithValueDefault() context.Context {
	m := make(map[string]interface{})
	// 修复：处理Snowflake ID生成错误，避免空值
	traceId, err := idutil.GenerateSnowflakeString(time.Now().UnixNano())
	if err != nil {
		belogs.Warn("NewCtxWithValueDefault(): generate traceId fail:", err)
		// 降级方案：使用随机ID保证非空
		traceId = uuidutil.GetUuid() // 生成一个随机UUID作为traceId，确保不为空
	}
	m["traceId"] = traceId

	opLogId, err := idutil.GenerateSnowflakeString(time.Now().UnixNano() + 1)
	if err != nil {
		belogs.Warn("NewCtxWithValueDefault(): generate opLogId fail:", err)
		opLogId = uuidutil.GetUuid()
	}
	m["opLogId"] = opLogId
	return context.WithValue(context.Background(), JWT_CTX_CustomClaims_Infos, m)
}
