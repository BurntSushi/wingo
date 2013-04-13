package event

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/BurntSushi/xgbutil"

	"github.com/BurntSushi/wingo/logger"
)

var subs subscriptions

func Notifier(X *xgbutil.XUtil, fp string) {
	fp = fp + "-notify"
	os.Remove(fp)

	listener, err := net.Listen("unix", fp)
	if err != nil {
		logger.Error.Fatalln("Could not start IPC event listener: %s", err)
	}
	defer listener.Close()

	subs = manageSubscriptions()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warning.Printf("Error accepting IPC event conn: %s", err)
			continue
		}
		go handleSubscriber(conn)
	}
}

func Notify(ev Event) {
	if subs.notify == nil {
		return
	}
	subs.notify <- ev
}

func handleSubscriber(conn net.Conn) {
	defer conn.Close()

	id, events := subs.subscribe()
	defer subs.unsubscribe(id)

	logger.Message.Printf("Accepted new event subscriber (id: %d).", id)

	encoder := json.NewEncoder(conn)
	for {
		select {
		case <-time.After(5 * time.Second):
			if err := encoder.Encode(eventToMap(Noop{})); err != nil {
				logger.Message.Printf("Subscriber timed out: %s", err)
				return
			}
			if _, err := fmt.Fprintf(conn, "%c", 0); err != nil {
				logger.Message.Printf("Subscriber timed out: %s", err)
				return
			}
		case ev := <-events:
			if err := encoder.Encode(eventToMap(ev)); err != nil {
				logger.Message.Printf("Error sending event: %s", err)
				return
			}
			if _, err := fmt.Fprintf(conn, "%c", 0); err != nil {
				logger.Message.Printf("Error sending event: %s", err)
				return
			}
		}
	}
}

// eventToMap converts an event struct into a map.
// This is a terrible hack in order to inject the event name automatically.
func eventToMap(ev Event) map[string]interface{} {
	rv := reflect.ValueOf(ev)
	rt := rv.Type()
	m := make(map[string]interface{})

	m["EventName"] = rt.Name()
	nf := rv.NumField()
	for i := 0; i < nf; i++ {
		m[rt.Field(i).Name] = rv.Field(i).Interface()
	}
	return m
}

type subscriptions struct {
	add chan chan subscriber // sends info back on the given channel
	remove chan int
	notify chan Event
}

type subscriber struct {
	id int
	events chan Event
}

func (ss subscriptions) subscribe() (int, chan Event) {
	recv := make(chan subscriber)
	ss.add <- recv
	scriber := <-recv
	return scriber.id, scriber.events
}

func (ss subscriptions) unsubscribe(id int) {
	ss.remove <- id
}

func manageSubscriptions() subscriptions {
	nextId := int(1)
	subscribed := make(map[int]chan Event)
	script := subscriptions{
		make(chan chan subscriber),
		make(chan int),
		make(chan Event),
	}

	go func() {
		for {
			select {
			case recv := <-script.add:
				subscribed[nextId] = make(chan Event, 100)
				recv <- subscriber{nextId, subscribed[nextId]}
				nextId++
			case id := <-script.remove:
				close(subscribed[id])
				delete(subscribed, id)
				logger.Message.Printf("Subscriber disconnected (id: %d).", id)
			case ev := <-script.notify:
				for _, subscriber := range subscribed {
					// Do a non-blocking send so that we drop notifications
					// when the client gets too busy (or fails).
					select {
					case subscriber <- ev:
					default:
					}
				}
			}
		}
	}()
	return script
}
