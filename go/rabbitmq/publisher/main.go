package main

import (
	"log"

	"github.com/streadway/amqp"
)

type Exchange struct {
	Name string
	Type string
}

var Exchanges = []Exchange{
	{Name: "game.system", Type: "fanout"},
	{Name: "game.boss", Type: "fanout"},
	{Name: "game.event", Type: "fanout"},
	{Name: "game.chat", Type: "topic"},
}

func main() {
	conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
	ch, _ := conn.Channel()

	DeclareExchanges(ch)

	SendSystemBroadcast(ch, "Server maintenance in 10 minutes")
	SendBossEvent(ch, "World Boss has been spawned")
	SendBossEvent(ch, "World Boss has been defeated")
	SendEventStatus(ch, "New event started")
	SendEventStatus(ch, "Current event ended")
	SendGlobalChat(ch, "player123", "Hello world!")
}

func DeclareExchanges(ch *amqp.Channel) error {
	for _, ex := range Exchanges {
		err := ch.ExchangeDeclare(
			ex.Name,
			ex.Type,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func SendSystemBroadcast(ch *amqp.Channel, msg string) {
	publish(ch, "game.system", "", msg)
}

func SendBossEvent(ch *amqp.Channel, msg string) {
	publish(ch, "game.boss", "", msg)
}

func SendEventStatus(ch *amqp.Channel, msg string) {
	publish(ch, "game.event", "", msg)
}

func SendGlobalChat(ch *amqp.Channel, player string, message string) {
	body := player + ": " + message
	publish(ch, "game.chat", "chat.global", body)
}

// Generic publish
func publish(ch *amqp.Channel, exchange, routingKey, msg string) {
	ch.Publish(exchange, routingKey, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	log.Printf("[PUBLISH] ex=%s key=%s msg=%s\n", exchange, routingKey, msg)
}
