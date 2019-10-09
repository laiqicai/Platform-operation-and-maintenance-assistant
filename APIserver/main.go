package main

import (
	_ "APIserver/controllers"
	_ "APIserver/models"
	_ "APIserver/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}

