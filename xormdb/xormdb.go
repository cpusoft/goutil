package xormdb

import (
	"fmt"
	"net"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

// 修复：初始化为nil（未初始化前使用会显式panic，而非操作空结构体）
var XormEngine *xorm.Engine

// ////////////////////////////////////////////
// MySQL
func InitMySql() (err error) {
	user := conf.String("mysql::user")
	password := conf.String("mysql::password")
	server := conf.String("mysql::server")
	database := conf.String("mysql::database")
	maxidleconns := conf.Int("mysql::maxidleconns")
	maxopenconns := conf.Int("mysql::maxopenconns")
	XormEngine, err = InitMySqlParameter(user, password, server, database, maxidleconns, maxopenconns)
	if err != nil {
		belogs.Error("InitMySql(): fail: ", err)
		return err
	}
	return nil
}

func InitMySqlParameter(user, password, server, database string, maxidleconns, maxopenconns int) (engine *xorm.Engine, err error) {
	// 修复：charset改为utf8mb4，支持完整UTF-8（含emoji等4字节字符）
	openSql := user + ":" + password + "@tcp(" + server + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
	// 修复：删除未使用的logName变量
	belogs.Info("InitMySqlParameter(): server is: ", server, database)

	// 连接数据库
	engine, err = xorm.NewEngine("mysql", openSql)
	if err != nil {
		belogs.Error("InitMySqlParameter(): NewEngine failed, err:", err)
		return engine, err
	}

	// 连接测试 + 修复：Ping失败后关闭engine，避免连接泄漏
	if err := engine.Ping(); err != nil {
		belogs.Error("InitMySqlParameter(): Ping failed, err:", err)
		_ = engine.Close() // 关闭已创建的engine，避免资源泄漏
		return engine, err
	}

	// 设置连接池参数
	engine.SetMaxIdleConns(maxidleconns)
	engine.SetMaxOpenConns(maxopenconns)
	// show sql
	//engine.ShowSQL(true)
	/*
		http://blog.xorm.io/2016/1/4/1-about-mapper.html
		SnakeMapper
		SnakeMapper是默认的映射机制，他支持数据库表采用匈牙利命名法，而程序中采用驼峰式命名法。下面是一些常见的映射：
		表中名称		程序名称
		user_info	UserInfo
		id			Id

		SameMapper
		SameMapper就是数据库中的命名法和程序中是相同的。那么鉴于在Go中，基本上要求首字母必须大写。所以一般都是表中和程序中均采用驼峰式命名。下面是一些常见的映射：
		表中名称	程序名称
		UserInfo	UserInfo
		Id	Id


		GonicMapper
		GonicMapper是在SnakeMapper的基础上增加了特例，对于常见的缩写不新增下划线处理。这个同时也符合golint的规则。下面是一些常见的映射：
		表中名称	程序名称
		user_info	UserInfo
		id	ID
		url	URL
	*/
	engine.SetTableMapper(names.SnakeMapper{})

	return engine, nil
}

// ///////////////////////////////////////////
// SQLite
func InitSqlite() (err error) {
	filepath := conf.String("sqlite::filepath")
	maxidleconns := conf.Int("sqlite::maxidleconns")
	maxopenconns := conf.Int("sqlite::maxopenconns")
	XormEngine, err = InitSqliteParameter(filepath, maxidleconns, maxopenconns)
	if err != nil {
		belogs.Error("InitSqlite(): fail:", err)
		return err
	}
	return nil
}

// 修复：参数名改为sqliteFilePath，避免与filepath包名混淆
func InitSqliteParameter(sqliteFilePath string, maxidleconns, maxopenconns int) (engine *xorm.Engine, err error) {
	belogs.Info("InitSqliteParameter(): sqliteFilePath: ", sqliteFilePath)

	// 连接数据库
	engine, err = xorm.NewEngine("sqlite3", sqliteFilePath)
	if err != nil {
		belogs.Error("InitSqliteParameter(): NewEngine failed, sqliteFilePath:", sqliteFilePath, err)
		return engine, err
	}

	// 连接测试 + 修复：Ping失败后关闭engine
	if err := engine.Ping(); err != nil {
		belogs.Error("InitSqliteParameter(): Ping failed, err:", err)
		_ = engine.Close()
		return engine, err
	}

	// SQLite连接池修复：强制SetMaxOpenConns(1)（SQLite不支持多连接，否则易锁库）
	// 保留配置参数兼容，但覆盖为安全值
	engine.SetMaxIdleConns(maxidleconns)
	engine.SetMaxOpenConns(1) // 核心修复：SQLite必须单连接
	engine.SetTableMapper(names.SnakeMapper{})

	return engine, nil
}

// ////////////////////////////////////////////
// PostgreSQL
func InitPostgreSQL() (err error) {
	user := conf.String("postgresql::user")
	password := conf.String("postgresql::password")
	server := conf.String("postgresql::server")
	database := conf.String("postgresql::database")
	maxidleconns := conf.Int("postgresql::maxidleconns")
	maxopenconns := conf.Int("postgresql::maxopenconns")
	XormEngine, err = InitPostgreSQLParameter(user, password, server, database, maxidleconns, maxopenconns)
	if err != nil {
		belogs.Error("InitPostgreSQL(): fail: ", err)
		return err
	}
	return nil
}

func InitPostgreSQLParameter(user, password, server, database string, maxidleconns, maxopenconns int) (engine *xorm.Engine, err error) {
	host, port, err := net.SplitHostPort(server)
	if err != nil {
		belogs.Error("InitPostgreSQLParameter(): SplitHostPort fail, server:", server, err)
		return nil, err // 修复：逻辑更清晰，直接返回nil而非engine（此时engine未初始化）
	}
	str := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)
	// 修复：删除未使用的logName变量
	belogs.Info("InitPostgreSQLParameter(): server:", server, "  database:", database)

	// 连接数据库
	engine, err = xorm.NewEngine("postgres", str)
	if err != nil {
		belogs.Error("InitPostgreSQLParameter(): NewEngine failed, err:", err)
		return engine, err
	}

	// 连接测试 + 修复：Ping失败后关闭engine
	if err := engine.Ping(); err != nil {
		belogs.Error("InitPostgreSQLParameter(): Ping failed, err:", err)
		_ = engine.Close()
		return engine, err
	}

	// 设置连接池参数
	engine.SetMaxIdleConns(maxidleconns)
	engine.SetMaxOpenConns(maxopenconns)
	engine.SetTableMapper(names.SnakeMapper{})

	return engine, nil
}

// //////////////////////////////////
// Session utils
func NewSession() (*xorm.Session, error) {
	session := XormEngine.NewSession()
	if err := session.Begin(); err != nil {
		return nil, RollbackAndLogError(session, "session.Begin() fail", err)
	}
	return session, nil
}

// commit session, if err, will rollback.
// must return error, so not use in defer,
func CommitSession(session *xorm.Session) error {
	if err := session.Commit(); err != nil {
		belogs.Error("main():Commit fail")
		return RollbackAndLogError(session, "session.Commit fail", err)
	}
	return nil
}

// 修复：检查Rollback()的错误并日志，避免丢失关键错误
func RollbackAndLogError(session *xorm.Session, msg string, err error) error {
	if err != nil {
		belogs.Error(msg, err)
		if session != nil {
			if rollbackErr := session.Rollback(); rollbackErr != nil {
				belogs.Error("RollbackAndLogError(): rollback fail, msg:", msg, rollbackErr)
			}
		}
		return err
	}
	return nil
}

// 新增：补充engine关闭函数，避免程序退出时资源泄漏
func CloseXormEngine() error {
	if XormEngine != nil {
		return XormEngine.Close()
	}
	return nil
}
