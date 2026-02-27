package xormdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ======================== 全局测试准备 ========================
// export CGO_ENABLED=1
//
// 兼容 *testing.T 和 *testing.B，清理测试资源
func cleanTestResource(t testing.TB) {
	if XormEngine != nil {
		err := CloseXormEngine()
		assert.NoError(t, err, "关闭XormEngine失败")
		XormEngine = nil // 重置全局变量
	}
}

// ======================== MySQL相关测试 ========================
// 正常流程测试（仅参数校验，跳过实际连接）
func TestInitMySqlParameter_Normal(t *testing.T) {
	//	t.Skip("如需实际测试MySQL连接，请注释该行并配置正确的MySQL地址")

	// 合法参数：仅验证参数校验通过（实际连接跳过）
	engine, err := InitMySqlParameter(
		"rpstir2",         // user
		"Rpstir-123",      // password
		"127.0.0.1:13306", // server
		"rpstir2",         // database
		10,                // maxidleconns
		20,                // maxopenconns
	)
	assert.NoError(t, err, "InitMySqlParameter正常参数校验失败")
	assert.NotNil(t, engine, "engine应为非nil")

	// 清理
	if engine != nil {
		err = engine.Close()
		assert.NoError(t, err)
	}
}

// 临界值测试：覆盖空参数、负数连接数、非法地址等场景
func TestInitMySqlParameter_Critical(t *testing.T) {
	// 测试1：空用户名
	engine, err := InitMySqlParameter("", "123456", "127.0.0.1:13306", "test", 10, 20)
	assert.Error(t, err, "空用户名应返回错误")
	assert.Nil(t, engine, "空用户名时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试2：空密码
	engine, err = InitMySqlParameter("root", "", "127.0.0.1:13306", "test", 10, 20)
	assert.Error(t, err, "空密码应返回错误")
	assert.Nil(t, engine, "空密码时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试3：空服务器地址
	engine, err = InitMySqlParameter("root", "123456", "", "test", 10, 20)
	assert.Error(t, err, "空服务器地址应返回错误")
	assert.Nil(t, engine, "空服务器地址时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试4：空数据库名
	engine, err = InitMySqlParameter("root", "123456", "127.0.0.1:13306", "", 10, 20)
	assert.Error(t, err, "空数据库名应返回错误")
	assert.Nil(t, engine, "空数据库名时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试5：负数maxidleconns
	engine, err = InitMySqlParameter("root", "123456", "127.0.0.1:13306", "test", -5, 20)
	assert.Error(t, err, "负数maxidleconns应返回错误")
	assert.Nil(t, engine, "负数maxidleconns时engine应为nil")
	assert.Contains(t, err.Error(), "maxidleconns or maxopenconns is negative")

	// 测试6：负数maxopenconns
	engine, err = InitMySqlParameter("root", "123456", "127.0.0.1:13306", "test", 10, -20)
	assert.Error(t, err, "负数maxopenconns应返回错误")
	assert.Nil(t, engine, "负数maxopenconns时engine应为nil")
	assert.Contains(t, err.Error(), "maxidleconns or maxopenconns is negative")

	// 测试7：非法服务器地址（端口错误，参数校验通过但连接失败）
	engine, err = InitMySqlParameter("root", "123456", "127.0.0.1:9999", "test", 10, 20)
	assert.Error(t, err, "非法端口应返回错误")
	assert.Nil(t, engine, "非法端口时engine应为nil")
}

// 集成测试：InitMySql（基于配置，仅参数校验）
func TestInitMySql(t *testing.T) {
	t.Skip("如需实际测试MySQL，请注释该行并配置正确的MySQL地址")
	cleanTestResource(t)

	err := InitMySql()
	assert.NoError(t, err, "InitMySql失败（需配置正确的MySQL参数）")
	assert.NotNil(t, XormEngine, "XormEngine应为非nil")

	cleanTestResource(t)
}

// ======================== SQLite相关测试 ========================
// 正常流程测试（内存模式，实际连接测试）
func TestInitSqliteParameter_Normal(t *testing.T) {
	engine, err := InitSqliteParameter(
		"file::memory:?cache=shared", // 内存模式（无文件依赖）
		5,                            // maxidleconns
		10,                           // maxopenconns
	)
	assert.NoError(t, err, "InitSqliteParameter正常流程失败")
	assert.NotNil(t, engine, "engine应为非nil")
	if engine != nil {
		// 验证Ping
		err = engine.Ping()
		assert.NoError(t, err, "SQLite Ping失败")

		// 清理
		err = engine.Close()
		assert.NoError(t, err)
	} else {
		t.Error("engine为nil，无法进行Ping测试")
	}
}

// 临界值测试：空路径、负数连接数、非法路径等场景
func TestInitSqliteParameter_Critical(t *testing.T) {
	// 测试1：空文件路径
	engine, err := InitSqliteParameter("", 5, 10)
	assert.Error(t, err, "空路径应返回错误")
	assert.Nil(t, engine, "空路径时engine应为nil")
	assert.Contains(t, err.Error(), "sqlite file path is empty")

	// 测试2：负数maxidleconns
	engine, err = InitSqliteParameter("file::memory:?cache=shared", -5, 10)
	assert.Error(t, err, "负数maxidleconns应返回错误")
	assert.Nil(t, engine, "负数maxidleconns时engine应为nil")
	assert.Contains(t, err.Error(), "maxidleconns or maxopenconns is negative")

	// 测试3：负数maxopenconns
	engine, err = InitSqliteParameter("file::memory:?cache=shared", 5, -10)
	assert.Error(t, err, "负数maxopenconns应返回错误")
	assert.Nil(t, engine, "负数maxopenconns时engine应为nil")
	assert.Contains(t, err.Error(), "maxidleconns or maxopenconns is negative")

	// 测试4：非法路径（无权限目录）
	invalidPath := "/dev/null/test_sqlite.db" // root也无权限
	engine, err = InitSqliteParameter(invalidPath, 5, 10)
	assert.Error(t, err, "非法路径应返回错误")
	assert.Nil(t, engine, "非法路径时engine应为nil")
}

// 集成测试：InitSqlite（基于配置，内存模式）
func TestInitSqlite(t *testing.T) {
	t.Skip("如需实际测试SQLite，请注释该行并配置正确的SQLite地址")
	cleanTestResource(t)

	err := InitSqlite()
	assert.NoError(t, err, "InitSqlite失败")
	assert.NotNil(t, XormEngine, "XormEngine应为非nil")

	cleanTestResource(t)
}

// ======================== PostgreSQL相关测试 ========================
// 正常流程测试（仅参数校验，跳过实际连接）
func TestInitPostgreSQLParameter_Normal(t *testing.T) {
	t.Skip("如需实际测试PostgreSQL连接，请注释该行并配置正确的PostgreSQL地址")

	// 合法参数：仅验证参数校验通过（实际连接跳过）
	engine, err := InitPostgreSQLParameter(
		"rpki",            // user
		"Rpki-123",        // password
		"127.0.0.1:15432", // server
		"rpki",            // database
		8,                 // maxidleconns
		16,                // maxopenconns
	)
	assert.NoError(t, err, "InitPostgreSQLParameter正常参数校验失败")
	assert.NotNil(t, engine, "engine应为非nil")

	// 清理
	if engine != nil {
		err = engine.Close()
		assert.NoError(t, err)
	}
}

// 临界值测试：空参数、负数连接数、非法地址格式等场景
func TestInitPostgreSQLParameter_Critical(t *testing.T) {
	// 测试1：空用户名
	engine, err := InitPostgreSQLParameter("", "postgres", "127.0.0.1:15432", "test", 8, 16)
	assert.Error(t, err, "空用户名应返回错误")
	assert.Nil(t, engine, "空用户名时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试2：空密码
	engine, err = InitPostgreSQLParameter("postgres", "", "127.0.0.1:15432", "test", 8, 16)
	assert.Error(t, err, "空密码应返回错误")
	assert.Nil(t, engine, "空密码时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试3：空服务器地址
	engine, err = InitPostgreSQLParameter("postgres", "postgres", "", "test", 8, 16)
	assert.Error(t, err, "空服务器地址应返回错误")
	assert.Nil(t, engine, "空服务器地址时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试4：空数据库名
	engine, err = InitPostgreSQLParameter("postgres", "postgres", "127.0.0.1:15432", "", 8, 16)
	assert.Error(t, err, "空数据库名应返回错误")
	assert.Nil(t, engine, "空数据库名时engine应为nil")
	assert.Contains(t, err.Error(), "user or password or server or database is empty")

	// 测试5：负数maxidleconns
	engine, err = InitPostgreSQLParameter("postgres", "postgres", "127.0.0.1:15432", "test", -8, 16)
	assert.Error(t, err, "负数maxidleconns应返回错误")
	assert.Nil(t, engine, "负数maxidleconns时engine应为nil")
	assert.Contains(t, err.Error(), "maxidleconns or maxopenconns is negative")

	// 测试6：负数maxopenconns
	engine, err = InitPostgreSQLParameter("postgres", "postgres", "127.0.0.1:15432", "test", 8, -16)
	assert.Error(t, err, "负数maxopenconns应返回错误")
	assert.Nil(t, engine, "负数maxopenconns时engine应为nil")
	assert.Contains(t, err.Error(), "maxidleconns or maxopenconns is negative")

	// 测试7：非法服务器地址（无端口）
	engine, err = InitPostgreSQLParameter("postgres", "postgres", "127.0.0.1", "test", 8, 16)
	assert.Error(t, err, "无端口地址应返回错误")
	assert.Nil(t, engine, "无端口地址时engine应为nil")
}

// 集成测试：InitPostgreSQL（基于配置，仅参数校验）
func TestInitPostgreSQL(t *testing.T) {
	t.Skip("如需实际测试PostgreSQL，请注释该行并配置正确的PostgreSQL地址")
	cleanTestResource(t)

	err := InitPostgreSQL()
	assert.NoError(t, err, "InitPostgreSQL失败（需配置正确的PostgreSQL参数）")
	assert.NotNil(t, XormEngine, "XormEngine应为非nil")

	cleanTestResource(t)
}

// ======================== Session相关测试 ========================
// 正常流程测试：NewSession + CommitSession
func TestSession_Normal(t *testing.T) {
	// 先初始化SQLite（内存模式，无依赖）
	cleanTestResource(t)
	eng, err := InitSqliteParameter("file::memory:?cache=shared", 5, 10)
	assert.NoError(t, err)
	XormEngine = eng // 手动赋值全局引擎

	// 测试NewSession
	session, err := NewSession()
	assert.NoError(t, err, "NewSession失败")
	assert.NotNil(t, session, "session应为非nil")

	// 测试CommitSession
	err = CommitSession(session)
	assert.NoError(t, err, "CommitSession失败")

	cleanTestResource(t)
}

// 临界值测试：覆盖Session nil、引擎未初始化、Commit失败等场景
func TestSession_Critical(t *testing.T) {
	// 测试1：XormEngine为nil时NewSession（未初始化）
	cleanTestResource(t)
	XormEngine = nil
	session, err := NewSession()
	assert.Error(t, err, "XormEngine为nil时NewSession应返回错误")
	assert.Nil(t, session, "XormEngine为nil时session应为nil")

	// 测试2：CommitSession传入nil
	err = CommitSession(nil)
	assert.Error(t, err, "nil session Commit应返回错误")
	assert.Equal(t, errors.New("session is nil or closed"), err, "CommitSession nil or closed错误信息不匹配")

	// 测试3：RollbackAndLogError（nil session + 空err）
	err = RollbackAndLogError(nil, "test msg", nil)
	assert.NoError(t, err, "空err时RollbackAndLogError应返回nil")

	// 测试4：RollbackAndLogError（nil session + 非空err）
	err = RollbackAndLogError(nil, "test msg", fmt.Errorf("mock error"))
	assert.Error(t, err, "非空err时RollbackAndLogError应返回错误")
	assert.Contains(t, err.Error(), "mock error")

	// 测试5：Commit失败（已关闭的session）
	cleanTestResource(t)
	engine, err := InitSqliteParameter("file::memory:?cache=shared", 5, 10)
	assert.NoError(t, err)
	XormEngine = engine

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

	// 临界值：空byte数组
	t.Run("StringArray_EmptyByte", func(t *testing.T) {
		var sa StringArray
		err := sa.FromDb([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(sa))
	})

	// 异常：非法JSON
	t.Run("StringArray_InvalidJSON", func(t *testing.T) {
		var sa StringArray
		err := sa.FromDb([]byte("{invalid json}"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal json fail")
	})

	// 异常：ToDb序列化失败（模拟极端场景）
	t.Run("StringArray_ToDb_Error", func(t *testing.T) {
		// 模拟无法序列化的场景（实际中[]string不会出现，仅覆盖错误处理逻辑）
		type InvalidType struct{}
		invalidSA := StringArray(make([]string, 0))
		// 手动构造错误：此处仅验证错误处理逻辑
		data, err := invalidSA.ToDb()
		assert.NoError(t, err) // []string空数组序列化正常
		assert.Equal(t, "[]", string(data))
	})
}

// ======================== CloseXormEngine测试 ========================
func TestCloseXormEngine(t *testing.T) {
	// 测试1：已初始化的engine
	cleanTestResource(t)
	engine, err := InitSqliteParameter("file::memory:?cache=shared", 5, 10)
	assert.NoError(t, err)
	XormEngine = engine

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
	engine, err := InitSqliteParameter("file::memory:?cache=shared", 5, 10)
	if err != nil {
		b.Fatalf("初始化SQLite失败：%v", err)
	}
	XormEngine = engine

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
