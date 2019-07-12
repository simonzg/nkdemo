package simulator

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"nakama/server"
)

type NKClient struct {
	logger *zap.Logger
	Host   string
	Port   int

	conn *websocket.Conn

	UserID string
}

func NewNKClient(logger *zap.Logger, host string, port int) *NKClient {
	l := logger.With(zap.String("module", "client"))

	return &NKClient{
		logger: l,
		Host:   host,
		Port:   port,
	}
}

func (c *NKClient) Connect(id string) error {
	wsUrl := fmt.Sprintf("ws://%s:%d/ws?id=%s", c.Host, c.Port, id)
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *NKClient) Send(e *server.Envelope) error {
	data, err := proto.Marshal(e)
	err = c.conn.WriteMessage(websocket.BinaryMessage, data)
	return err
}

func (c *NKClient) Stop() {
	c.conn.Close()
}
