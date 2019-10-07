package routers

import (
	"APIserver/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/nameserver", &controllers.NameServerController{})
	beego.Router("/dns", &controllers.DnsController{})
	beego.Router("/memstat", &controllers.MemStatController{})
}
