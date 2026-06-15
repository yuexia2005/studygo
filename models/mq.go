package models

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	MQConn    *amqp.Connection
	MQChannel *amqp.Channel
)

func InitMQ() {
	url := os.Getenv("RABBITMQ_URL")
	//设置默认连接地址
	if url == "" {
		url = "amqp://admin:password@rabbitmq:5672/"
	}

	var err error
	//拨号连接 RabbitMQ
	MQConn, err = amqp.Dial(url)
	if err != nil {
		log.Fatalf("无法连接 RabbitMQ: %v", err)
	}

	//打开信道
	MQChannel, err = MQConn.Channel()
	if err != nil {
		log.Fatalf("无法打开 MQ 通道: %v", err)
	}

	_, err = MQChannel.QueueDeclare(
		"video_tasks", //队列名
		true,          //持久化
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("无法声明队列: %v", err)
	}
}

func CloseMQ() {
	if MQChannel != nil {
		MQChannel.Close()
	}
	if MQConn != nil {
		MQConn.Close()
	}
}
