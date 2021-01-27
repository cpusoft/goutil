package dbconf

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/conf"
	_ "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

/*
CREATE TABLE *** (
	id int(10) unsigned NOT NULL primary key auto_increment,
	section varchar(128)  NOT NULL COMMENT 'section',
	myKey varchar(128)  NOT NULL  COMMENT 'key',
	myValue varchar(1024)  NOT NULL  COMMENT 'value',
	defaultMyValue varchar(1024)  NOT NULL  COMMENT 'default value',
	updateTime datetime NOT NULL COMMENT 'update time',
	unique sectionMyKey (section,myKey)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin comment='rpstir2 configuration'
*/
type dbConfModel struct {
	Id             uint64 `json:"id"  xorm:"id"`
	Section        string `json:"section"  xorm:"section"`
	MyKey          string `json:"myKey"  xorm:"myKey"`
	MyValue        string `json:"myValue"  xorm:"myValue"`
	DefaultMyValue string `json:"defaultMyValue"  xorm:"defaultMyValue"`
}

// first call xormdb.InitMySql()
// then call this dbConf.InitDbConf(tableName)
func InitDbConf(tableName string) (err error) {
	// start mysql
	dbConfModels := make([]dbConfModel, 0)
	sql :=
		`select id, section, myKey, myValue, DefaultMyValue
	    from ` + tableName + `    
		order by id `
	err = xormdb.XormEngine.Sql(sql).Find(&dbConfModels)
	if err != nil {
		belogs.Error("InitDbConf(): find fail:", err)
		return err
	}
	for i := range dbConfModels {
		err = conf.SetString(dbConfModels[i].Section+"::"+dbConfModels[i].MyKey, dbConfModels[i].MyValue)
		if err != nil {
			belogs.Error("InitDbConf(): SetString fail:", jsonutil.MarshalJson(dbConfModels[i]), err)
			// not return ,just skip to next
		}
	}
	return nil
}

func SetString(tableName, section, myKey, myValue string) (err error) {
	session, err := xormdb.NewSession()
	defer session.Close()

	now := time.Now()
	sql := `update ` + tableName + ` set myValue=?, updateTime=? 
		where section=? and myKey=? `
	_, err = session.Exec(sql, myValue, now, section, myKey)
	if err != nil {
		belogs.Error("SetString(): update fail:", tableName, section, myKey, myValue, err)
		return xormdb.RollbackAndLogError(session, "SetString(): update fail:"+
			tableName+","+section+","+myKey+","+myValue+", fail: ", err)
	}
	err = conf.SetString(section+"::"+myKey, myValue)
	if err != nil {
		belogs.Error("SetString(): conf.SetString fail:", tableName, section, myKey, myValue, err)
		return xormdb.RollbackAndLogError(session, "SetString(): conf.SetString fail:"+
			tableName+","+section+","+myKey+","+myValue+", fail: ", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("SetString(): CommitSession fail :", tableName, section, myKey, myValue, err)
		return xormdb.RollbackAndLogError(session, "SetString(): CommitSession fail:"+
			tableName+","+section+","+myKey+","+myValue+", fail: ", err)
	}
	return
}

func String(key string) string {
	return conf.String(key)
}

func Int(key string) int {
	return conf.Int(key)
}

func Strings(key string) []string {
	return conf.Strings(key)
}

func Bool(key string) bool {
	return conf.Bool(key)
}

func DefaultBool(key string, defaultVal bool) bool {
	return conf.DefaultBool(key, defaultVal)
}

//destpath=${rpstir2::datadir}/rsyncrepo   --> replace ${rpstir2::datadir}
//-->/root/rpki/data/rsyncrepo --> get /root/rpki/data/rsyncrepo
func VariableString(key string) string {
	return conf.VariableString(key)
}
