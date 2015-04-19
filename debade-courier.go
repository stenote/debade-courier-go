package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
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
func (this *Debade) send(d []byte, rk string) {

	msg := AMQP.Publishing{
		DeliveryMode: AMQP.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         []byte(d),
	}

	if len(rk) == 0 {
		rk = this.t
	}

	err := this.c.Publish(this.e, rk, false, false, msg)

	if err != nil {
		log.Fatalf("publish: %v", err)
	}
}

func main() {

	var (
		// flag
		conf_file  string
		verbose    bool
		concurrent int

		// debade
		rk string
	)

	// flag 配置设定
	flag.StringVar(&conf_file, "f", "/etc/debade/courier.yml", "Debade config file")
	flag.BoolVar(&verbose, "v", true, "Verbose mode")
	flag.IntVar(&concurrent, "n", 10, "Number of multiple requests to make at a time")

	flag.Parse()

	// 文件目录判断
	if _, err := os.Stat(conf_file); err != nil {
		log.Fatalf("Cannot Load %s: %v", conf_file, err)
	}

	yaml, err := YAML.Open(conf_file)

	if err != nil {
		log.Fatalf("Cannot Open %s: %v", conf_file, err)
	}

	// 设置并发数
	runtime.GOMAXPROCS(concurrent)

	//定义 Debade 合集
	Ds := make(map[string]*Debade)

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

		// durable false
		// autoDelete true
		// internal false
		// 与 debade-trigger 符合
		err = channel.ExchangeDeclare(exchange, _type, false, true, false, false, nil)

		if err != nil {
			log.Fatalf("exchange.declare: %v", err)
		}

		n := name.(string)

		Ds[n] = &Debade{
			c: channel,
			e: exchange,
			t: _type,
		}
	}

	// 进行启动 ZMQ 进行 ZMQ 处理
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

	// 死循环遍历
	for {
		// 等待接受数据
		if str, _ := socket.Recv(ZMQ.DONTWAIT); len(str) > 0 {

			j, _ := JSON.LoadString(str)

			if verbose {
				log.Println(str)
			}

			// 判断是否包含 queue
			if queue, ok := j.CheckGet("queue"); ok {
				d, _ := j.Get("data").MarshalJSON()

				if verbose {
					log.Printf("data: %s\n", d)
				}

				q, _ := queue.String()

				// 包含 routing
				if jrk, ok := j.CheckGet("routing"); ok {
					rk, _ = jrk.String()
				} else {
					rk = ""
				}

				go Ds[q].send(d, rk)
			}
		}
	}
}
