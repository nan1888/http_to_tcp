package main

//服务器端
import (
	"encoding/json"
	"fmt"
	"log"
	"net" //支持通讯的包
	"net/http"

	"github.com/gin-gonic/gin"
)

type connmsg struct {
	Id string
}

var conn_gw map[string]net.Conn

//开始服务器
func startServer() {
	//连接主机、端口，采用ｔｃｐ方式通信，监听8987端口
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":8987")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	fmt.Println("start success!")
	conn_gw = make(map[string]net.Conn)
	for {
		//等待客户端接入
		conn, err := listener.Accept()
		checkError(err)
		//开一个goroutines处理客户端消息，这是golang的特色，实现并发就只go一下就好
		go doServerStuff(conn)
	}
}

//处理客户端消息
func doServerStuff(conn net.Conn) {
	connmsgs := connmsg{}
	fmt.Println("client connect")

	buf := make([]byte, 100)
	n, _ := conn.Read(buf) //读取客户机发的消息
	jsonerr := json.Unmarshal(buf[0:n], &connmsgs)
	if jsonerr != nil {
		fmt.Println("error:", jsonerr)
		return
	}
	/*把conn和gwid入库*/
	conn_gw[connmsgs.Id] = conn
	_, err := conn.Write([]byte("{\"errcode\":\"0\",\"errmsg\":\"success\"}\n"))
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
	}
	fmt.Println(connmsgs.Id)

}

func start_httpserver() {
	router := gin.Default()

	router.POST("/test1", func(c *gin.Context) {
		post_id := c.PostForm("id")
		post_msg := c.PostForm("msg")
		post_type := "0"
		conn, ok := conn_gw[post_id]
		if ok {
			fmt.Println(post_id)
			charg_json := "{\"msg\":\"" + post_msg + "\","
			charg_json = charg_json + "\"msgtype\":\"" + post_type + "\","
			charg_json = charg_json + "\"id\":\"" + post_id + "\"}\n"
			_, err := conn.Write([]byte(charg_json))
			if err != nil {
				fmt.Println("Can't resolve address: ", err)
			}
			fmt.Println(charg_json)
			c.String(http.StatusOK, "ok")
		} else {
			c.String(http.StatusOK, "can not find gwid,please check gwid")
		}

	})
	router.Run(":8083")
}

//检查错误
func checkError(err error) int {
	if err != nil {
		if err.Error() == "EOF" {
			return 0
		}
		log.Fatal("an error!", err.Error())
		return -1
	}
	return 1
}
func main() {
	//开启服务
	conn_gw = make(map[string]net.Conn)
	go start_httpserver()
	startServer()
}
