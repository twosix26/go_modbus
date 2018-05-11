package main

import (
	"git.leaniot.cn/publicLib/go-modbus"
	"time"
	"log"
	"encoding/json"
	"io/ioutil"
	"strings"
	"strconv"
	"gopkg.in/yaml.v2"
)
var config Config

type Table struct {
	Define	string 	`json:"define"`
	Unit 	string 	`json:"unit"`
	Type 	string 	`json:"type"`
	Digits 	int 	`json:"digits"`
}

type Device struct {
	Address string  `yaml:"address"`
	SlaveId	byte    `yaml:"slave_id"`
}

type Config struct {
	Device []Device `yaml:"device"`
}

func GetBit1(word []byte, bit uint16) bool {
	return uint(word[bit/8])>>(bit%8)&0x01 == 0x01
}

func DataPointTabler(output map[string]Table)  {
	//解析json文件点表
	b, e := ioutil.ReadFile("data_map.json")
	if e != nil {panic(e)}
	if e = json.Unmarshal(b, &output); e != nil {panic(e)}
}

func String2Uint16(s string) uint16 {
    //将string类型转成uint16
   	i, err := strconv.ParseUint(s, 10, 16) 
    if err != nil {
        panic(err)
    }
    return uint16(i)
}

func GenModbusClient() (modbus.Client, error){
	//建立modbusTCP连接
	handler := modbus.NewTCPClientHandler(config.Device[0].Address)
	handler.Timeout = 10 * time.Second
	handler.SlaveId = config.Device[0].SlaveId	
	e := handler.Connect()
	if e != nil {
		log.Fatalf("%v", e)
		return nil, e
	}
	defer handler.Close()
	return modbus.NewClient(handler), nil
}

func ReadData(client modbus.Client, m map[string]Table) {
	//根据点表通过modbusTCP从设备读取数据
	for key, value := range m{
		if strings.Contains(key, "."){	
			//read a bit
			address := strings.Split(key,".")
			register := String2Uint16(address[0])
			register_bit := String2Uint16(address[1])
			r, err := client.ReadHoldingRegisters(uint16(register), 1)
			if err != nil {
				log.Println(err)
				return
			}
			b := GetBit1(r, register_bit)
			log.Printf("%s : %t", value.Define, b) 
		}else {							
			//read 16-bits
			register := String2Uint16(key) 
			r, err := client.ReadHoldingRegisters(uint16(register), 1)
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("%s : %d", value.Define, r)
		}
	}
}

func ConfigInit() {
	//程序的初始化函数
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	configContent, _ := ioutil.ReadFile("config.yml")
	yaml.Unmarshal(configContent, &config)

}

func main() {
	//初始化
	ConfigInit()
	
	//建立连接
	client, e := GenModbusClient()
	if e != nil{
		log.Fatalf("%v", e)
		return
	}

	//解析json点表
	PointTable := make(map[string]Table)
	DataPointTabler(PointTable)
	
	//读取数据
	for{
		ReadData(client, PointTable)
		time.Sleep(time.Second * 2)	
	}
	
}
