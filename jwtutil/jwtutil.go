package jwtutil

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	jwt "github.com/golang-jwt/jwt/v5"
)

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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return token.SignedString([]byte(password))
}

func ParseToken(token string, password string) (*CustomClaims, error) {
	cTmp := &CustomClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		cTmp,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(password), nil
		},
	)

	if err != nil {
		belogs.Error("ParseToken(): ParseWithClaims fail:", err)
		return nil, err
	}
	claims := &CustomClaims{}
	var ok bool
	if claims, ok = parsedToken.Claims.(*CustomClaims); !ok {
		belogs.Error("ParseToken(): is not ok, or is not valid:", ok)
		return nil, err
	}

	if !parsedToken.Valid {
		belogs.Error("ParseToken(): is not valid:", parsedToken)
		return nil, errors.New("token is invalid")
	}
	return claims, nil
}
