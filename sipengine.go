package sipengine

import (
	"bytes"
	"context"
	"net"
)

type Engine struct {
	//Context will be used for managing shutdown operations.
	address string
	ctx context.Context
	steps []SIPStep
	channels ChannelMap
}

type ChannelMap struct {
	Ingress chan *Message
	Egress chan *Message
	Error chan error
}

//NewEngine will return the engined used for processing. We are allowing
//users to provide their own channels as input to let them decide
//how (or if) they should be buffered. This decouples this layer from
//configuration or assumptions about the user's preferences for traffic.
func NewEngine(address string, ctx context.Context, channelmap ChannelMap, steps... SIPStep) Engine {
	return Engine{
		address: address,
		ctx:  ctx,
		steps: steps,
		channels: channelmap,
	}
}

//Start listening and funneling off requests
func (e Engine) ListenAndServe() error {
	pc, err := net.ListenPacket("udp", e.address)

	if err != nil {
		return err
	}

	//Processing for messages on ingress channel
	go func(){
		for {
			select {
				case m := <- e.channels.Ingress:
					for _, v := range e.steps {
						err := v(m)
						if err != nil {
							e.channels.Error <- err
						}
					}
				case <- e.ctx.Done():
					return
				default:
					//
			}
		}
	}()

	//Begin pulling messages off of the listener.
	go func(){
		for {
			//We should think about having the buffer size value provided somehow
			buf := make([]byte, 2048)
			_, addr, err := pc.ReadFrom(buf)
			if err != nil {
				e.channels.Error <- err
				return
			}

			message, err := NewMessage(bytes.NewReader(buf), e.ctx)

			message.Detail.From = addr.String()

			if err != nil {
				e.channels.Error <- err
				return
			}

			e.channels.Ingress <- message
		}
	}()

	// `Close`ing the packet "connection" means cleaning the data structures
	// allocated for holding information about the listening socket.
	defer pc.Close()

	//Also need to look for context cancellations to close the connection.
	select {
		case <- e.ctx.Done():
			pc.Close()
			err = NewShutDownSignalError("Context has been cancelled. Shutdown initiating")
	}

	return err
}