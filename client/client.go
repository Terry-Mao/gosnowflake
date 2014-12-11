package client

import (
	"github.com/samuel/go-zookeeper/zk"
)

type Client struct {
	zkServers []string
}

func NewClient(zkServers []string) *Client {
	return &Client{zkServers: zkServers}
}

func (client *Client) Dial() {
}
