package cfgsrv

import (
	"github.com/tonjun/wsclient"
)

type Client struct {
	cli *wsclient.WSClient
}

func NewClient(serverAddress string) *Client {
	return &Client{}
}

func (c *Client) GetConfig() (string, error) {
	return "", nil
}
