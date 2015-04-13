package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	JSON "github.com/gobs/simplejson"
	ZMQ "github.com/pebbe/zmq4"
	AMQP "github.com/streadway/amqp"
	YAML "menteslibres.net/gosexy/yaml"
)

//构造 Debade 结构
type Debade struct {
	c *AMQP.Channel //channel对象

	e string //exchange
	t string //type
}

//消息发送
func (this *Debade) send(d []byte) {

	msg := AMQP.Publishing{
		DeliveryMode: AMQP.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         []byte(d),
	}

	err := this.c.Publish(this.e, this.t, false, false, msg)

	if err != nil {
		log.Fatalf("publish: %v", err)
	}
}

func main() {

	//flag 变量定义
	var (
		conf_file string
		verbose   bool
	)

	// flag 配置设定
	flag.StringVar(&conf_file, "c", "/etc/debade/courier.yml", "debade config file")
	flag.BoolVar(&verbose, "v", true, "Verbose mode")

	flag.Parse()

	// 文件目录判断
	if _, err := os.Stat(conf_file); err != nil {
		log.Fatalf("Cannot Load %s: %v", conf_file, err)
	}

	yaml, err := YAML.Open(conf_file)

	if err != nil {
		log.Fatalf("Cannot Open %s: %v", conf_file, err)
	}

	//定义 Debade 合计
	Ds := make(map[interface{}]*Debade)

	//对 Yaml 中的 Servers 尝试进行 Connect
	for name, _ := range yaml.Get("servers").(map[interface{}]interface{}) {

		host := yaml.Get("servers", name, "host").(string)
		port := yaml.Get("servers", name, "port").(int)
		username := yaml.Get("servers", name, "username").(string)
		password := yaml.Get("servers", name, "password").(string)

		exchange := yaml.Get("servers", name, "exchange").(string)

		if len(exchange) == 0 {
			exchange = "default"
		}

		_type := yaml.Get("servers", name, "type").(string)

		if len(_type) == 0 {
			_type = "fanout"
		}

		amqp_url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)

		if verbose {
			log.Printf("AMQP: %s\n", amqp_url)
			log.Printf("exchange: %s\n", exchange)
		}

		connection, err := AMQP.Dial(amqp_url)
		defer connection.Close()

		if err != nil {
			log.Fatalf("RabbitMQ Connection Error %v", err)
		}

		channel, err := connection.Channel()
		defer channel.Close()

		if err != nil {
			log.Fatalf("RabbitMQ Connection Cannot Get Channel %v", err)
		}

		err = channel.ExchangeDeclare(exchange, _type, true, false, false, false, nil)

		if err != nil {
			log.Fatalf("exchange.declare: %v", err)
		}

		Ds[name] = &Debade{
			c: channel,
			e: exchange,
			t: _type,
		}
	}

	//进行启动 ZMQ 进行 ZMQ 处理
	context, err := ZMQ.NewContext()

	if err != nil {
		log.Fatalf("Cannot Create ZMQ Context: %v", err)
	}

	socket, err := context.NewSocket(ZMQ.PULL)
	defer socket.Close()

	if err != nil {
		log.Fatalf("Cannot Create ZMQ Socket: %v", err)
	}

	socket.Bind("tcp://0.0.0.0:3333")

	log.Println("Listend tcp://0.0.0.0:3333")

	//死循环遍历
	for {
		//等待接受数据
		str, _ := socket.Recv(ZMQ.DONTWAIT)

		if len(str) > 0 {

			j, _ := JSON.LoadString(str)

			if verbose {
				log.Println(str)
			}

			//判断是否包含 queue
			if queue, ok := j.CheckGet("queue"); ok {

				d, _ := j.Get("data").MarshalJSON()

				if verbose {
					log.Printf("data: %s\n", d)
				}

				//fork
				go Ds[queue].send(d)

			}
		}
	}
}
