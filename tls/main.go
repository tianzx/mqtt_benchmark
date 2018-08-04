package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const BASE_TOPIC string = "/mqtt-bench/benchmark"

var Debug bool = false

var DefaultHandlerResults []*SubscribeResult

var mapResult = make(map[string]int)

var tlsConfig *tls.Config

var infoLog *log.Logger

type SubscribeResult struct {
	Count int // 订阅结果
}

type CertConfig interface{}

//单向认证
type ServerCertConfig struct {
	CertConfig
	ServerCertFile string
}

//双向认证
type ClientCertConfig struct {
	CertConfig
	RootCAFile     string
	ClientCertFile string
	ClientKeyFile  string
}

func CreateServerTlsConfig(serverCertFile string) *tls.Config {
	certpool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(serverCertFile)
	if err == nil {
		certpool.AppendCertsFromPEM(pem)
	}

	return &tls.Config{
		RootCAs: certpool,
	}
}

func CreateClientTlsConfig(rootCAFile string, clientCertFile string, clientKeyFile string) *tls.Config {
	certpool := x509.NewCertPool()
	rootCA, err := ioutil.ReadFile(rootCAFile)
	if err == nil {
		certpool.AppendCertsFromPEM(rootCA)
	}

	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		panic(err)
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}
}

type ExecuteOptions struct {
	Broker            string     // Broker URI
	Qos               byte       // QoS(0|1|2)
	Retain            bool       // Retain
	Topic             string     // Topic
	Username          string     //
	Password          string     //
	CertConfig        CertConfig // 证书相关
	ClientNum         int        //
	Count             int        //
	MessageSize       int        //
	UseDefaultHandler bool       //
	PreTime           int        //
	IntervalTime      int        //
	ClientIdsFileName string
}

func Execute(exec func(clients []*mqtt.Client, opts ExecuteOptions, clientId *ClientId, param ...string) int, opts ExecuteOptions) {
	startTime := time.Now()
	infoLog.Printf("%s start benchmark\n", time.Now())
	message := CreateFixedSizeMessage(opts.MessageSize)
	DefaultHandlerResults = make([]*SubscribeResult, opts.ClientNum)
	//mqtt client lists
	clients := make([]*mqtt.Client, opts.ClientNum)
	//hasError := false
	clientId := new(ClientId)
	clientId.initClientId(opts.ClientIdsFileName)
	infoLog.Println("clientId size :" + strconv.Itoa(len(clientId.ClientIds)))
	//list := clientId.ClientIds;
	for i := 0; i < opts.ClientNum; i++ {
		client := Connect(i, opts, *clientId)
		if i%1000 == 0 {
			infoLog.Println("connected num :" + strconv.Itoa(i))
		}
		if client == nil {
			infoLog.Println(clientId.ClientIds[i])
		}
		clients[i] = &client
	}
	//if hasError {
	//	for i := 0; i < len(clients); i++ {
	//		client := clients[i]
	//		if client != nil {
	//			//(*client)
	//		}
	//	}
	//	return
	//}
	endTimeConn := time.Now()
	durationConn := (endTimeConn.Sub(startTime)).Seconds()
	infoLog.Printf(" connect duration %fs \n", durationConn)
	throughputConn := float64(len(clients)) / float64(durationConn) * 1000
	infoLog.Printf("\nResult : broker=%s, totalClients=%d, totalCount=%d, duration=%fs, throughputConn=%.2fconnMessage/sec\n",
		opts.Broker, opts.ClientNum, len(clients), durationConn, throughputConn)
	totalCount := exec(clients, opts, clientId, message)
	endTime := time.Now()

	infoLog.Printf("%s end benchmark\n", time.Now())
	duration := (endTime.Sub(startTime)).Seconds()
	// messages/sec
	throughput := float64(totalCount) / float64(duration) * 1000
	infoLog.Printf("\nResult : broker=%s, totalClients=%d, totalCount=%d, duration=%fs, throughput=%.2fsubMessage/sec\n",
		opts.Broker, opts.ClientNum, totalCount, duration, throughput)

	time.Sleep(100000 * time.Second)
}
func Connect(id int, execOpts ExecuteOptions, clientIds ClientId) mqtt.Client {
	//通过读取数据库中的clients为其分配一个单独的clientId
	clientId := clientIds.getClientId(id)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(execOpts.Broker)
	opts.SetClientID(clientId)

	certConfig := execOpts.CertConfig
	switch c := certConfig.(type) {
	case ServerCertConfig:
		if tlsConfig == nil {
			tlsConfig = CreateServerTlsConfig(c.ServerCertFile)
		}
		opts.SetTLSConfig(tlsConfig)
	case ClientCertConfig:
		if tlsConfig == nil {
			tlsConfig = CreateClientTlsConfig(c.RootCAFile, c.ClientCertFile, c.ClientKeyFile)
		}
		opts.SetTLSConfig(tlsConfig)
	default:
		if Debug {
			infoLog.Println("no tls config")
		}
	}
	if execOpts.UseDefaultHandler == true {
		var result *SubscribeResult = &SubscribeResult{}
		result.Count = 0
		var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
			result.Count++
			if Debug {
				infoLog.Printf("Received at defaultHandler : topic=%s ,message=%s\n", msg.Topic(), msg.Payload())
			}
		}
		opts.SetDefaultPublishHandler(handler)
		DefaultHandlerResults[id] = result
	}
	client := mqtt.NewClient(opts)
	//go client.Connect()
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		infoLog.Printf("Connected error : %s\n", token.Error())
		return nil
	}
	return client
}

func Disconnect(client mqtt.Client) {
	client.Disconnect(10)
}

func AsyncDisconnect(clients []*mqtt.Client) {
	wg := new(sync.WaitGroup)
	for _, client := range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Disconnect(*client)
		}()
	}
	wg.Wait()
}
func CreateFixedSizeMessage(size int) string {
	var buffer bytes.Buffer
	for i := 0; i < size; i++ {
		buffer.WriteString(strconv.Itoa(i % 10))
	}

	message := buffer.String()
	return message
}

func PublishAllClient(clients []*mqtt.Client, opts ExecuteOptions, clientId *ClientId, param ...string) int {
	message := param[0]
	wg := new(sync.WaitGroup)
	totalCount := 0
	for id := 0; id < len(clients); id++ {
		wg.Add(1)
		client := *clients[id]
		go func(clientId string) {
			defer wg.Done()

			for index := 0; index < opts.Count; index++ {
				topic := fmt.Sprintf(opts.Topic+"/%s"+"/outbox", clientId)
				if Debug {
					infoLog.Printf("Publish : id=%d, count=%d, topic=%s\n", clientId, index, topic)
				}
				Publish(client, topic, opts.Qos, opts.Retain, message)
				totalCount++
				//每个client每隔一段时间上报数据
				//if opts.IntervalTime > 0 {
				//	time.Sleep(time.Duration(opts.IntervalTime) * time.Millisecond)
				//}
			}
		}(clientId.ClientIds[id])
	}
	wg.Wait()
	return totalCount
}
func Publish(client mqtt.Client, topic string, qos byte, retain bool, message string) {
	if client != nil {
		token := client.Publish(topic, qos, retain, message)
		if token.Wait() && token.Error() != nil {
			infoLog.Printf("publish error: %s\n", token.Error())
		}
	} else {
		infoLog.Printf("client is  is nil: %s\n", &client)
	}

}

func SubscribeAllClient(clients []*mqtt.Client, opts ExecuteOptions, clientId *ClientId, param ...string) int {
	wg := new(sync.WaitGroup)
	results := make([]*SubscribeResult, len(clients))
	for id := 0; id < len(clients); id++ {
		wg.Add(1)
		client := *clients[id]
		topic := fmt.Sprintf("/clients"+opts.Topic+"/%s"+"/inbox", clientId.ClientIds[id])
		go func(id int) {
			defer wg.Done()
			results[id] = Subscribe(client, topic, opts.Qos)
		}(id)
	}
	wg.Wait()

	totalCount := 0
	for id := 0; id < len(results); id++ {
		totalCount += results[id].Count
	}
	return totalCount
}
func Subscribe(client mqtt.Client, topic string, qos byte) *SubscribeResult {
	var result *SubscribeResult = &SubscribeResult{}
	result.Count = 0
	var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		infoLog.Printf("Received message : topic=%s, message=%s\n", msg.Topic(), msg.Payload())
	}
	if client != nil {
		token := client.Subscribe(topic, qos, handler)
		if token.Wait() && token.Error() != nil {
			infoLog.Printf("Subscribe error: %s\n", token.Error())
		} else {
			result.Count++
		}
	} else {
		infoLog.Printf("client is  is nil: %s\n", &client)
	}
	return result
}
func main() {
	broker := flag.String("broker", "tcp://{host}:{port}", "URI of MQTT broker (required)")
	action := flag.String("action", "p|pub or s|sub", "Publish or Subscribe or Subscribe(with publishing) (required)")
	qos := flag.Int("qos", 0, "MQTT QoS(0|1|2)")
	retain := flag.Bool("retain", false, "MQTT Retain")
	//topic := flag.String("topic", BASE_TOPIC, "Base topic")
	//username := flag.String("broker-username", "", "Username for connecting to the MQTT broker")
	//password := flag.String("broker-password", "", "Password for connecting to the MQTT broker")
	tls := flag.String("tls", "", "TLS mode. 'server:certFile' or 'client:rootCAFile,clientCertFile,clientKeyFile'")
	clients := flag.Int("clients", 10, "Number of clients")
	count := flag.Int("count", 100, "Number of loops per client")
	size := flag.Int("size", 1024, "Message size per publish (byte)")
	useDefaultHandler := flag.Bool("support-unknown-received", false, "Using default messageHandler for a message that does not match any known subscriptions")
	preTime := flag.Int("pretime", 0, "Pre wait time (ms)")
	intervalTime := flag.Int("intervaltime", 0, "Interval time per message (ms)")
	debug := flag.Bool("x", false, "Debug mode")
	fileName := flag.String("cId", "", "clientId file")
	flag.Parse()
	strArray := strings.Split(*fileName, "_")
	logFileName := "/data/app/log/mqtt_bench.log" + strArray[2]
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("open file error !")
	}
	infoLog = log.New(logFile, "info", log.Llongfile)
	infoLog.Println(fileName)

	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	//validate broker url
	if broker == nil || *broker == "" || *broker == "tcp://{host}:{port}" {
		infoLog.Printf("Invalid argument : -broker -> %s\n", *broker)
		return
	}

	var method string = ""

	if *action == "p" || *action == "sub" {
		method = "pub"
	} else if *action == "s" || *action == "sub" {
		method = "sub"
	}

	if method != "pub" && method != "sub" {
		infoLog.Printf("Invalid argument : -action -> %s\n", *action)
		return
	}

	var certConfig CertConfig = nil

	//parse TLS mode
	if *tls == "" {
		infoLog.Println("mqtt client in pure tcp mode")
	} else if strings.HasPrefix(*tls, "server:") {
		var strArray = strings.Split(*tls, "server:")
		serverCertFile := strings.TrimSpace(strArray[1])
		if FileExists(serverCertFile) == false {
			infoLog.Printf("File is not found. : certFile -> %s\n", serverCertFile)
			return
		}

		certConfig = ServerCertConfig{
			ServerCertFile: serverCertFile}
	} else if strings.HasPrefix(*tls, "client:") {
		infoLog.Println("client tls")
		var strArray = strings.Split(*tls, "client:")
		var configArray = strings.Split(strArray[1], ",")
		rootCAFile := strings.TrimSpace(configArray[0])
		clientCertFile := strings.TrimSpace(configArray[1])
		clientKeyFile := strings.TrimSpace(configArray[2])

		if FileExists(rootCAFile) == false {
			infoLog.Printf("File is not found. : rootCAFile -> %s\n", rootCAFile)
			return
		}
		if FileExists(clientCertFile) == false {
			infoLog.Printf("File is not found. : rootCAFile -> %s\n", clientCertFile)
			return
		}
		if FileExists(clientKeyFile) == false {
			infoLog.Printf("File is not found. : rootCAFile -> %s\n", clientKeyFile)
			return
		}
		certConfig = ClientCertConfig{
			RootCAFile:     rootCAFile,
			ClientCertFile: clientCertFile,
			ClientKeyFile:  clientKeyFile,
		}
	} else {
		infoLog.Println("please pass right cert file")
	}

	execOpts := ExecuteOptions{}
	execOpts.Broker = *broker
	execOpts.Qos = byte(*qos)
	execOpts.Retain = *retain
	//execOpts.Topic =
	execOpts.ClientNum = *clients
	execOpts.Count = *count
	execOpts.MessageSize = *size
	execOpts.UseDefaultHandler = *useDefaultHandler
	execOpts.PreTime = *preTime
	execOpts.IntervalTime = *intervalTime
	execOpts.CertConfig = certConfig
	execOpts.ClientIdsFileName = *fileName
	Debug = *debug

	switch method {
	case "pub":
		infoLog.Println("client pub mode")
		Execute(PublishAllClient, execOpts)
	case "sub":
		infoLog.Println("client sub mode")
		Execute(SubscribeAllClient, execOpts)
	default:
		infoLog.Println("just test connect num")
		//Execute(execOpts)
	}
	select {}
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
