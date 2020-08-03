package discussion

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 64 * 1024
)

type DataPacket struct {
	Type string      `json:"type,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type ClientConnection interface {
	Start() error
	Stop()
	readPump() error
	writePump() error
	DataChan() chan interface{}
}

type client struct {
	ws      *websocket.Conn
	dataCh  chan interface{}
	stopCh  chan bool
	closed  bool
	stopped bool
	events  []Event
}

func NewClientConnection(ws *websocket.Conn, events ...Event) ClientConnection {
	return &client{
		ws:     ws,
		dataCh: make(chan interface{}, 256),
		stopCh: make(chan bool),
		events: events,
	}
}
func (c *client) Start() error {
	var errs []string
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if err := c.writePump(); err != nil && err != ErrConnClosed {
			errs = append(errs, err.Error())
		}
		wg.Done()
	}()
	go func() {
		if err := c.readPump(); err != nil && err != ErrConnClosed {
			errs = append(errs, err.Error())
		}
		wg.Done()
	}()
	wg.Wait()
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}

func (c *client) Stop() {
	if !c.stopped {
		c.stopped = true
		close(c.stopCh)
	}
}

func (c *client) DataChan() chan interface{} {
	return c.dataCh
}

func (c *client) readPump() error {
	defer c.Stop()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetCloseHandler(
		func(code int, text string) error {
			log.Warn("Connection closed by client")
			return nil
		})
	c.ws.SetPongHandler(
		func(string) error {
			return c.ws.SetReadDeadline(time.Now().Add(pongWait))
		})
	for {
		select {
		case <-c.stopCh:
			return nil
		default:
		}
		if data, err := c.readMsg(); err != nil {
			return err
		} else {
			found := false
			for _, event := range c.events {
				if instance, ok := event.IsMyInstance(data, nil); ok {
					found = true
					err := event.ReadProcess(instance)
					if err != nil {
						return err
					}
					break
				}
			}
			if !found {
				err := c.writeMsg(websocket.TextMessage, DataPacket{
					Type: "warning",
					Data: fmt.Sprintf("Unsupported or invalid event: %v", string(data)),
				})
				if err != nil {
					return err
				}
			}
		}
	}
}

func (c *client) writePump() error {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.Stop()
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-c.dataCh:
			if !ok {
				// The hub closed the channel.
				if err := c.writeMsg(websocket.CloseMessage, nil); err != nil {
					return err
				}
				return nil
			}
			found := false
			for _, datatype := range c.events {
				if instance, ok := datatype.IsMyInstance(nil, message); ok {
					dataPacket := datatype.WriteProcess(instance)
					err := c.writeMsg(websocket.TextMessage, dataPacket)
					if err != nil {
						return err
					}
					found = true
					break
				}
			}
			if !found {
				log.Warnf("Sending data type not determined: %v", message)
			}
		case <-ticker.C:
			if err := c.writeMsg(websocket.PingMessage, nil); err != nil {
				return err
			}
		case <-c.stopCh:
			return nil
		}
	}
}

func (c *client) writeMsg(messageType int, msg interface{}) error {
	if !c.closed {
		err := c.ws.SetWriteDeadline(time.Now().Add(writeWait))
		if err != nil {
			return c.isClosedErr(err)
		}
		if msg != nil {
			err := c.ws.WriteJSON(msg)
			if err != nil {
				return c.isClosedErr(err)
			}
		} else {
			err := c.ws.WriteMessage(messageType, []byte{})
			if err != nil {
				return c.isClosedErr(err)
			}
		}
		return nil
	}
	return ErrConnClosed
}

func (c *client) readMsg() ([]byte, error) {
	err := c.ws.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return nil, c.isClosedErr(err)
	}
	_, msg, err := c.ws.ReadMessage()
	if err != nil {
		return nil, c.isClosedErr(err)
	}
	return msg, nil
}

func (c *client) close() error {
	if !c.closed {
		c.closed = true
		err := c.ws.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) isClosedErr(err error) error {
	if c.closed || err == websocket.ErrCloseSent ||
		websocket.IsCloseError(err,
			websocket.CloseAbnormalClosure,
			websocket.CloseNormalClosure,
			websocket.CloseNoStatusReceived,
		) {
		return ErrConnClosed
	}
	return err
}

func isErrConnClosed(msg string, err error) error {
	if err != ErrConnClosed {
		return errors.New(msg + err.Error())
	}
	return nil
}
