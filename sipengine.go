package sipengine

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
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
					go func(m *Message){
						for _, v := range e.steps {
							err := v(m)
							if err != nil {

								e.channels.Error <- err

								//Is it a termination error?
								if _, ok := errors.Cause(err).(*MessageTerminationError); ok {
									//Let's stop all processing for the routine for this message
									return
								}
							}
						}
					}(m)

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