package server

import (
	"fmt"
	"go-redis/handler"
	"go-redis/logger"
	"go-redis/protocol"
	"io"
	"net"
)

type Client struct {
	id       string
	conn     net.Conn
	parser   *protocol.Parser
	router   *handler.Router
	shutdown chan struct{}
}

func NewClient(conn net.Conn, router *handler.Router, id string) *Client {
	return &Client{
		id:       id,
		conn:     conn,
		parser:   protocol.NewParser(conn),
		router:   router,
		shutdown: make(chan struct{}),
	}
}

func (c *Client) Serve() {
	logger.Infof("[%s] Client connected from %s", c.id, c.conn.RemoteAddr())
	defer logger.Infof("[%s] Client disconnected", c.id)
	defer c.conn.Close()

	for {
		select {
		case <-c.shutdown:
			return
		default:
		}

		cmd, err := c.parser.Parse()
		if err != nil {
			if err == io.EOF {
				return
			}

			logger.Errorf("[%s] Parse error: %v", c.id, err)
			errorResp := protocol.Error(fmt.Sprintf("ERR parse error: %v", err))
			c.sendResponse(errorResp)
			continue
		}

		logger.Debugf("[%s] Received command: %+v", c.id, cmd)

		response := c.router.Route(cmd)

		if err := c.sendResponse(response); err != nil {
			logger.Errorf("[%s] Failed to send response: %v", c.id, err)
			return
		}
	}
}

func (c *Client) sendResponse(resp *protocol.Value) error {
	data := protocol.Serialize(resp)
	logger.Debug(resp)

	_, err := c.conn.Write([]byte(data))
	if err != nil {
		return err
	}

	logger.Debugf("[%s] Sent response: %s", c.id, data)
	return nil
}

func (c *Client) Close() error {
	close(c.shutdown)
	return c.conn.Close()
}
