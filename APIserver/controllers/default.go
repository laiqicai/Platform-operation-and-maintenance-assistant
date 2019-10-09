package controllers

import (
	"APIserver/models"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type NameServerController struct {
	beego.Controller
}

type DnsController struct {
	beego.Controller
}

type MemStatController struct {
	beego.Controller
}

type UpdateNameServerTmp struct {//解析客户端发送的修改nameserver的json
	Servers []struct {
		AgentIP          string
		BeforeNameServer string
		AfterNameServer  string
	}
}

type UpdateDnsTmp struct { //解析客户端发送的修改dns的json
	Servers []struct {
		AgentIP         string
		AfterDomainName string
		AfterDomainIP   string
	}
}

type NameServerTmp struct { //解析agent发送的nameserver的json信息
	Servers []struct {
		AgentIP      string `json:"AgentIp"`
		NameServerIP string `json:"NameServerIP"`
	} `json:"Servers"`
}
type DnsTmp struct { //解析agent发送的dns的json信息
	Servers []struct {
		AgentIP    string `json:"AgentIp"`
		DomainName string `json:"DomainName"`
		DomainIP   string `json:"DomainIp"`
	} `json:"Servers"`
}
type MemTmp struct { //解析agent发送的memstat的json信息
	AgentIp   string
	MemStat   string
	TimeStamp string
}
type DnsRes struct { //返回客户端查询的agent的DNS信息
	AgentIp    string
	DomainName string `orm:"column(domain_name)"`
	DomainIP   string `orm:"column(domain_ip)"`
}
type NameServerRes struct { //返回客户端查询的nameserver信息
	AgentIp    string
	NameServer string `json:"NameServerIP"`
}

type MemRes struct { //返回客户端查询的agent的memstat信息
	AgentIp               string
	MemUtilizationAverage int
}
type MemStat struct { //从数据库中查询出的内存信息的临时存储结构
	MemStat string
}

type Dnsslice struct { //APIserver向agent发送dns修改信息
	Servers []DnsRes
}

type Memslice struct {
	Servers []MemRes
}

type NameServerslice struct {
	Servers []NameServerRes
}

func (this *NameServerController) Post() { //处理来自agent的nameserver信息，存储到数据库
	var nameServerTmp NameServerTmp
	data := this.Ctx.Input.RequestBody
	err := json.Unmarshal(data, &nameServerTmp)
	if err != nil {
		fmt.Println("json.Unmarshal is err:", err.Error())
	}
	for i := 0; i < len(nameServerTmp.Servers); i++ {
		var nameServer models.NameServer
		o := orm.NewOrm()
		nameServer.AgentIp = nameServerTmp.Servers[i].AgentIP
		nameServer.NameServer = nameServerTmp.Servers[i].NameServerIP
		id, err := o.Insert(&nameServer)
		if err == nil {
			fmt.Println(id)
		}
	}
}
func (this *DnsController) Post() { //处理来自agent的dns信息，存储到数据库
	var dnsTmp DnsTmp
	data := this.Ctx.Input.RequestBody
	err := json.Unmarshal(data, &dnsTmp)
	if err != nil {
		fmt.Println("json.Unmarshal is err:", err.Error())
	}
	for i := 0; i < len(dnsTmp.Servers); i++ {
		var dns models.Dns
		o := orm.NewOrm()
		dns.AgentIp = dnsTmp.Servers[i].AgentIP
		dns.DomainIp = dnsTmp.Servers[i].DomainIP
		dns.DomainName = dnsTmp.Servers[i].DomainName
		id, err := o.Insert(&dns)
		if err == nil {
			fmt.Println(id)
		}
	}
}
func (this *MemStatController) Post() { //处理来自agent的mem信息，存到数据库
	var memTmp MemTmp
	data := this.Ctx.Input.RequestBody
	err := json.Unmarshal(data, &memTmp)
	if err != nil {
		fmt.Println("json.Unmarshal is err:", err.Error())
	}
	local, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", memTmp.TimeStamp, local)
	var mem models.MemStat
	o := orm.NewOrm()
	mem.AgentIp = memTmp.AgentIp
	mem.MemStat = memTmp.MemStat
	mem.TimeStamp = t
	id, err := o.Insert(&mem)
	if err == nil {
		fmt.Println(id)
	}

}
func (this *NameServerController) Get() { //客户查询nameserver功能
	keys, ok := this.Ctx.Request.URL.Query()["agentip"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}
	var s NameServerslice
	//fmt.Println(len(agentIpTmp.Servers))
	//var wg sync.WaitGroup
	if len(keys) != 0 {
		//wg.Add(len(keys))
		//for i := 0;i < len(agentIpTmp.Servers);i++  {
		for i, _ := range keys {
			//println("AgentIPInindex: %v", i)
			//go func(){
			//defer wg.Done()
			var nameServers []NameServerRes
			//println("AgentIPInindex: %v", i)
			agentIP := keys[i]
			//println("AgentIP: %v", agentIP)
			o := orm.NewOrm()
			_, err := o.Raw("SELECT agent_ip, name_server FROM name_server WHERE agent_ip = ? ", agentIP).QueryRows(&nameServers)
			//fmt.Println(agentIP)
			if err != nil {
				fmt.Printf("Quary NameServer, agentIP: %v, err:%v", agentIP, err)
			}
			//fmt.Println(nameServers[i].NameServer)

			for i, _ := range nameServers {
				s.Servers = append(s.Servers, NameServerRes{NameServer: nameServers[i].NameServer, AgentIp: agentIP})
			}
			//}()

		}
	}
	mystruct := &s
	this.Data["json"] = mystruct
	this.ServeJSON()
}
func (this *DnsController) Get() { //客户查询agent的dns信息
	keys, ok := this.Ctx.Request.URL.Query()["agentip"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}
	var s Dnsslice
	for i := 0; i < len(keys); i++ {
		agentIP := keys[i]
		o := orm.NewOrm()
		var dns []DnsRes
		_, err := o.Raw("SELECT domain_name,domain_ip,agent_ip FROM dns WHERE agent_ip = ? ", agentIP).QueryRows(&dns)
		if err != nil {
			fmt.Printf("Quary dns, agentIP: %v, err:%v", agentIP, err)
		}
		for i := 0; i < len(dns); i++ {
			s.Servers = append(s.Servers, DnsRes{DomainIP: dns[i].DomainIP, DomainName: dns[i].DomainName, AgentIp: agentIP})
		}
	}
	mystruct := &s
	this.Data["json"] = mystruct
	this.ServeJSON()

}
func (this *MemStatController) Get() { //客户查询agent的内存信息
	agentips, ok := this.Ctx.Request.URL.Query()["agentip"]
	if !ok || len(agentips[0]) < 1 {
		log.Println("Url Param 'agentip' is missing")
		return
	}
	starttimes, ok := this.Ctx.Request.URL.Query()["starttime"]
	if !ok || len(starttimes[0]) < 1 {
		log.Println("Url Param 'starttime' is missing")
		return
	}
	endtimes, ok := this.Ctx.Request.URL.Query()["endtime"]
	if !ok || len(endtimes[0]) < 1 {
		log.Println("Url Param 'endtime' is missing")
		return
	}
	var s Memslice
	for i := 0; i < len(agentips); i++ {
		agentIP := agentips[i]
		startTime := starttimes[i]
		endTime := endtimes[i]
		o := orm.NewOrm()
		var memstat []MemStat
		_, err := o.Raw("select mem_stat from mem_stat where time_stamp>=? and time_stamp < ? and agent_ip = ? ", startTime, endTime, agentIP).QueryRows(&memstat)
		if err != nil {
			fmt.Printf("Quary dns, agentIP: %v, err:%v", agentIP, err)
		}
		sum := 0
		for i := 0; i < len(memstat); i++ {
			tmp, _ := strconv.Atoi(memstat[i].MemStat)
			sum = sum + tmp
		}
		if len(memstat) > 0 {
			average := sum / len(memstat)
			s.Servers = append(s.Servers, MemRes{AgentIp: agentIP, MemUtilizationAverage: average})

		}
	}
	mystruct := &s
	this.Data["json"] = mystruct
	this.ServeJSON()

}
func (this *NameServerController) Put() { //APIserver 根据客户端修改信息修改nameServer，并发送至相应的agent对其进行信息的修改
	var updateNameServerTmp UpdateNameServerTmp
	data := this.Ctx.Input.RequestBody
	err := json.Unmarshal(data, &updateNameServerTmp)
	if err != nil {
		fmt.Println("json.Unmarshal is err:", err.Error())
	}
	for i := 0; i < len(updateNameServerTmp.Servers); i++ {
		agentIP := updateNameServerTmp.Servers[i].AgentIP
		beforeNameServer := updateNameServerTmp.Servers[i].BeforeNameServer
		AfterNameServer := updateNameServerTmp.Servers[i].AfterNameServer
		o := orm.NewOrm()
		p, err := o.Raw("UPDATE name_server SET name_server = ? WHERE name_server = ? and agent_ip = ?").Prepare()
		if err != nil {
			fmt.Printf("Quary dns, agentIP: %v, err:%v", agentIP, err)
		}
		res, err := p.Exec(AfterNameServer, beforeNameServer, agentIP)
		defer p.Close()
		fmt.Println(reflect.TypeOf(res))
		if err == nil { //表示的是APIserver修改自身数据库成功后，将信息进行相应的传递到agent上
			o := orm.NewOrm()
			var nameServers []NameServerRes
			_, err := o.Raw("SELECT name_server FROM name_server WHERE agent_ip = ? ", agentIP).QueryRows(&nameServers)
			if err != nil {
				fmt.Printf("Quary NameServer, agentIP: %v, err:%v", agentIP, err)
			}
			var s NameServerslice
			for i := 0; i < len(nameServers); i++ {
				s.Servers = append(s.Servers, NameServerRes{NameServer: nameServers[i].NameServer})
			}
			b, err := json.Marshal(s)
			if err != nil {
				fmt.Println("json err:", err)
			}
			body := bytes.NewBuffer([]byte(b))
			http.Post("http://"+agentIP+":9090/nameserver", "application/json;charset=utf-8", body)
			fmt.Println("http://" + agentIP + ":9090/nameserver")

		}
	}

}
func (this *DnsController) Put() { //APIserver根据客户端发送信息修改自身的dns信息，并发送至agent进行修改自身dns信息
	var updateDnsTmp UpdateDnsTmp
	data := this.Ctx.Input.RequestBody
	err := json.Unmarshal(data, &updateDnsTmp)
	if err != nil {
		fmt.Println("json.Unmarshal is err:", err.Error())
	}
	for i := 0; i < len(updateDnsTmp.Servers); i++ {
		agentIP := updateDnsTmp.Servers[i].AgentIP
		AfterDomainName := updateDnsTmp.Servers[i].AfterDomainName
		AfterDomainIP := updateDnsTmp.Servers[i].AfterDomainIP
		fmt.Println(agentIP)
		fmt.Println(AfterDomainIP)
		o := orm.NewOrm()
		p, err := o.Raw("update dns set domain_ip = ? ,domain_name = ? where agent_ip = ? and(domain_ip = ? or domain_name = ? )").Prepare()
		if err != nil {
			fmt.Printf("Quary dns, agentIP: %v, err:%v", agentIP, err)
		}
		res, err := p.Exec(AfterDomainIP, AfterDomainName, agentIP, AfterDomainIP, AfterDomainName)
		defer p.Close()
		fmt.Println(res)
		if err == nil { //APIserver修改成功
			o := orm.NewOrm()
			var dns []DnsRes
			_, err := o.Raw("SELECT domain_name,domain_ip FROM dns WHERE agent_ip = ? ", agentIP).QueryRows(&dns)
			if err != nil {
				fmt.Printf("Quary NameServer, agentIP: %v, err:%v", agentIP, err)
			}
			var s Dnsslice
			for i := 0; i < len(dns); i++ {
				s.Servers = append(s.Servers, DnsRes{DomainName: dns[i].DomainName, DomainIP: dns[i].DomainIP})
			}
			b, err := json.Marshal(s)
			if err != nil {
				fmt.Println("json err:", err)
			}
			body := bytes.NewBuffer([]byte(b))
			http.Post("http://"+agentIP+":9090/dns", "application/json;charset=utf-8", body)
		}
	}

}
