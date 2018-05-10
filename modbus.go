package main

import (
	"git.leaniot.cn/publicLib/go-modbus"
	"time"
	"log"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"strings"
	"strconv"
)

type Output struct {
	Define string `json:"define"`
	Unit string `json:"unit"`
	Type string `json:"type"`
	Digits int `json:"digits"`
}

func GetBit1(word []byte, bit uint) bool {
	return uint(word[bit/8])>>(bit%8)&0x01 == 0x01
}

func data_out(output map[string]Output)  {
//	var output = make(map[string]Output)
	b, e := ioutil.ReadFile("data_map.json")
	if e != nil {panic(e)}

	if e = json.Unmarshal(b, &output); e != nil {panic(e)}
	
//	return output
//	fmt.Println(len(output), output)
}

func string2uint(y string) uint16 {
    //strconv.Atoi 就是将 string 类型 转成 int
   	i, err := strconv.ParseUint(y, 10, 16) 
    if err != nil {
        panic(err)
    }
    return uint16(i)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//handler := modbus.NewTCPClientHandler("172.16.15.131:502")
	handler := modbus.NewTCPClientHandler("192.168.1.100:502")
	handler.Timeout = 10 * time.Second
	handler.SlaveId = 0x01	
	e := handler.Connect()
	if e != nil {
		log.Fatalf("%v", e)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)

	data_get := make(map[string]Output)
	data_out(data_get)
//	fmt.Println(data_get)
	for key/*, value*/ := range data_get{
		if strings.Contains(key, "."){
			fmt.Println(key) 
		}else {
//			fmt.Println(value)
			ikey := string2uint(key) 
			r, err := client.ReadHoldingRegisters(uint16(ikey), 1)
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("%d",r)
		}
	}
//	r, err := client.ReadHoldingRegisters(8000, 14)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//	log.Printf("%d",r)

}
