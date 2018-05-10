package main

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"strings"
)

type Output struct {
	Define string `json:"define"`
	Unit string `json:"unit"`
	Type string `json:"type"`
	Digits int `json:"digits"`
}

func testf(output map[string]Output)  {
//	var output = make(map[string]Output)
	b, e := ioutil.ReadFile("data_map.json")
	if e != nil {panic(e)}

	if e = json.Unmarshal(b, &output); e != nil {panic(e)}
}

func main(){
	test := make(map[string]Output)
	testf(test)
	for key, value := range test{
		if strings.Contains(key, "."){
			fmt.Println(key) 
		}else {
			fmt.Println(value)
		}
	}
//	fmt.Println(len(output), output)
}