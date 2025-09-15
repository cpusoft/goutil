package jwtutil

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	jwt "github.com/golang-jwt/jwt/v5"
)

// same as in zaplogs.go
type CustomClaims struct {
	// [usrId]=***,[userName]=***,[ownerId]=***
	UserInfos map[string]string `json:"userInfos"`
	// [opLogId]=***
	OpInfos              map[string]string `json:"opInfos"`
	TraceId              string            `json:"traceId"`
	jwt.RegisteredClaims                   // 内嵌标准的声明
}

/*
claims := CustomJwtClaims{
		UserID:   1001,
		Username: "alice_smith",
		Role:     "editor",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "my_application",       // 签发者
			Subject:   "user_authentication",  // 主题
			Audience:  jwt.ClaimStrings{"app"}, // 受众
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)), // 过期时间
			NotBefore: jwt.NewNumericDate(time.Now()),      // 生效时间（立即生效）
			IssuedAt:  jwt.NewNumericDate(time.Now()),      // 签发时间
			ID:        "unique-token-id-12345",             // 令牌ID
		},
	}

*/
// GenToken
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
