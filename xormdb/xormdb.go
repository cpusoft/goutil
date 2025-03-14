package xormdb

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/stringutil"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var XormEngine = &xorm.Engine{}

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
	//DB, err = sql.Open("mysql", "rpstir:Rpstir-123@tcp(202.173.9.21:13306)/rpstir")

	openSql := user + ":" + password + "@tcp(" + server + ")/" + database + "?charset=utf8&parseTime=True&loc=Local"
	logName := filepath.Base(os.Args[0])
	belogs.Info("InitMySqlParameter(): server is: ", server, database, logName)

	//连接数据库
	engine, err = xorm.NewEngine("mysql", openSql)
	if err != nil {
		belogs.Error("InitMySqlParameter(): NewEngine failed, err:", err)
		return engine, err
	}
	//连接测试
	if err := engine.Ping(); err != nil {
		belogs.Error("InitMySqlParameter(): Ping failed, err:", err)
		return engine, err
	}

	//设置连接池的空闲数大小
	engine.SetMaxIdleConns(maxidleconns)
	//设置最大打开连接数
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

func InitSqliteParameter(filepath string, maxidleconns, maxopenconns int) (engine *xorm.Engine, err error) {

	belogs.Info("InitSqliteParameter(): filepath: ", filepath)

	//连接数据库
	engine, err = xorm.NewEngine("sqlite3", filepath)
	if err != nil {
		belogs.Error("InitSqliteParameter(): NewEngine failed, filepath:", filepath, err)
		return engine, err
	}
	//连接测试
	if err := engine.Ping(); err != nil {
		belogs.Error("InitSqliteParameter(): Ping failed, err:", err)
		return engine, err
	}

	//设置连接池的空闲数大小
	engine.SetMaxIdleConns(maxidleconns)
	//设置最大打开连接数
	engine.SetMaxOpenConns(maxopenconns)
	// show sql
	//engine.ShowSQL(true)
	engine.SetTableMapper(names.SnakeMapper{})

	return engine, nil

}

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
	// db, err := xorm.NewPostgreSQL("postgres://postgres:123@localhost:5432/test?sslmode=disable")
	host, port, err := net.SplitHostPort(server)
	if err != nil {
		belogs.Error("InitPostgreSQLParameter(): SplitHostPort fail, server:", server, err)
		return nil, err
	}
	str := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)
	logName := filepath.Base(os.Args[0])
	belogs.Info("InitPostgreSQLParameter(): server:", server,
		"  database:", database, "  logName:", logName)

	//连接数据库
	engine, err = xorm.NewEngine("postgres", str)
	if err != nil {
		belogs.Error("InitPostgreSQLParameter(): NewEngine failed, err:", err)
		return engine, err
	}
	//连接测试
	if err := engine.Ping(); err != nil {
		belogs.Error("InitPostgreSQLParameter(): Ping failed, err:", err)
		return engine, err
	}

	//设置连接池的空闲数大小
	engine.SetMaxIdleConns(maxidleconns)
	//设置最大打开连接数
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

// get new session, and begin session
func NewSession() (*xorm.Session, error) {
	// open mysql session
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

// when session is error, will rollback and log the error
func RollbackAndLogError(session *xorm.Session, msg string, err error) error {
	if err != nil {
		belogs.Error(msg, err)
		if session != nil {
			session.Rollback()
		}
		return err
	}
	return nil
}

func SqlNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func SqlNullInt(s int64) sql.NullInt64 {
	if s < 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: s,
		Valid: true,
	}
}

// Deprecated: should using stringutil.Int64sToInString
func Int64sToInString(s []int64) string {
	return stringutil.Int64sToInString(s)
}
