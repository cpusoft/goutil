package ginserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// ======================== 全局初始化 ========================
func TestMain(m *testing.M) {
	// 初始化：禁用gin调试模式、配置日志输出
	gin.SetMode(gin.TestMode)

	// 运行测试
	exitCode := m.Run()

	// 退出测试
	os.Exit(exitCode)
}

// generateTestJwtToken 生成测试用的有效JWT Token
func generateTestJwtToken(t *testing.T) string {
	// 构造测试Claims
	claims := map[string]interface{}{
		"uid":  1001,
		"name": "test-user",
		"role": "admin",
	}

	customClaims := &jwtutil.CustomClaims{
		Infos: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)), //过期时间
			NotBefore: jwt.NewNumericDate(time.Now()),                    //生效时间（立即生效）
			IssuedAt:  jwt.NewNumericDate(time.Now()),                    //签发时间
			Issuer:    "test-issuer",
			Subject:   "test-subject",
			ID:        "550e8400-e29b-41d4-a716-446655440000",
		},
	}
	// 生成Token（有效期1小时）
	token, err := jwtutil.GenToken(customClaims, conf.DefaultString("jwt::secret", "test-jwt-secret-123456"))
	assert.NoError(t, err, "生成测试JWT Token失败")
	return token
}

// generateInvalidJwtToken 生成测试用的无效JWT Token（篡改签名）
func generateInvalidJwtToken(t *testing.T) string {
	// 先用正确密钥生成Token，再用错误密钥重新签名（模拟篡改）
	validToken := generateTestJwtToken(t)
	// 解析Token后用错误密钥重新生成
	claims, err := jwtutil.ParseToken(validToken, conf.DefaultString("jwt::secret", "test-jwt-secret-123456"))
	assert.NoError(t, err)
	invalidToken, err := jwtutil.GenToken(claims, "wrong-secret")
	assert.NoError(t, err)
	return invalidToken
}

// getTestJwtSecret 获取测试用的JWT密钥（兼容conf包无Set方法的场景）
func getTestJwtSecret() string {
	// 优先从配置读取，读取不到则用测试默认值
	secret := conf.String("jwt::secret")
	if secret == "" {
		secret = "test-jwt-secret-123456" // 测试兜底密钥
	}
	return secret
}

// ======================== 单元测试 + 临界值测试 ========================
func TestSetToContextWithValue(t *testing.T) {
	// 禁用Gin调试模式（测试必备）
	gin.SetMode(gin.TestMode)

	// 定义测试用例（覆盖所有临界场景+正常场景）
	tests := []struct {
		name        string
		setClaims   bool        // 是否在Gin上下文设置Claims
		claimsValue interface{} // Claims值（不同类型临界值）
		authHeader  string      // 请求头Authorization值（空/非空/特殊值）
		wantClaims  interface{} // 预期Claims值
		wantAuth    interface{} // 预期Authorization值（注意：无Claims时为nil）
		wantIsBgCtx bool        // 预期是否返回空Background()（无Claims时为true）
	}{
		// 临界值1：无Claims + AuthHeader非空 → 返回空Background，所有值为nil
		{
			name:        "no_claims_with_auth",
			setClaims:   false,
			claimsValue: nil,
			authHeader:  "Bearer test-token",
			wantClaims:  nil,
			wantAuth:    nil,
			wantIsBgCtx: true,
		},
		// 临界值2：无Claims + AuthHeader为空 → 返回空Background，所有值为nil
		{
			name:        "no_claims_no_auth",
			setClaims:   false,
			claimsValue: nil,
			authHeader:  "",
			wantClaims:  nil,
			wantAuth:    nil,
			wantIsBgCtx: true,
		},
		// 临界值3：有Claims（nil） + AuthHeader非空 → Claims为nil，Auth为对应值
		{
			name:        "claims_nil_with_auth",
			setClaims:   true,
			claimsValue: nil,
			authHeader:  "Bearer test-token",
			wantClaims:  nil,
			wantAuth:    "Bearer test-token",
			wantIsBgCtx: false,
		},
		// 临界值4：有Claims（字符串） + AuthHeader为空 → Claims为字符串，Auth为空字符串
		{
			name:        "claims_string_no_auth",
			setClaims:   true,
			claimsValue: "invalid-claims-string",
			authHeader:  "",
			wantClaims:  "invalid-claims-string",
			wantAuth:    "",
			wantIsBgCtx: false,
		},
		// 临界值5：有Claims（int） + AuthHeader含特殊字符 → 验证类型/值传递
		{
			name:        "claims_int_special_auth",
			setClaims:   true,
			claimsValue: 12345,
			authHeader:  "Bearer 123@abc 测试", // 特殊字符
			wantClaims:  12345,
			wantAuth:    "Bearer 123@abc 测试",
			wantIsBgCtx: false,
		},
		// 临界值6：有Claims（map） + AuthHeader大小写混合 → 验证map传递+header原样返回
		{
			name:        "claims_map_mixed_auth",
			setClaims:   true,
			claimsValue: map[string]interface{}{"uid": 1001, "role": "admin"},
			authHeader:  "  BeArEr   mock-valid-token  ", // 大小写+空格
			wantClaims:  map[string]interface{}{"uid": 1001, "role": "admin"},
			wantAuth:    "  BeArEr   mock-valid-token  ",
			wantIsBgCtx: false,
		},
		// 正常场景：有Claims（map） + AuthHeader标准值 → 全匹配
		{
			name:        "normal_claims_normal_auth",
			setClaims:   true,
			claimsValue: map[string]interface{}{"uid": 1001, "role": "admin", "name": "test"},
			authHeader:  "Bearer mock-valid-token",
			wantClaims:  map[string]interface{}{"uid": 1001, "role": "admin", "name": "test"},
			wantAuth:    "Bearer mock-valid-token",
			wantIsBgCtx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 创建兼容所有Gin版本的上下文
			req, _ := http.NewRequest("GET", "/test", nil)
			// 设置Authorization头
			req.Header.Set(JWT_HEADER_AUTHORIZATION, tt.authHeader)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = req

			// 2. 设置Claims到Gin上下文（注意：用string(JWT_CTX_CustomClaims_Infos)作为key）
			if tt.setClaims {
				c.Set(string(JWT_CTX_CustomClaims_Infos), tt.claimsValue)
			}

			// 3. 执行测试函数
			ctx := SetToContextWithValue(c, RequestIDFieldSnake)

			// 4. 断言1：验证是否返回空Background()（无Claims场景）
			if tt.wantIsBgCtx {
				// 空Background()的特征：无父上下文，无值
				assert.Equal(t, context.Background(), ctx, "[%s] 预期返回空Background()", tt.name)
			}

			// 5. 断言2：验证Claims值
			gotClaims := ctx.Value(JWT_CTX_CustomClaims_Infos)
			assert.Equal(t, tt.wantClaims, gotClaims,
				"[%s] Claims值不匹配\n预期：%#v\n实际：%#v",
				tt.name, tt.wantClaims, gotClaims)

			// 6. 断言3：验证Authorization值
			gotAuth := ctx.Value(JWT_HEADER_AUTHORIZATION)
			assert.Equal(t, tt.wantAuth, gotAuth,
				"[%s] Authorization值不匹配\n预期：%#v\n实际：%#v",
				tt.name, tt.wantAuth, gotAuth)
		})
	}
}

// TestGetCustomClaims 测试Claims获取（覆盖正常/异常/类型错误）
func TestGetCustomClaims(t *testing.T) {
	tests := []struct {
		name        string
		ctxValue    interface{}            // 上下文存入的值
		wantEmpty   bool                   // 是否返回空map
		wantContent map[string]interface{} // 期望的map内容
	}{
		// 临界值1：上下文无值
		{
			name:        "ctx_no_value",
			ctxValue:    nil,
			wantEmpty:   true,
			wantContent: map[string]interface{}{},
		},
		// 临界值2：值为int类型（非map）
		{
			name:        "ctx_value_int",
			ctxValue:    12345,
			wantEmpty:   true,
			wantContent: map[string]interface{}{},
		},
		// 临界值3：值为字符串类型（非map）
		{
			name:        "ctx_value_string",
			ctxValue:    "test-string",
			wantEmpty:   true,
			wantContent: map[string]interface{}{},
		},
		// 正常场景：值为合法map
		{
			name:        "ctx_value_valid_map",
			ctxValue:    map[string]interface{}{"uid": 1001.0, "name": "test-user"}, // JWT解析后数字为float64
			wantEmpty:   false,
			wantContent: map[string]interface{}{"uid": 1001.0, "name": "test-user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 创建上下文并设置值
			ctx := context.WithValue(context.Background(), JWT_CTX_CustomClaims_Infos, tt.ctxValue)

			// 2. 执行函数
			claims := GetCustomClaims(ctx)

			// 3. 断言结果
			assert.Equal(t, tt.wantEmpty, len(claims) == 0)
			assert.Equal(t, tt.wantContent, claims)
		})
	}
}

// TestGetAuthHeader 测试Authorization头获取（覆盖正常/异常/类型错误）
func TestGetAuthHeader(t *testing.T) {
	tests := []struct {
		name        string
		ctxValue    interface{} // 上下文存入的值
		wantEmpty   bool        // 是否返回空字符串
		wantContent string      // 期望的字符串内容
	}{
		// 临界值1：上下文无值
		{
			name:        "ctx_no_value",
			ctxValue:    nil,
			wantEmpty:   true,
			wantContent: "",
		},
		// 临界值2：值为int类型（非字符串）
		{
			name:        "ctx_value_int",
			ctxValue:    12345,
			wantEmpty:   true,
			wantContent: "",
		},
		// 临界值3：值为空字符串
		{
			name:        "ctx_value_empty_string",
			ctxValue:    "",
			wantEmpty:   true,
			wantContent: "",
		},
		// 正常场景：值为合法字符串
		{
			name:        "ctx_value_valid_string",
			ctxValue:    "Bearer " + generateTestJwtToken(t),
			wantEmpty:   false,
			wantContent: "Bearer " + generateTestJwtToken(t),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 创建上下文并设置值
			ctx := context.WithValue(context.Background(), JWT_HEADER_AUTHORIZATION, tt.ctxValue)

			// 2. 执行函数
			authHeader := GetAuthHeader(ctx)

			// 3. 断言结果
			assert.Equal(t, tt.wantEmpty, authHeader == "")
			assert.Equal(t, tt.wantContent, authHeader)
		})
	}
}

// ======================== 性能测试（Benchmark） ========================

// BenchmarkJwtAuthMiddleware_ValidToken 基准测试：有效Token场景下的中间件性能
func BenchmarkJwtAuthMiddleware_ValidToken(b *testing.B) {
	// 1. 预生成有效Token
	validToken := generateTestJwtToken(&testing.T{})
	authHeader := "Bearer " + validToken

	// 2. 初始化请求和上下文
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(JWT_HEADER_AUTHORIZATION, authHeader)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 3. 重置计时器（排除初始化耗时）
	b.ResetTimer()

	// 4. 执行基准测试
	for i := 0; i < b.N; i++ {
		middleware := jwtAuthMiddleware()
		middleware(c)
		//	c.Reset() // 重置上下文状态，避免污染下一次执行
	}
}

// BenchmarkJwtAuthMiddleware_InvalidToken 基准测试：无效Token场景下的中间件性能
func BenchmarkJwtAuthMiddleware_InvalidToken(b *testing.B) {
	// 1. 预生成无效Token
	invalidToken := generateInvalidJwtToken(&testing.T{})
	authHeader := "Bearer " + invalidToken

	// 2. 初始化请求和上下文
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(JWT_HEADER_AUTHORIZATION, authHeader)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 3. 重置计时器
	b.ResetTimer()

	// 4. 执行基准测试
	for i := 0; i < b.N; i++ {
		middleware := jwtAuthMiddleware()
		middleware(c)
		//		c.Reset()
	}
}
func BenchmarkContextOperations(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// 初始化上下文（使用指定常量）
	validToken := "mock-valid-token"
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(JWT_HEADER_AUTHORIZATION, "Bearer "+validToken)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	// 设置Claims：使用指定常量
	c.Set(JWT_CTX_CustomClaims_Infos, map[string]interface{}{
		"uid":  1001.0,
		"name": "test-user",
		"role": "admin",
	})

	b.ResetTimer()

	// 基准测试逻辑
	for i := 0; i < b.N; i++ {
		ctx := SetToContextWithValue(c, RequestIDFieldSnake)
		_ = GetCustomClaims(ctx)
		_ = GetAuthHeader(ctx)
	}
}

// 基准测试：分「有Claims」和「无Claims」两个核心场景（覆盖高频使用场景）
func BenchmarkSetToContextWithValue(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// 定义基准测试场景
	benchmarks := []struct {
		name       string
		setClaims  bool   // 是否设置Claims
		authHeader string // Authorization头值
	}{
		{name: "no_claims", setClaims: false, authHeader: ""},                         // 无Claims（高频异常场景）
		{name: "with_claims", setClaims: true, authHeader: "Bearer mock-valid-token"}, // 有Claims（高频正常场景）
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// 初始化上下文（只执行一次，避免计入基准时间）
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set(JWT_HEADER_AUTHORIZATION, bm.authHeader)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = req

			// 设置Claims（只执行一次）
			if bm.setClaims {
				c.Set(string(JWT_CTX_CustomClaims_Infos), map[string]interface{}{
					"uid":  1001,
					"role": "admin",
					"name": "test-user",
				})
			}

			// 重置计时器（排除初始化耗时）
			b.ResetTimer()

			// 执行基准测试（循环b.N次）
			for i := 0; i < b.N; i++ {
				_ = SetToContextWithValue(c, RequestIDFieldSnake)
			}
		})
	}
}
func BenchmarkContextOperationsFullChain(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// 初始化上下文（模拟真实业务场景）
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(JWT_HEADER_AUTHORIZATION, "Bearer mock-valid-token")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set(string(JWT_CTX_CustomClaims_Infos), map[string]interface{}{
		"uid":  1001,
		"role": "admin",
		"name": "test-user",
	})

	b.ResetTimer()

	// 模拟业务中完整的上下文操作链
	for i := 0; i < b.N; i++ {
		ctx := SetToContextWithValue(c, RequestIDFieldSnake)
		_ = GetCustomClaims(ctx)
		_ = GetAuthHeader(ctx)
	}
}

///////////////////////////////////////////////////////////
/*
// start server
func TestJwt(t *testing.T) {
	start := time.Now()
	var g errgroup.Group

	serverHttpPort := "1024"

	gin.SetMode(gin.DebugMode)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	public := engine.Group("/public")
	{
		public.POST("/hello", Hello)
		public.POST("/login", Login)
	}
	auth := engine.Group("/auth")
	{
		RouterGroupRegisterJwt(auth)
		auth.POST("/work", Work)
	}
	g.Go(func() error {
		//	belogs.Info("startRpServer(): server run http on :", serverHttpPort)
		err := engine.Run(":" + serverHttpPort)
		return err
	})

	if err := g.Wait(); err != nil {
		//belogs.Error("startRpServer(): fail, will exit, err:", err)
		fmt.Println("startRpServer(): fail, will exit,", "port", serverHttpPort, " err:", err)
	}
	//belogs.Info("startRpServer(): server end, time(s):", time.Since(start))
	fmt.Println("startRpServer(): server end", "time(s)", time.Since(start))

}
func Hello(c *gin.Context) {
	String(c, "hello")
}

// will generate and return jwt token
func Login(c *gin.Context) {
	m := make(map[string]interface{})
	m["ownerId"] = "1001"
	m["ownerName"] = "beijing-user1"
	m["opUserId"] = "2002"
	m["opUserName"] = "beijing-user2"
	m["traceId"] = "550e8400-e29b-41d4-a716-446655440000"
	m["opLogId"] = "3003"

	claims := jwtutil.CustomClaims{
		Infos: m,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)), //过期时间
			NotBefore: jwt.NewNumericDate(time.Now()),                    //生效时间（立即生效）
			IssuedAt:  jwt.NewNumericDate(time.Now()),                    //签发时间
		},
	}
	token, _ := jwtutil.GenToken(&claims, conf.String("jwt::secret"))
	fmt.Println("Login(): get claim:", "token", token)
	String(c, token)
}

// will use jwt token in header, and will pass jwt check
func Work(c *gin.Context) {
	fmt.Println("Work(): start")
	String(c, "work ok")
}
*/
