package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type NameServer struct {
	AgentIp      string
	NameServerIP string
}
type Dns struct {
	AgentIp    string
	DomainName string
	DomainIp   string
}
type MemStat struct {
	AgentIp   string
	MemStat   string
	TimeStamp string
}
type UpdateNameServer struct {
	Servers []struct {
		NameServerIP string
	}
}
type UpdateDns struct { //APIserver传过来的dns修改信息
	Servers []struct {
		DomainName string
		DomainIp   string
	}
}
type NameServerslice struct {
	Servers []NameServer
}
type Dnsslice struct {
	Servers []Dns
}

func GetMemTotal() int {
	cmd := exec.Command("/bin/sh", "-c", `cat /mem/meminfo | grep MemTotal`)
	// 获取输出对象，可以从该对象中读取输出结果
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	// 保证关闭输出流
	defer stdout.Close()
	// 运行命令
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	// 读取输出结果
	opBytes, err := ioutil.ReadAll(stdout)
	Pattern := `-?[1-9]\d*`
	findMatched, _ := regexp.Compile(Pattern)
	memInfo := findMatched.FindStringSubmatch(string(opBytes))
	if len(memInfo) > 0 {
		meminfo, _ := strconv.Atoi(memInfo[0])
		if err != nil {
			log.Fatal(err)
		}
		return meminfo
	}
	return -1
}
func GetMemFree() int {
	cmd := exec.Command("/bin/sh", "-c", `cat /mem/meminfo | grep MemFree`)
	// 获取输出对象，可以从该对象中读取输出结果
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	// 保证关闭输出流
	defer stdout.Close()
	// 运行命令
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	// 读取输出结果
	opBytes, err := ioutil.ReadAll(stdout)
	Pattern := `-?[1-9]\d*`
	findMatched, _ := regexp.Compile(Pattern)
	memInfo := findMatched.FindStringSubmatch(string(opBytes))
	if len(memInfo) > 0 {
		meminfo, _ := strconv.Atoi(memInfo[0])
		if err != nil {
			log.Fatal(err)
		}
		return meminfo
	}
	return -1
}
func nameServer(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		//将来自APIserver的nameserver信息解析出来
		err := r.ParseForm()
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}
		var nameServerTmp UpdateNameServer
		result, _ := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(result, &nameServerTmp)
		if err != nil {
			fmt.Println("json.Unmarshal is err:", err.Error())
		}
		//fmt.Println(nameServerTmp.Servers[0].NameServerIP)
		//将解析出的文件信息进行覆盖写入
		f, err := os.OpenFile("/etc/resolv.conf", os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println("file create failed. err: " + err.Error())
		} else {
			_, err = f.Write([]byte("nameserver " + nameServerTmp.Servers[0].NameServerIP + "\r\n"))
			//fmt.Println("write succeed!")
			f.Close()
		}
		//按行进行追加写入
		for i := 0; i < len(nameServerTmp.Servers); i++ {
			f, err := os.OpenFile("/etc/resolv.conf", os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println("file create failed. err: " + err.Error())
			} else {
				_, err = f.Write([]byte("nameserver " + nameServerTmp.Servers[i].NameServerIP + "\r\n"))
				//fmt.Println("write succeed!")

			}
		}
		f.Close()
	}
}

func dns(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}
		var dnsTmp UpdateDns
		result, _ := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(result, &dnsTmp)
		if err != nil {
			fmt.Println("json.Unmarshal is err:", err.Error())
		}
		//fmt.Println(dnsTmp.Servers[0].DomainName)
		//将解析出的文件信息进行覆盖写入
		f, err := os.OpenFile("/etc/hosts", os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println("file create failed. err: " + err.Error())
		} else {
			_, err = f.Write([]byte(dnsTmp.Servers[0].DomainIp + " " + dnsTmp.Servers[0].DomainName + "\r\n"))
			f.Close()
		}
		for i := 0; i < len(dnsTmp.Servers); i++ {
			f, err := os.OpenFile("/etc/hosts", os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println("file create failed. err: " + err.Error())
			} else {
				_, err = f.Write([]byte(dnsTmp.Servers[i].DomainIp + " " + dnsTmp.Servers[i].DomainName + "\r\n"))
				//fmt.Println("write succeed!")
			}
		}
		f.Close()

	}
}

func registerNameServer() {
	//获取本机Ip地址
	var agentIp string
	agentIp = getAgentIP()
	//将resolv文件信息按行读取并以json的格式进行发送到APIserver上
	var s NameServerslice
	userFile := "/etc/resolv.conf"
	fl, err := os.Open(userFile)
	if err != nil {
		fmt.Println(userFile, err)
		return
	}
	defer fl.Close()
	br := bufio.NewReader(fl)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		tmp := string(a)
		comma := strings.Index(tmp, " ")
		if tmp[0:comma+1] == "nameserver " {
			s.Servers = append(s.Servers, NameServer{AgentIp: agentIp, NameServerIP: tmp[comma+1:]})
		}
	}
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println("json err:", err)
	}
	body := bytes.NewBuffer([]byte(b))
	http.Post("http://10.10.28.62:8080/nameserver", "application/json;charset=utf-8", body)

}

func registerDns() {
	var agentIp string
	agentIp = getAgentIP()
	var s Dnsslice
	userFile := "/etc/hosts"
	fl, err := os.Open(userFile)
	if err != nil {
		fmt.Println(userFile, err)
		return
	}
	defer fl.Close()
	br := bufio.NewReader(fl)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		tmp := string(a)
		if len(tmp) != 0 {
			if tmp[0] != '#' && tmp[0] != ' ' { //
				leftIndex := strings.Index(tmp, " ")
				ipV4Pattern := `((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)`
				ip, _ := regexp.Compile(ipV4Pattern)
				matchedIP := string(ip.Find(a))
				if matchedIP == "" {
					matchedIP = tmp[:leftIndex] //因为ipv6的正则表达式太复杂，所以采用字符串查找的方式，查找第一个空格
				}
				//domainPattern := `[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}(\.[a-zA-Z0-9][a-zA-Z0-9_-]{0,62})*(\.[a-zA-Z][a-zA-Z0-9]{0,10}){1}|[A-Za-z][A-Za-z]*`
				//doMainName, _ := regexp.Compile(domainPattern)
				//matcheddMname := string(doMainName.Find(a))
				matcheddMname := strings.TrimSpace(tmp[leftIndex:])
				s.Servers = append(s.Servers, Dns{DomainName: matcheddMname, DomainIp: matchedIP, AgentIp: agentIp})
			}
		}

	}
	b, err := json.Marshal(s)
	if err != nil {
		fmt.Println("json err:", err)
	}
	body := bytes.NewBuffer([]byte(b))
	http.Post("http://10.10.28.62:8080/dns", "application/json;charset=utf-8", body)
}
func sendMemStat() {
	fmt.Println("afsafaf")
	time.LoadLocation("Asia/Shanghai")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 访问内存利用率信息并将信息发送到APIserver上
			var s MemStat
			var agentIp string
			agentIp = getAgentIP()
			total := GetMemTotal()
			free := GetMemFree()
			usedPersent := (total - free) * 100 / total
			memStat := strconv.Itoa(usedPersent)
			timeUnix := time.Now().Local().Format("2006-01-02 15:04:05")
			s = MemStat{
				AgentIp:   agentIp,
				MemStat:   memStat,
				TimeStamp: timeUnix,
			}
			b, err := json.Marshal(s)
			if err != nil {
				fmt.Println("json err:", err)
			}
			body := bytes.NewBuffer([]byte(b))
			http.Post("http://10.10.28.62:8080/memstat", "application/json;charset=utf-8", body)
			fmt.Println("看看时间", timeUnix)
		}
	}

}
func getAgentIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
func main() {
	//TODO
	// Check user
	// use config open apiserver url
	registerNameServer()                       //向APIserver上传nameserver信息
	registerDns()                              //向APIserver上传dns信息
	go sendMemStat()                           //隔一分钟向APIserver上传内存信息
	http.HandleFunc("/nameserver", nameServer) // APIserver修改nameserver
	http.HandleFunc("/dns", dns)               //APIserver修改dns信息
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
