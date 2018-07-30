package main

import (
	"os"
	"bufio"
	"strings"
	"io"
)

type ClientId struct {
	ClientIds []string
}

func (this *ClientId) ReadLine(fileName string) error {
	list := []string{}
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	buf := bufio.NewReader(f)

	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			if err == io.EOF { //读取结束，会报EOF
				this.ClientIds = list
				return nil
			}
			return nil
		}
		list = append(list, line)
	}
}

func (this *ClientId) getClientId(index int) string {
	clientId := this.ClientIds[index];
	return clientId
}

func (this *ClientId)initClientId(fileName string) {
	this.ReadLine(fileName)
}

//func main() {
//	clientId := new(ClientId)
//	//list := []string{}
//	fmt.Println(1)
//	//clientId.ReadLine("/Users/zixuan.tian/GolandProjects/mqtt-benchmark/tls/client_ids")
//	clientId.initClientId("/Users/zixuan.tian/GolandProjects/mqtt-benchmark/tls/client_ids")
//	list := clientId.ClientIds;
//	fmt.Println(2)
//	for index, value := range list {
//		fmt.Printf("arr[%d]=%s \n", index, value)
//	}
//	//fmt.Println()
//	fmt.Println("finish load clientIds...")
//	//read()
//	//fmt.Println(list)
//}
