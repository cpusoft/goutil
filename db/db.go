package db

import (
	"database/sql"
	belogs "github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"

	conf "github.com/cpusoft/goutil/conf"
)

var ConnDb = &sql.DB{}

// Obsolete, suggest use "github.com/cpusoft/goutil/xormdb
func InitMySql() {
	user := conf.String("mysql::user")
	password := conf.String("mysql::password")
	server := conf.String("mysql::server")
	database := conf.String("mysql::database")

	openSql := user + ":" + password + "@tcp(" + server + ")/" + database
	belogs.Info("InitMySql(): server is :", server, database)

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

// update sql, control transaction in func
func UpdateInsideTx(sqlStr string, args ...interface{}) error {
	// need drop
	tx, err := ConnDb.Begin()
	if err != nil {
		belogs.Error(err)
		return err
	}
	defer TxCommitOrRollback(tx)

	belogs.Debug("UpdateWithTx():" + sqlStr)
	_, err = tx.Exec(sqlStr, args...)
	if err != nil {
		panic(err)
	}
	return err

}

// // update sql, control transaction out func
func UpdateOutsideTx(tx *sql.Tx, sqlStr string, args ...interface{}) error {
	belogs.Debug("UpdateNoTx():" + sqlStr)
	_, err := tx.Exec(sqlStr, args...)
	return err
}
