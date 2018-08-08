package main

import (
	"time"
	"log"
	"strings"
	"strconv"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"encoding/binary"
	"net/http"
	_ "net/http/pprof"
	"git.leaniot.cn/publicLib/go-modbus"
	"gopkg.in/yaml.v2"
	"errors"
)

var config Config

type Table struct {
	Define string `json:"define"`
	Unit   string `json:"unit"`
	Type   string `json:"type"`
	Digits int    `json:"digits"`
}

//type TableNew struct {
//	Key    string      `json:"key"`
//	Define string      `json:"define"`
//	Unit   string      `json:"unit"`
//	Type   string      `json:"type"`
//	Digits int         `json:"digits"`
//	Data   interface{} `json:"data"`
//}
//type TableSend struct {
//	Key  string
//	Data interface{}
//}

type MessageSender struct {
	Data []map[string]interface{} `json:"data"`
}

type Device struct {
	Jsonfile string `yaml:"filename"`  //点表文件
	Address  string `yaml:"address"`   //设备地址
	SlaveId  byte   `yaml:"slave_id"`  //
	DeviceID string `yaml:"device_id"` //变频器设备ID
	Posturl  string `yaml:"post_url"`  //post到后端url
}

type Config struct {
	Device []Device `yaml:"device"`
}

func GetBit1(word []byte, bit uint16) bool {
	return uint(word[bit/8])>>(bit%8)&0x01 == 0x01
}

func DataPointTabler(output map[string]Table) {
	//解析json文件点表
	b, e := ioutil.ReadFile(config.Device[0].Jsonfile)
	if e != nil {
		panic(e)
	}
	if e = json.Unmarshal(b, &output); e != nil {
		panic(e)
	}
}

func String2Uint16(s string) uint16 {
	//将string类型转成uint16
	i, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		panic(err)
	}
	return uint16(i)
}

func GenModbusClient() (modbus.Client, error) {
	//建立modbusTCP连接
	handler := modbus.NewTCPClientHandler(config.Device[0].Address)
	handler.Timeout = 20 * time.Second
	handler.SlaveId = config.Device[0].SlaveId
	e := handler.Connect()
	if e != nil {
		log.Printf("%v", e)
		return nil, e
	}
	defer handler.Close()
	return modbus.NewClient(handler), nil
}

func PostJson(url string, b []byte) (*http.Response, error) {
	//post to server
	c := &http.Client{Timeout: 5 * time.Second,}
	reqNew := bytes.NewBuffer([]byte(b))
	req, _ := http.NewRequest("POST", url+"/", reqNew)
	req.Header.Add("Content-type", "application/json")
	return c.Do(req)
}

func ReadData(client modbus.Client, m map[string]Table) (MessageSender, error) {
	var MessageSendArray = make(map[string]interface{})
	var reg interface{}
	var stop error
	//根据点表通过modbusTCP从设备读取数据
	for key := range m {
		if strings.Contains(key, ".") {
			//read a bit
			address := strings.Split(key, ".")
			//addr := address[0] + "_" + address[1]
			register := String2Uint16(address[0])
			register_bit := String2Uint16(address[1])
			r, err := client.ReadHoldingRegisters(uint16(register), 1)
			if err != nil {
				stop = errors.New("error")
				log.Println(err)
				log.Println(key)
				break
			}
			reg = GetBit1(r, register_bit)
			MessageSendArray[key] = reg
		} else {
			//read 16-bits
			register := String2Uint16(key)
			r, err := client.ReadHoldingRegisters(uint16(register), 1)
			if err != nil {
				stop = errors.New("error")
				log.Println(err)
				log.Println(key)
				break
			}
			reg = r
			for i := 0; i < len(r); i += 2 { //byte to int
				reg = binary.BigEndian.Uint16(r[i : i+2])
			}
			MessageSendArray[key] = reg
		}
	}
	MessageSendArray["5000"] = config.Device[0].DeviceID //添加设备ID到“5000”字段
	messageSender := MessageSender{}

	messageSender.Data = append(messageSender.Data, MessageSendArray)
	log.Print(messageSender)
	return messageSender, stop
}

func SendData(messageSender MessageSender) {
	//向后端接口发送数据
	b, e := json.Marshal(messageSender) //序列化json
	if e != nil {
		log.Print(e)
	}
	//log.Println(string(b))
	rsp, e := PostJson(config.Device[0].Posturl, b)
	if e != nil {
		log.Printf("Send request to failed: %v", e)
		return
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 201 && rsp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(rsp.Body)
		log.Print(string(body))
	} else {
		log.Printf("Post to Success")
	}

}

func ConfigInit() {
	//程序的初始化函数
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	configContent, _ := ioutil.ReadFile("config.yml")
	yaml.Unmarshal(configContent, &config)
	//go func() {
	//	log.Println(http.ListenAndServe(":6060", nil))
	//}()
}

func main() {
	//初始化
	ConfigInit()

	//解析json点表
	PointTable := make(map[string]Table)
	DataPointTabler(PointTable)

	for {
		//建立连接
		client, e := GenModbusClient()
		if e != nil {
			log.Printf("%v", e)
			time.Sleep(time.Second * 5)
			continue
		}

		//读取数据和上传数据
		for {
			message, e:= ReadData(client, PointTable)
			if e != nil{
				log.Printf("Connct closed")
				break
			}
			SendData(message)
			time.Sleep(time.Second * 5)
		}
	}

}
