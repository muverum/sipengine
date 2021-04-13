package sipengine

import (
	"bytes"
	"context"
	"net"

	"github.com/pkg/errors"
)

type Engine struct {
	//Context will be used for managing shutdown operations.
	address  string
	ctx      context.Context
	steps    []SIPStep
	channels ChannelMap
}

type ChannelMap struct {
	Ingress chan *Message
	Egress  chan *Message
	Error   chan error
}

//NewEngine will return the engined used for processing. We are allowing
//users to provide their own channels as input to let them decide
//how (or if) they should be buffered. This decouples this layer from
//configuration or assumptions about the user's preferences for traffic.
func NewEngine(address string, ctx context.Context, channelmap ChannelMap, steps ...SIPStep) Engine {
	return Engine{
		address:  address,
		ctx:      ctx,
		steps:    steps,
		channels: channelmap,
	}
}

//Start listening and funneling off requests
func (e Engine) ListenAndServe() error {

	pc, err := net.ListenPacket("udp", e.address)

	if err != nil {
		return errors.Wrap(err, "failure during startup")
	}

	//Processing for messages on ingress channel
	go func() {
		for {
			select {
			case m := <-e.channels.Ingress:
				//Spin off a separate routine for each message to process in order to
				//allow for termination of execution as users signal for it.
				mctx, mcancel := context.WithCancel(e.ctx)
				defer mcancel()
				go func() {
					for _, v := range e.steps {
						select {
						//Don't work on this message if the context is already cancelled
						//Or completed
						case <-mctx.Done():
							return
						default:
							//
						}
						err := v(m)
						if err != nil {
							//Is it a termination error?
							if _, ok := errors.Cause(err).(*MessageTerminationError); ok {
								//Let's stop all processing for the routine for this message
								return
							} else {
								e.channels.Error <- errors.Wrap(err, "error during pipeline step processing")
							}
						}
					}
				}()

			case <-e.ctx.Done():
				//The context has marked itself as complete so let's stop processing new messages
				return
			default:
				//
			}
		}
	}()

	//Begin pulling messages off of the listener.
	go func() {
		for {
			//We should think about having the buffer size value provided somehow
			buf := make([]byte, 2048)
			//_ , addr -> Used to set from address
			_, _, err := pc.ReadFrom(buf)
			if err != nil {
				//This is a scenario where we would want to stop as we can't listen from the network
				//for some reason
				e.channels.Error <- errors.Wrap(err, "unable to read from listening socket")
				return
			}

			message, err := NewMessage(bytes.NewReader(buf), e.ctx)

			//message.Detail.From = addr.String()

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
	case <-e.ctx.Done():
		pc.Close()
		//Wait for existing messages to filter out
		//TODO: How to track active dialogs and messages without state
		err = NewShutDownSignalError("Context has been cancelled. Shutdown initiating")
	}

	return err
}
