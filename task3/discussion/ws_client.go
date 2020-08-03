package discussion

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
	maxMessageSize = 64*1024
)

// DataType is determines methods of data types sending over websocket.
type DataType interface {
	IsMyInstance([]byte, interface{}) (interface{}, bool)
	WriteProcess(interface{}) *DataPacket
	ReadProcess(interface{}) error
}

// DataPacket is format of data which sent over websocket.
type DataPacket struct {
	Type string      `json:"type,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// Client is some single client connected over websocket.
type Client struct {
	ConnGroup *ConnGroup
	WsConn    *websocket.Conn
	Send       chan interface{}
	connClosed bool
	DataTypes  []DataType
}

// WriteMsg writes message to client and have writeWait limited duration
// to write message
func (c *Client) WriteMsg(messageType int, msg interface{}) error {
	if !c.connClosed {
		err := c.WsConn.SetWriteDeadline(time.Now().Add(writeWait))
		if err != nil {
			return err
		}
		if msg != nil {
			err := c.WsConn.WriteJSON(msg)
			if err != nil {
				return err
			}
		} else {
			err := c.WsConn.WriteMessage(messageType, []byte{})
			if err != nil {
				return err
			}
		}
		return nil
	}
	return ErrConnClosed
}

// ReadMsg reads message from client and have pondWait limited duration
// to read message
func (c *Client) ReadMsg() ([]byte, error) {
	err := c.WsConn.SetReadDeadline(time.Now().Add(pongWait))
	if c.connClosed {
		return nil, ErrConnClosed
	}
	if err != nil {
		return nil, err
	}
	_, msg, err := c.WsConn.ReadMessage()
	if c.connClosed {
		return nil, ErrConnClosed
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// ReadPump is loop which reads messages from client until connection closed
func (c *Client) ReadPump(wg *sync.WaitGroup) error {
	defer func() {
		c.ConnGroup.unregister <- c
		wg.Done()
	}()
	c.WsConn.SetReadLimit(maxMessageSize)
	c.WsConn.SetCloseHandler(
		func(code int, text string) error {
			log.Warn("Connection closed by client")
			c.connClosed = true
			err := c.WsConn.Close()
			if err != nil {
				return err
			}
			return nil
		})
	c.WsConn.SetPongHandler(
		func(string) error {
			return c.WsConn.SetReadDeadline(time.Now().Add(pongWait))
		})
	for {
		if data, err := c.ReadMsg(); err != nil {
			return isErrConnClosed("Can't process reading: ", err)
		} else {
			found := false
			for _, datatype := range c.DataTypes {
				if instance, ok := datatype.IsMyInstance(data, nil); ok {
					err := datatype.ReadProcess(instance)
					if err != nil {
						return isErrConnClosed("Can't process reading: ", err)
					}
					found = true
					break
				}
			}
			if !found {
				err := c.WriteMsg(websocket.TextMessage, DataPacket{
					Type: "warning",
					Data: fmt.Sprintf("Unsupported or invalid data: %v", string(data)),
				})
				if err != nil {
					return isErrConnClosed("", err)
				}
			}
		}
	}
}

// WritePump is loop which writes messages to client until connection closed
func (c *Client) WritePump(wg *sync.WaitGroup) error {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		wg.Done()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel.
				if err := c.WriteMsg(websocket.CloseMessage, []byte{}); err != nil {
					return isErrConnClosed("Can't write close message", err)
				}
				return nil
			}
			found := false
			for _, datatype := range c.DataTypes {
				if instance, ok := datatype.IsMyInstance(nil, message); ok {
					dataPacket := datatype.WriteProcess(instance)
					err := c.WriteMsg(websocket.TextMessage, dataPacket)
					if err != nil {
						return isErrConnClosed("Can't process writing: ", err)
					}
					found = true
					break
				}
			}
			if !found {
				log.Warnf("Sending data type not determined: %v", message)
			}
		case <-ticker.C:
			if err := c.WriteMsg(websocket.PingMessage, nil); err != nil {
				return isErrConnClosed("Can't write ping message", err)
			}
		}
	}
}

func isErrConnClosed(msg string, err error) error {
	if err != ErrConnClosed {
		return errors.New(msg + err.Error())
	}
	return nil
}
