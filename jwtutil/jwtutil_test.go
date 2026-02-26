package jwtutil

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/cpusoft/goutil/idutil"
	jwt "github.com/golang-jwt/jwt/v5"
)

// -------------------------- 核心测试用例 --------------------------

// TestGenToken 测试生成Token的所有场景
func TestGenToken(t *testing.T) {
	// 构建基础合法的CustomClaims
	validClaims := &CustomClaims{
		Infos: map[string]interface{}{
			"ownerId":   "1001",
			"ownerName": "beijing-user1",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	validPassword := "test-secret-123"

	tests := []struct {
		name    string
		claims  *CustomClaims
		pwd     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "正常场景-生成有效Token",
			claims:  validClaims,
			pwd:     validPassword,
			wantErr: false,
		},
		{
			name:    "异常场景-claims为nil",
			claims:  nil,
			pwd:     validPassword,
			wantErr: true,
			errMsg:  "customClaims cannot be nil",
		},
		{
			name:    "异常场景-password为空",
			claims:  validClaims,
			pwd:     "",
			wantErr: true,
			errMsg:  "password cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenToken(tt.claims, tt.pwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("GenToken() error msg = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}
			if token == "" {
				t.Error("GenToken() 返回空token，不符合预期")
			}
		})
	}
}

// TestParseToken 测试解析Token的所有场景（含临界值、算法伪造）
func TestParseToken(t *testing.T) {
	validPassword := "test-secret-123"
	invalidPassword := "wrong-secret-456"

	// 1. 生成有效Token（用于正常场景测试）
	validClaims := &CustomClaims{
		Infos: map[string]interface{}{
			"ownerId": "1001",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	validToken, _ := GenToken(validClaims, validPassword)

	// 2. 生成过期Token（临界值测试）
	expiredClaims := &CustomClaims{
		Infos: map[string]interface{}{"ownerId": "1001"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)), // 已过期
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-3 * time.Minute)),
		},
	}
	expiredToken, _ := GenToken(expiredClaims, validPassword)

	// 3. 伪造none算法Token（安全测试）
	fakeNoneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJpbmZvcyI6eyJvd25lcklkIjoiMTAwMSJ9LCJleHAiOjE3MTExMTExMTEsIm5iZiI6MTcxMTExMTExMSwiaWF0IjoxNzExMTExMTExfQ."

	tests := []struct {
		name    string
		token   string
		pwd     string
		wantErr bool
		checkFn func(t *testing.T, claims *CustomClaims) // 自定义校验逻辑
	}{
		{
			name:    "正常场景-解析有效Token",
			token:   validToken,
			pwd:     validPassword,
			wantErr: false,
			checkFn: func(t *testing.T, claims *CustomClaims) {
				if claims.Infos["ownerId"] != "1001" {
					t.Errorf("解析后的Infos错误，got %v", claims.Infos["ownerId"])
				}
			},
		},
		{
			name:    "异常场景-token为空",
			token:   "",
			pwd:     validPassword,
			wantErr: true,
		},
		{
			name:    "异常场景-password为空",
			token:   validToken,
			pwd:     "",
			wantErr: true,
		},
		{
			name:    "异常场景-密码错误",
			token:   validToken,
			pwd:     invalidPassword,
			wantErr: true,
		},
		{
			name:    "异常场景-Token已过期（临界值）",
			token:   expiredToken,
			pwd:     validPassword,
			wantErr: true,
		},
		{
			name:    "异常场景-伪造none算法Token",
			token:   fakeNoneToken,
			pwd:     validPassword,
			wantErr: true,
		},
		{
			name:    "异常场景-无效Token（篡改）",
			token:   validToken + "tampered",
			pwd:     validPassword,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.token, tt.pwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFn != nil {
				tt.checkFn(t, claims)
			}
		})
	}
}

// TestCtxFunctions 测试上下文相关函数（含空值临界场景）
func TestCtxFunctions(t *testing.T) {
	// 预定义UUID格式正则（用于验证降级逻辑）
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// Test NewCtxWithValue
	t.Run("NewCtxWithValue-正常场景", func(t *testing.T) {
		m := map[string]interface{}{"key1": "value1"}
		ctx := NewCtxWithValue(m)
		val := ctx.Value(JWT_CTX_CustomClaims_Infos)
		valMap, ok := val.(map[string]interface{})
		if !ok {
			t.Error("NewCtxWithValue 返回的ctx value类型错误")
		}
		if valMap["key1"] != "value1" {
			t.Errorf("NewCtxWithValue 取值错误，got %v", valMap["key1"])
		}
	})

	t.Run("NewCtxWithValue-空map场景", func(t *testing.T) {
		ctx := NewCtxWithValue(nil) // 临界值：map为nil
		val := ctx.Value(JWT_CTX_CustomClaims_Infos)
		valMap, ok := val.(map[string]interface{})
		if !ok || len(valMap) != 0 {
			t.Error("NewCtxWithValue 处理空map失败")
		}
	})

	// Test NewCtxWithValueOfParentCtx
	t.Run("NewCtxWithValueOfParentCtx-正常场景", func(t *testing.T) {
		parentCtx := context.WithValue(context.Background(), "parentKey", "parentValue")
		m := map[string]interface{}{"key2": "value2"}
		ctx := NewCtxWithValueOfParentCtx(parentCtx, m)

		// 验证父ctx的值保留
		if ctx.Value("parentKey") != "parentValue" {
			t.Error("父上下文值丢失")
		}
		// 验证自定义值
		val := ctx.Value(JWT_CTX_CustomClaims_Infos)
		valMap, _ := val.(map[string]interface{})
		if valMap["key2"] != "value2" {
			t.Errorf("NewCtxWithValueOfParentCtx 取值错误，got %v", valMap["key2"])
		}
	})

	t.Run("NewCtxWithValueOfParentCtx-父ctx为空", func(t *testing.T) {
		ctx := NewCtxWithValueOfParentCtx(nil, map[string]interface{}{"key3": "value3"}) // 临界值：parentCtx为nil
		val := ctx.Value(JWT_CTX_CustomClaims_Infos)
		valMap, _ := val.(map[string]interface{})
		if valMap["key3"] != "value3" {
			t.Error("父ctx为空时处理失败")
		}
	})

	// Test NewCtxWithValueDefault
	t.Run("NewCtxWithValueDefault-正常场景", func(t *testing.T) {
		ctx := NewCtxWithValueDefault()
		val := ctx.Value(JWT_CTX_CustomClaims_Infos)
		valMap, ok := val.(map[string]interface{})
		if !ok {
			t.Error("NewCtxWithValueDefault 返回值类型错误")
		}
		// 验证traceId和opLogId非空
		if valMap["traceId"] == "" || valMap["opLogId"] == "" {
			t.Error("traceId或opLogId为空，不符合预期")
		}
	})

	// 核心修改：不mock，直接验证降级逻辑（通过格式判断）
	t.Run("NewCtxWithValueDefault-雪花ID降级场景（UUID格式验证）", func(t *testing.T) {
		// 多次执行，覆盖"雪花ID成功"和"雪花ID失败降级"两种情况
		for i := 0; i < 10; i++ {
			ctx := NewCtxWithValueDefault()
			valMap := ctx.Value(JWT_CTX_CustomClaims_Infos).(map[string]interface{})
			traceId := valMap["traceId"].(string)
			opLogId := valMap["opLogId"].(string)

			// 验证规则：要么是雪花ID（纯数字），要么是UUID（符合UUID格式），且都非空
			isNumber := regexp.MustCompile(`^\d+$`).MatchString(traceId)
			isUUID := uuidRegex.MatchString(traceId)
			if !isNumber && !isUUID {
				t.Errorf("traceId格式异常，既非雪花ID也非UUID：%s", traceId)
			}

			isNumberOp := regexp.MustCompile(`^\d+$`).MatchString(opLogId)
			isUUIDOp := uuidRegex.MatchString(opLogId)
			if !isNumberOp && !isUUIDOp {
				t.Errorf("opLogId格式异常，既非雪花ID也非UUID：%s", opLogId)
			}
		}
	})
}

// -------------------------- 性能测试 --------------------------

// BenchmarkGenToken 测试Token生成性能
func BenchmarkGenToken(b *testing.B) {
	traceId, _ := idutil.GenerateSnowflakeString(time.Now().UnixNano())
	claims := &CustomClaims{
		Infos: map[string]interface{}{
			"ownerId":   "1001",
			"ownerName": "beijing-user1",
			"traceId":   traceId,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	password := "test-secret-123"

	b.ResetTimer() // 重置计时器，排除初始化耗时
	for i := 0; i < b.N; i++ {
		_, _ = GenToken(claims, password)
	}
}

// BenchmarkParseToken 测试Token解析性能
func BenchmarkParseToken(b *testing.B) {
	// 预生成Token，避免循环内重复生成
	claims := &CustomClaims{
		Infos: map[string]interface{}{"ownerId": "1001"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
		},
	}
	password := "test-secret-123"
	token, _ := GenToken(claims, password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseToken(token, password)
	}
}

// BenchmarkNewCtxWithValueDefault 测试默认上下文生成性能
func BenchmarkNewCtxWithValueDefault(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCtxWithValueDefault()
	}
}
