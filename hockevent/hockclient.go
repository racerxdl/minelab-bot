package hockevent

import (
	zmq "github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type HockClient interface {
	Send(toSend <-chan HockEvent)
	Recv() <-chan HockEvent
	Stop()
}

func Connect(target string) (HockClient, error) {
	log.Infof("ZMQ Connecting to %q", target)
	var err error
	c := &hockClient{}
	c.ctx, err = zmq.NewContext()
	if err != nil {
		return nil, err
	}
	c.socket, err = c.ctx.NewSocket(zmq.PAIR)
	if err != nil {
		return nil, err
	}

	err = c.socket.Connect(target)
	if err != nil {
		return nil, err
	}
	c.recv = make(chan HockEvent, 100)

	// Sender routine
	go func() {
		var err error
		for err != zmq.ErrorSocketClosed {
			if c.toSend != nil {
				for p := range c.toSend {
					pdata := Serialize(p)
					_, err = c.socket.Send(pdata, 0)
				}
			}
		}
	}()

	// Receiver routine
	go func() {
		var err error
		var msg string
		var event HockEvent
		defer close(c.recv)
		for err != zmq.ErrorSocketClosed {
			msg, err = c.socket.Recv(0)
			if err != nil {
				if err != zmq.ErrorMoreExpected {
					log.Errorf("error receiving message: %s", err)
				}
				continue
			}
			event, err = Deserialize([]byte(msg))
			if err == nil {
				c.recv <- event
			} else {
				log.Errorf("error parsing msg %q: %s", msg, err)
			}
		}
	}()

	return c, nil
}

type hockClient struct {
	ctx    *zmq.Context
	socket *zmq.Socket
	toSend <-chan HockEvent
	recv   chan HockEvent
}

func (h *hockClient) Send(toSend <-chan HockEvent) {
	h.toSend = toSend
}

func (h *hockClient) Recv() <-chan HockEvent {
	return h.recv
}

func (h *hockClient) Stop() {
	_ = h.socket.Close()
	_ = h.ctx.Term()
}
