package event

import (
	"context"
	"sync"

	"github.com/difof/errors"
)

type EventId int

// Bus is a simple event bus that allows for subscribing to events and
// publishing them. Bus is thread-safe.
type Bus struct {
	subscribers map[string]map[EventId]EventHandler
	counter     EventId
	lock        sync.Mutex
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string]map[EventId]EventHandler),
	}
}

// Subscribe adds a new event handler to the bus for a given topic.
// removes subscription when context is done before receiving event
func (b *Bus) Subscribe(ctx context.Context, topic string, handler EventHandler) (id EventId) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.subscribers[topic] == nil {
		b.subscribers[topic] = make(map[EventId]EventHandler)
	}

	id = b.counter
	b.subscribers[topic][b.counter] = handler
	b.counter++

	go func(topic string, id EventId, handler EventHandler) {
		<-ctx.Done()
		b.Unsubscribe(topic, id, handler)
	}(topic, id, handler)

	return
}

// Unsubscribe removes an event handler from the bus for a given topic.
func (b *Bus) Unsubscribe(topic string, id EventId, handler EventHandler) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.subscribers[topic] == nil {
		return
	}

	delete(b.subscribers[topic], id)
}

// Publish publishes an event to all subscribers of a given topic.
func (b *Bus) Publish(topic string, event Event) (err error) {
	if event == nil {
		return
	}

	b.lock.Lock()
	defer b.lock.Unlock()
	defer errors.Recover(&err)

	for i, handler := range b.subscribers[topic] {
		errors.Mustf(handler(event))("handler %d failed", i)
	}

	return
}
