package main

import (
	"log"

	"github.com/streadway/amqp"
)

var Exchanges = []struct {
	Name string
	Type string
}{
	{"game.system", "fanout"},
	{"game.boss", "fanout"},
	{"game.event", "fanout"},
	{"game.chat", "topic"},
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	DeclareExchanges(ch)

	// Create unique queue per WebSocket server
	q, _ := ch.QueueDeclare("", false, true, true, false, nil)

	// Bind
	binds := []struct {
		Exchange   string
		RoutingKey string
	}{
		{"game.system", ""},
		{"game.boss", ""},
		{"game.event", ""},
		{"game.chat", "chat.#"},
	}

	for _, b := range binds {
		if err := ch.QueueBind(q.Name, b.RoutingKey, b.Exchange, false, nil); err != nil {
			log.Fatal(err)
		}
	}

	msgs, _ := ch.Consume(q.Name, "", true, true, false, false, nil)

	log.Println("[Consumer] listening for game events...")

	for msg := range msgs {
		log.Printf("[EVENT] ex=%v body=%v\n", msg.Exchange, string(msg.Body))
	}
}

func DeclareExchanges(ch *amqp.Channel) error {
	for _, ex := range Exchanges {
		if err := ch.ExchangeDeclare(
			ex.Name,
			ex.Type,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return err
		}
	}
	return nil
}
