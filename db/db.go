package db

import (
	"database/sql"
	belogs "github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	_ "strconv"

	. "conf"
)

var ConnDb = &sql.DB{}

func InitMySql() {
	user := Configure.String("mysql::user")
	password := Configure.String("mysql::password")
	server := Configure.String("mysql::server")
	database := Configure.String("mysql::database")

	openSql := user + ":" + password + "@tcp(" + server + ")/" + database
	belogs.Info("openSql is ", openSql)

	db, err := sql.Open("mysql", openSql) //sql.Open("mysql", "rpstir:Rpstir-123@tcp(192.168.138.135:3306)/rpstir")
	if err != nil {
		belogs.Error(err)
		return
	}
	ConnDb = db
	belogs.Info("ConnDb is ", ConnDb)
}

func TxCommitOrRollback(tx *sql.Tx) {
	if e := recover(); e != nil {
		belogs.Error("will rollback:", e)
		err := tx.Rollback()
		if err != sql.ErrTxDone && err != nil {
			belogs.Error(err)
		}
	} else {
		belogs.Debug("db will commit")
		tx.Commit()
	}
}

// 在update内部实现事务性，但如果调用方有大量的操作需要一个事务完成时，这样就不好了
func UpdateInsideTx(sqlStr string, args ...interface{}) error {
	// need drop
	tx, err := ConnDb.Begin()
	if err != nil {
		belogs.Error(err)
		return err
	}
	defer TxCommitOrRollback(tx)

	belogs.Info("UpdateWithTx():" + sqlStr)
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		panic(err)
	}
	return err

}

// 在update内部不事务性，在调用方有大量的操作需要一个事务完成时，调用方自己实现事务,传入事务的tx参数
func UpdateOutsideTx(tx *sql.Tx, sqlStr string, args ...interface{}) error {
	belogs.Info("UpdateNoTx():" + sqlStr)
	_, err := tx.Exec(sqlStr, args...)
	return err
}
