//代码就和狗屎一样!
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	_ "time"

	YAML "menteslibres.net/gosexy/yaml"

	ZMQ "github.com/pebbe/zmq4"
	_ "github.com/streadway/amqp"
)

func main() {

	// flag 配置设定
	conf_file := flag.String("c", "/etc/debade/courier.yml", "debade config file")
	verbose := flag.Bool("v", true, "show more log")
	help := flag.Bool("h", false, "show this message")

	flag.Parse()

	if *help {
		fmt.Fprintln(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}

	if _, err := os.Stat(*conf_file); err != nil {
		log.Fatal("hehe")
	}

	yaml, _:= YAML.Open(*conf_file)

	fmt.Println(yaml.Get("servers"))

	i := yaml.Get("servers", "my_server1")

	fmt.Println(i)

	/*
		for s, info := range i {
			fmt.Println(s)
			fmt.Println(info)
		}
	*/

	//fmt.Println(yaml.Read("servers"))

	fmt.Println(*verbose)

	context, _ := ZMQ.NewContext()

	socket, _ := context.NewSocket(ZMQ.PULL)
	defer socket.Close()

	socket.Bind("tcp://127.0.0.1:5555")
	log.Printf("Listend %s")

	for {

		str, _ := socket.Recv(ZMQ.DONTWAIT)

		if len(str) > 0 {
		}

	}

	/*
		amqp_url := "amqp://guest:guest@172.17.0.2:5672/"

		connection, err := amqp.Dial(amqp_url)

		if err != nil {
			log.Fatal("Conection %v", err)
		}

		defer connection.Close()

		channel, _ := connection.Channel()

		defer channel.Close()

		err = channel.ExchangeDeclare("logfooos", "topic", true, false, false, false, nil)
		if err != nil {
			log.Fatalf("exchange.declare: %v", err)
		}

		msg := amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "text/plain",
			Body:         []byte("Go Go AMQP!"),
		}

		err = channel.Publish("logs", "topic", false, false, msg)

		if err != nil {
			log.Fatalf("basic.publish: %v", err)
		}

	*/

}
