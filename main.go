package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tarm/serial"
	_serial "go.bug.st/serial"
	"log"
	"net/http"
	"strconv"
)

var (
	configFileName = "config.conf"
	ports          []string // 设备中的串口列表
	portSet        map[string]struct{}

	flagVar = struct {
		Server struct {
			Host string `default:"127.0.0.1" required:"true" env:"ip"`
			Port int    `default:"8080" required:"true"`
		}
		Serial struct {
			Name string
			Baud string
		}
	}{}

	// WebSocket
	upgrade = websocket.Upgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func init() {
	fmt.Println(`命令执行参数说明：
    -host 访问地址
    -p 端口，默认8080
    -sn 串口名称
    -sb 波特率，典型值：110,300,600,1200,2400,4800,9600,14400,19200,38400,43000,56000,57600,115200,128000,256000`)

	// 初始化命令行参数定义
	flag.StringVar(&flagVar.Server.Host, "host", "127.0.0.1", "http service host")
	flag.IntVar(&flagVar.Server.Port, "p", 8080, "http service port")
	flag.StringVar(&flagVar.Serial.Name, "sn", "", "串口名称")
	// https://blog.csdn.net/qq_40147893/article/details/106539081
	flag.StringVar(&flagVar.Serial.Baud, "sb", "", "波特率")
	flag.Parse()

	// https://blog.csdn.net/linxue110/article/details/107778845
	var err error
	ports, err = _serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("没有找到串口")
	}
	m, err := ToMapSetStrictE(ports)
	if err == nil {
		portSet = m.(map[string]struct{})
	}
}

func scan(tip string, vars ...any) {
	fmt.Println(tip)
	scan, err := fmt.Scan(vars...)
	if err != nil {
		log.Fatal(err)
	}
	if scan != len(vars) {
		log.Fatal("请输入完整参数")
	}
	//fmt.Sscan() // 从指定字符串中读取数据
	//fmt.Fscan() // 从io.Reader中读取数据
	/*var msg string
	  reader := bufio.NewReader(os.Stdin) // 标准输入输出
	  msg, _ = reader.ReadString('\n')    // 回车结束
	  msg = strings.TrimSpace(msg)        // 去除最后一个空格
	  fmt.Printf(msg)*/
}

// 命令窗口输入串口名称
func scanSerialName() string {
	if ports == nil || len(ports) == 0 {
		log.Fatal("没有找到串口")
	}
	for _, port := range ports {
		log.Printf("找到串口: %v\n", port)
	}
	var str string
	scan("请输入串口名称：", &str)
	return str
}

// 命令窗口输入串口波特率
func scanSerialBaud() string {
	var baudStr string
	scan("请输入串口波特率：", &baudStr)
	var err error
	_, err = strconv.Atoi(baudStr)
	if err != nil {
		scanSerialBaud()
	}
	if err != nil {
		fmt.Println("请输入数字")
		scanSerialBaud()
	}
	return baudStr
}

func connectSerialPort(fc func(data []byte)) {
	baud, err := strconv.Atoi(flagVar.Serial.Baud)
	if err != nil {
		log.Fatal(err)
	}
	/*mode := &_serial.Mode{
	  	BaudRate: baud,
	  }
	  port, err := _serial.Open(flagVar.Serial.Name, mode)
	  if err != nil {
	  	log.Fatal(err)
	  }
	  log.Println(port)*/

	// https://blog.csdn.net/weixin_45904051/article/details/123218647
	// 设置串口编号
	ser := &serial.Config{Name: flagVar.Serial.Name, Baud: baud}
	// 打开串口
	conn, err := serial.OpenPort(ser)
	if err != nil {
		log.Fatal(err)
	}
	// 启动一个协程循环发送
	/*go func() {
		for {
			revData := []byte("123456")
			_, err := conn.Write(revData)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("Tx:%X \n", revData)
			time.Sleep(time.Second)
		}
	}()*/

	// 保持数据持续接收
	for {
		buf := make([]byte, 1024)
		lens, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		//str:= hex.EncodeToString(buf)
		revData := buf[:lens]
		log.Printf("Rx:%X \n", revData)

		fc(revData)
	}
}

func main() {
	initConfig()

	log.SetFlags(0)
	http.Handle("/", http.FileServer(http.FS(index)))
	http.HandleFunc("/ws", echo)
	err := http.ListenAndServe(flagVar.Server.Host+":"+strconv.Itoa(flagVar.Server.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	// 完成和Client HTTP >>> WebSocket 的协议升级
	c, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)
	// 向任何连接的 WebSocket 客户端发送消息
	err = c.WriteMessage(websocket.TextMessage, []byte("Hi Client!"))
	if err != nil {
		log.Fatalln(err)
	}
	for {
		// 接收客户端message
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		switch mt {
		case websocket.TextMessage:
			data, err := json.Marshal(message)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(data)
		case websocket.BinaryMessage:

		case websocket.CloseMessage:
			err := c.Close()
			if err != nil {
				log.Fatal(err)
			}
		case websocket.PingMessage:

		case websocket.PongMessage:

		default:
			fmt.Print("========default================")
		}
		// 向客户端发送message
		connectSerialPort(func(data []byte) {
			err = c.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("write:", err)
			}
		})
	}
}
