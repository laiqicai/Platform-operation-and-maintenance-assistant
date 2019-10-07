package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)


type NameServer struct {
	Id    int
	AgentIp string
	NameServer string
}
type Dns struct {
	Id    int
	AgentIp  string
	DomainName  string
	DomainIp string
}
type MemStat struct {
	Id    int
	AgentIp string
	MemStat string
	TimeStamp time.Time `orm:"type(datetime)"`
}
func init() {
	// set default database
	orm.RegisterDataBase("default", "mysql", "root:laiqicai@tcp(127.0.0.1:3306)/laiqicai?charset=utf8&loc=Asia%2FShanghai", 30)
	// register model
	orm.RegisterModel(new(NameServer),new(Dns),new(MemStat))
	// create table
	orm.RunSyncdb("default", false, true)
}
