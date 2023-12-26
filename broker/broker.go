package broker

import (
	"context"
	"github.com/difof/collection"
)

const DefaultChannel = "::"

// Broker is a message broadcaster to multiple subscribers (channels).
type Broker[ChannelT comparable, MsgT any] struct {
	pub            chan collection.Tuple[ChannelT, MsgT]
	sub            chan *Subscription[ChannelT, MsgT]
	unsub          chan *Subscription[ChannelT, MsgT]
	defaultChannel ChannelT
}

// New creates and starts a new Broker.
func New[ChannelT comparable, MsgT any](ctx context.Context, defaultChannel ChannelT) (b *Broker[ChannelT, MsgT]) {
	b = &Broker[ChannelT, MsgT]{
		pub:            make(chan collection.Tuple[ChannelT, MsgT], 1),
		sub:            make(chan *Subscription[ChannelT, MsgT], 1),
		unsub:          make(chan *Subscription[ChannelT, MsgT], 1),
		defaultChannel: defaultChannel,
	}

	go b.start(ctx)

	return
}

func NewFrom[MsgT any](ctx context.Context, _ MsgT) *Broker[string, MsgT] {
	return New[string, MsgT](ctx, DefaultChannel)
}

// start starts the broker. Must be called before adding any new subscribers.
// Will block until the broker is stopped.
func (b *Broker[ChannelT, MsgT]) start(ctx context.Context) {
	subs := map[ChannelT]map[*Subscription[ChannelT, MsgT]]struct{}{}

	for {
		select {
		case <-ctx.Done():
			for _, subs := range subs {
				for sub := range subs {
					close(sub.msgCh)
				}
			}
			return
		case sub := <-b.sub:
			if _, ok := subs[sub.channel]; !ok {
				subs[sub.channel] = map[*Subscription[ChannelT, MsgT]]struct{}{}
			}
			subs[sub.channel][sub] = struct{}{}
		case unsub := <-b.unsub:
			delete(subs[unsub.channel], unsub)
		case msg := <-b.pub:
			for sub := range subs[msg.Key()] {
				select {
				case sub.msgCh <- msg.Value():
				default:
				}
			}
		}
	}
}

// Publish publishes a message to the broker on default channel.
func (b *Broker[ChannelT, MsgT]) Publish(msg MsgT) {
	b.pub <- collection.NewTuple(b.defaultChannel, msg)
}

// PublishChannel publishes a message to the broker.
func (b *Broker[ChannelT, MsgT]) PublishChannel(channel ChannelT, msg MsgT) {
	b.pub <- collection.NewTuple(channel, msg)
}

// Subscribe subscribes to the broker on default channel.
func (b *Broker[ChannelT, MsgT]) Subscribe() *Subscription[ChannelT, MsgT] {
	return b.SubscribeChannel(b.defaultChannel)
}

// SubscribeChannel subscribes to the broker.
func (b *Broker[ChannelT, MsgT]) SubscribeChannel(channel ChannelT) *Subscription[ChannelT, MsgT] {
	sub := NewSubscription(b, channel, make(chan MsgT, 5))
	b.sub <- sub
	return sub
}

// Unsubscribe unsubscribes from the broker.
func (b *Broker[ChannelT, MsgT]) Unsubscribe(sub *Subscription[ChannelT, MsgT]) {
	b.unsub <- sub
	close(sub.msgCh)
}
