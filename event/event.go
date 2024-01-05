package event

type Event interface {
	Topic() string
}

type EventHandler func(Event) error
