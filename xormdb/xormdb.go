package xormdb

import (
	_ "encoding/json"
	_ "strconv"

	belogs "github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"

	conf "github.com/cpusoft/goutil/conf"
)

var XormEngine = &xorm.Engine{}

func InitMySql() error {
	//DB, err = sql.Open("mysql", "rpstir:Rpstir-123@tcp(202.173.9.21:13306)/rpstir")
	user := conf.String("mysql::user")
	password := conf.String("mysql::password")
	server := conf.String("mysql::server")
	database := conf.String("mysql::database")
	maxidleconns := conf.Int("mysql::maxidleconns")
	maxopenconns := conf.Int("mysql::maxopenconns")

	openSql := user + ":" + password + "@tcp(" + server + ")/" + database
	belogs.Info("InitMySql() is: ", openSql)

	//连接数据库
	engine, err := xorm.NewEngine("mysql", openSql)
	if err != nil {
		belogs.Error("NewEngine failed: ", err)
		return err
	}
	//连接测试
	if err := engine.Ping(); err != nil {
		belogs.Error("Ping failed: ", err)
		return err
	}
	//设置连接池的空闲数大小
	engine.SetMaxIdleConns(maxidleconns)
	//设置最大打开连接数
	engine.SetMaxOpenConns(maxopenconns)
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
	engine.SetTableMapper(core.SnakeMapper{})
	XormEngine = engine
	return nil

}
