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

func main() {

	var (
		conf_file string
		verbose   bool
		chs       map[string](*AMQP.Channel) //key 为 queue 名称, value 为 channel
		exs       map[string]string          //key 为 queue 名称, value 为 exchange
		types     map[string]string          //key为 queue 名称, value 为 type
		_type     string                     //queue 类型
		exchange  string                     //exchange名称
		channel   *AMQP.Channel              //channel对象
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

	context, _ := ZMQ.NewContext()

	socket, _ := context.NewSocket(ZMQ.PULL)
	defer socket.Close()

	socket.Bind("tcp://0.0.0.0:3333")

	log.Println("Listend tcp://0.0.0.0:3333")

	//死循环遍历
	for {
		//等待接受数据
		str, _ := socket.Recv(ZMQ.DONTWAIT)

		if len(str) > 0 {

			j, err := JSON.LoadString(str)

			if verbose {
				log.Println(str)
			}

			//可正常解析
			if err == nil {

				//判断是否包含 queue
				if queue, ok := j.CheckGet("queue"); ok {

					q, _ := queue.String()

					c := yaml.Get("servers", q)

					//初始化内存
					chs = make(map[string](*AMQP.Channel))
					exs = make(map[string]string)
					types = make(map[string]string)

					if c != nil {

						channel, ok = chs[q]
						exchange, _ = exs[q]
						_type, _ = types[q]

						if !ok {
							host := yaml.Get("servers", q, "host").(string)
							port := yaml.Get("servers", q, "port").(int)
							username := yaml.Get("servers", q, "username").(string)
							password := yaml.Get("servers", q, "password").(string)

							exchange = yaml.Get("servers", q, "exchange").(string)

							if len(exchange) == 0 {
								exchange = "default"
							}

							_type := yaml.Get("servers", q, "type").(string)

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
								log.Fatal("RabbitMQ Connection Error %v", err)
							}

							channel, _ = connection.Channel()
							defer channel.Close()

							err = channel.ExchangeDeclarePassive(exchange, _type, true, false, false, false, nil)

							if err != nil {
								log.Fatalf("exchange.declare: %v", err)
							}

							chs[q] = channel
							exs[q] = exchange
							types[q] = _type

						}

						data := j.Get("data")
						d, _ := data.MarshalJSON()

						msg := AMQP.Publishing{
							DeliveryMode: AMQP.Persistent,
							Timestamp:    time.Now(),
							ContentType:  "application/json",
							Body:         []byte(d),
						}

						if verbose {
							log.Printf("data: %s\n", d)
						}

						err = channel.Publish(exchange, _type, false, false, msg)
						if err != nil {
							log.Fatalf("publish: %v", err)
						}
					}
				}
			}
		}
	}
}
