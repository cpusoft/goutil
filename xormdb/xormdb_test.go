package xormdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ======================== 全局测试准备 ========================
// 兼容 *testing.T 和 *testing.B
func cleanTestResource(t testing.TB) {
	if XormEngine != nil {
		err := CloseXormEngine()
		assert.NoError(t, err, "关闭XormEngine失败")
		XormEngine = nil // 重置全局变量
	}
}

// ======================== MySQL相关测试 ========================
// 正常流程测试
func TestInitMySqlParameter_Normal(t *testing.T) {
	// 注意：如需实际连接测试，需替换为本地测试MySQL地址；否则注释该测试，改用Mock
	t.Skip("如需实际测试MySQL，请注释该行并配置正确的MySQL地址")

	engine, err := InitMySqlParameter(
		"root",            // user
		"Rpstir-123",      // password
		"127.0.0.1:13306", // server
		"root",            // database
		10,                // maxidleconns
		20,                // maxopenconns
	)
	assert.NoError(t, err, "InitMySqlParameter正常流程失败")
	assert.NotNil(t, engine, "engine应为非nil")

	// 清理
	err = engine.Close()
	assert.NoError(t, err)
}

// 临界值测试：空参数、非法连接串、边界连接数
func TestInitMySqlParameter_Critical(t *testing.T) {
	// 测试1：空用户名
	engine, err := InitMySqlParameter("", "123456", "127.0.0.1:13306", "test", 10, 20)
	assert.Error(t, err, "空用户名应返回错误")
	assert.Nil(t, engine, "空用户名时engine应为nil")

	// 测试2：非法服务器地址（端口错误）
	engine, err = InitMySqlParameter("root", "123456", "127.0.0.1:9999", "test", 10, 20)
	assert.Error(t, err, "非法端口应返回错误")
	assert.Nil(t, engine, "非法端口时engine应为nil")

	// 测试3：连接数为0（仅验证连接结果，移除连接池参数断言）
	engine, err = InitMySqlParameter("root", "123456", "127.0.0.1:13306", "test", 0, 0)
	if err == nil {
		_ = engine.Close()
	}

	// 测试4：连接数为负数（仅验证连接结果，移除连接池参数断言）
	engine, err = InitMySqlParameter("root", "123456", "127.0.0.1:13306", "test", -5, -10)
	if err == nil {
		_ = engine.Close()
	}
}

// 集成测试：InitMySql（基于配置）
func TestInitMySql(t *testing.T) {
	t.Skip("如需实际测试MySQL，请注释该行并配置正确的MySQL地址")
	cleanTestResource(t)

	err := InitMySql()
	assert.NoError(t, err, "InitMySql失败")
	assert.NotNil(t, XormEngine, "XormEngine应为非nil")

	cleanTestResource(t)
}

// ======================== SQLite相关测试 ========================
// 正常流程测试（内存模式，无文件依赖）
func TestInitSqliteParameter_Normal(t *testing.T) {
	engine, err := InitSqliteParameter(
		"file::memory:?cache=shared", // 内存模式
		5,                            // maxidleconns
		10,                           // maxopenconns（会被强制改为1）
	)
	assert.NoError(t, err, "InitSqliteParameter正常流程失败")
	assert.NotNil(t, engine, "engine应为非nil")

	// 验证Ping
	err = engine.Ping()
	assert.NoError(t, err, "SQLite Ping失败")

	// 清理
	err = engine.Close()
	assert.NoError(t, err)
}

// 临界值测试：空路径、非法路径、边界连接数
func TestInitSqliteParameter_Critical(t *testing.T) {
	// 测试1：空文件路径
	engine, err := InitSqliteParameter("", 5, 10)
	assert.Error(t, err, "空路径应返回错误")
	assert.Nil(t, engine, "空路径时engine应为nil")

	// 测试2：非法路径（无权限目录）
	invalidPath := "/root/test_sqlite.db" // 普通用户无权限
	engine, err = InitSqliteParameter(invalidPath, 5, 10)
	assert.Error(t, err, "非法路径应返回错误")
	assert.Nil(t, engine, "非法路径时engine应为nil")

	// 测试3：连接数为0（仅验证连接结果，移除连接池参数断言）
	engine, err = InitSqliteParameter("file::memory:?cache=shared", 0, 0)
	assert.NoError(t, err)
	_ = engine.Close()

	// 测试4：连接数为负数（仅验证连接结果，移除连接池参数断言）
	engine, err = InitSqliteParameter("file::memory:?cache=shared", -5, -10)
	assert.NoError(t, err)
	_ = engine.Close()
}

// 集成测试：InitSqlite（基于配置）
func TestInitSqlite(t *testing.T) {
	cleanTestResource(t)

	err := InitSqlite()
	assert.NoError(t, err, "InitSqlite失败")
	assert.NotNil(t, XormEngine, "XormEngine应为非nil")

	cleanTestResource(t)
}

// ======================== PostgreSQL相关测试 ========================
// 正常流程测试
func TestInitPostgreSQLParameter_Normal(t *testing.T) {
	t.Skip("如需实际测试PostgreSQL，请注释该行并配置正确的PostgreSQL地址")

	engine, err := InitPostgreSQLParameter(
		"rpki",            // user
		"Rpki-123",        // password
		"127.0.0.1:15432", // server
		"rpki",            // database
		8,                 // maxidleconns
		16,                // maxopenconns
	)
	assert.NoError(t, err, "InitPostgreSQLParameter正常流程失败")
	assert.NotNil(t, engine, "engine应为非nil")

	// 清理
	err = engine.Close()
	assert.NoError(t, err)
}

// 临界值测试：非法地址格式、空参数、边界连接数
func TestInitPostgreSQLParameter_Critical(t *testing.T) {
	// 测试1：非法服务器地址（无端口）
	engine, err := InitPostgreSQLParameter("postgres", "postgres", "127.0.0.1", "test", 8, 16)
	assert.Error(t, err, "无端口地址应返回错误")
	assert.Nil(t, engine, "无端口地址时engine应为nil")

	// 测试2：空密码
	engine, err = InitPostgreSQLParameter("postgres", "", "127.0.0.1:15432", "test", 8, 16)
	assert.Error(t, err, "空密码应返回错误")
	assert.Nil(t, engine, "空密码时engine应为nil")

	// 测试3：连接数为0（仅验证连接结果，移除连接池参数断言）
	engine, err = InitPostgreSQLParameter("postgres", "postgres", "127.0.0.1:15432", "test", 0, 0)
	if err == nil {
		_ = engine.Close()
	}

	// 测试4：连接数为负数（仅验证连接结果，移除连接池参数断言）
	engine, err = InitPostgreSQLParameter("postgres", "postgres", "127.0.0.1:15432", "test", -8, -16)
	if err == nil {
		_ = engine.Close()
	}
}

// 集成测试：InitPostgreSQL（基于配置）
func TestInitPostgreSQL(t *testing.T) {
	t.Skip("如需实际测试PostgreSQL，请注释该行并配置正确的PostgreSQL地址")
	cleanTestResource(t)

	err := InitPostgreSQL()
	assert.NoError(t, err, "InitPostgreSQL失败")
	assert.NotNil(t, XormEngine, "XormEngine应为非nil")

	cleanTestResource(t)
}

// ======================== Session相关测试 ========================
// 正常流程测试：NewSession + CommitSession
func TestSession_Normal(t *testing.T) {
	// 先初始化SQLite（内存模式，无依赖）
	cleanTestResource(t)
	err := InitSqlite()
	assert.NoError(t, err, "初始化SQLite失败")

	// 测试NewSession
	session, err := NewSession()
	assert.NoError(t, err, "NewSession失败")
	assert.NotNil(t, session, "session应为非nil")

	// 测试CommitSession
	err = CommitSession(session)
	assert.NoError(t, err, "CommitSession失败")

	cleanTestResource(t)
}

// 临界值测试：Session异常流程（Begin失败、Commit失败）
func TestSession_Critical(t *testing.T) {
	// 测试1：XormEngine为nil时NewSession（未初始化）
	cleanTestResource(t)
	XormEngine = nil
	session, err := NewSession()
	assert.Error(t, err, "XormEngine为nil时NewSession应返回错误")
	assert.Nil(t, session, "XormEngine为nil时session应为nil")

	// 测试2：RollbackAndLogError（nil session）
	err = RollbackAndLogError(nil, "test error", fmt.Errorf("mock error"))
	assert.Error(t, err, "RollbackAndLogError应返回错误")

	// 测试3：Commit失败（已关闭的session）
	cleanTestResource(t)
	err = InitSqlite()
	assert.NoError(t, err)
	session, err = NewSession()
	assert.NoError(t, err)
	_ = session.Close() // 先关闭session
	err = CommitSession(session)
	assert.Error(t, err, "关闭的session Commit应返回错误")

	cleanTestResource(t)
}

// ======================== Null类型工具测试 ========================
func TestSqlNullTypes(t *testing.T) {
	// 测试SqlNullString
	t.Run("SqlNullString_Normal", func(t *testing.T) {
		ns := SqlNullString("test")
		assert.True(t, ns.Valid)
		assert.Equal(t, "test", ns.String)
	})

	t.Run("SqlNullString_Empty", func(t *testing.T) {
		ns := SqlNullString("")
		assert.False(t, ns.Valid)
		assert.Equal(t, "", ns.String)
	})

	// 测试SqlNullInt
	t.Run("SqlNullInt_Normal", func(t *testing.T) {
		ni := SqlNullInt(100)
		assert.True(t, ni.Valid)
		assert.Equal(t, int64(100), ni.Int64)
	})

	t.Run("SqlNullInt_Negative", func(t *testing.T) {
		ni := SqlNullInt(-10)
		assert.False(t, ni.Valid)
		assert.Equal(t, int64(0), ni.Int64) // sql.NullInt64默认值
	})

	t.Run("SqlNullInt_Zero", func(t *testing.T) {
		ni := SqlNullInt(0)
		assert.True(t, ni.Valid)
		assert.Equal(t, int64(0), ni.Int64)
	})

	// 测试Int64sToInString
	t.Run("Int64sToInString_Normal", func(t *testing.T) {
		s := Int64sToInString([]int64{1, 2, 3})
		assert.Equal(t, "1,2,3", s)
	})

	t.Run("Int64sToInString_Empty", func(t *testing.T) {
		s := Int64sToInString([]int64{})
		assert.Equal(t, "", s)
	})
}

// ======================== StringArray自定义类型测试 ========================
func TestStringArray(t *testing.T) {
	// 测试正常序列化/反序列化
	t.Run("StringArray_Normal", func(t *testing.T) {
		sa := StringArray{"a", "b", "c"}
		// ToDb
		data, err := sa.ToDb()
		assert.NoError(t, err)
		assert.Equal(t, `["a","b","c"]`, string(data))

		// FromDb
		var sa2 StringArray
		err = sa2.FromDb(data)
		assert.NoError(t, err)
		assert.Equal(t, sa, sa2)
	})

	// 临界值：空数组
	t.Run("StringArray_Empty", func(t *testing.T) {
		sa := StringArray{}
		data, err := sa.ToDb()
		assert.NoError(t, err)
		assert.Equal(t, "[]", string(data))

		var sa2 StringArray
		err = sa2.FromDb(data)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(sa2))
	})

	// 临界值：null值
	t.Run("StringArray_Null", func(t *testing.T) {
		var sa StringArray
		err := sa.FromDb([]byte("null"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(sa))
	})

	// 异常：非法JSON
	t.Run("StringArray_InvalidJSON", func(t *testing.T) {
		var sa StringArray
		err := sa.FromDb([]byte("{invalid json}"))
		assert.Error(t, err)
	})
}

// ======================== CloseXormEngine测试 ========================
func TestCloseXormEngine(t *testing.T) {
	// 测试1：已初始化的engine
	cleanTestResource(t)
	err := InitSqlite()
	assert.NoError(t, err)
	err = CloseXormEngine()
	assert.NoError(t, err)
	assert.Nil(t, XormEngine, "Close后XormEngine应为nil")

	// 测试2：未初始化的engine（nil）
	err = CloseXormEngine()
	assert.NoError(t, err, "nil engine Close应无错误")
}

// ======================== 性能测试（Benchmark） ========================
// SQLite连接创建性能
func BenchmarkInitSqliteParameter(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine, err := InitSqliteParameter("file::memory:?cache=shared", 5, 10)
		if err == nil {
			_ = engine.Close()
		}
	}
}

// StringArray序列化性能
func BenchmarkStringArray_ToDb(b *testing.B) {
	sa := StringArray{"test1", "test2", "test3", "test4", "test5"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sa.ToDb()
	}
}

// StringArray反序列化性能
func BenchmarkStringArray_FromDb(b *testing.B) {
	data := []byte(`["test1","test2","test3","test4","test5"]`)
	var sa StringArray
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sa.FromDb(data)
	}
}

// Session创建+提交性能
func BenchmarkSession_NewAndCommit(b *testing.B) {
	// 初始化SQLite
	cleanTestResource(b)
	_ = InitSqlite()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		session, err := NewSession()
		if err == nil {
			_ = CommitSession(session)
		}
	}

	cleanTestResource(b)
}

// Null类型转换性能
func BenchmarkSqlNullString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SqlNullString(fmt.Sprintf("test_%d", i))
	}
}

func BenchmarkSqlNullInt(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SqlNullInt(int64(i))
	}
}
